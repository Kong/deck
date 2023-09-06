package file

import (
	"context"
	"errors"
	"fmt"

	"github.com/blang/semver/v4"
	"github.com/kong/deck/konnect"
	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
)

const ratelimitingAdvancedPluginName = "rate-limiting-advanced"

type stateBuilder struct {
	targetContent   *Content
	rawState        *utils.KongRawState
	konnectRawState *utils.KonnectRawState
	currentState    *state.KongState
	defaulter       *utils.Defaulter
	kongVersion     semver.Version

	selectTags   []string
	skipCACerts  bool
	intermediate *state.KongState

	client *kong.Client
	ctx    context.Context

	schemasCache map[string]map[string]interface{}

	disableDynamicDefaults bool

	isKonnect bool

	checkRoutePaths bool

	isConsumerGroupScopedPluginSupported bool

	err error
}

// uuid generates a UUID string and returns a pointer to it.
// It is a variable for testing purpose, to override and supply
// a deterministic UUID generator.
var uuid = func() *string {
	return kong.String(utils.UUID())
}

var ErrWorkspaceNotFound = fmt.Errorf("workspace not found")

func (b *stateBuilder) build() (*utils.KongRawState, *utils.KonnectRawState, error) {
	// setup
	var err error
	b.rawState = &utils.KongRawState{}
	b.konnectRawState = &utils.KonnectRawState{}
	b.schemasCache = make(map[string]map[string]interface{})

	b.intermediate, err = state.NewKongState()
	if err != nil {
		return nil, nil, err
	}

	defaulter, err := defaulter(b.ctx, b.client, b.targetContent, b.disableDynamicDefaults, b.isKonnect)
	if err != nil {
		return nil, nil, err
	}
	b.defaulter = defaulter

	if utils.Kong300Version.LTE(b.kongVersion) {
		b.checkRoutePaths = true
	}

	if utils.Kong340Version.LTE(b.kongVersion) || b.isKonnect {
		b.isConsumerGroupScopedPluginSupported = true
	}

	// build
	b.certificates()
	if !b.skipCACerts {
		b.caCertificates()
	}
	b.services()
	b.routes()
	b.upstreams()
	b.consumerGroups()
	b.consumers()
	b.plugins()
	b.enterprise()

	// konnect
	b.konnect()

	// result
	if b.err != nil {
		return nil, nil, b.err
	}
	return b.rawState, b.konnectRawState, nil
}

func (b *stateBuilder) ingestConsumerGroupScopedPlugins(cg FConsumerGroupObject) error {
	var plugins []FPlugin
	for _, plugin := range cg.Plugins {
		plugin.ConsumerGroup = utils.GetConsumerGroupReference(cg.ConsumerGroup)
		plugins = append(plugins, FPlugin{
			Plugin: kong.Plugin{
				ID:     plugin.ID,
				Name:   plugin.Name,
				Config: plugin.Config,
				ConsumerGroup: &kong.ConsumerGroup{
					ID: cg.ID,
				},
			},
		})
	}
	return b.ingestPlugins(plugins)
}

func (b *stateBuilder) addConsumerGroupPlugins(
	cg FConsumerGroupObject, cgo *kong.ConsumerGroupObject,
) error {
	for _, plugin := range cg.Plugins {
		if utils.Empty(plugin.ID) {
			current, err := b.currentState.ConsumerGroupPlugins.Get(
				*plugin.Name, *cg.ConsumerGroup.ID,
			)
			if errors.Is(err, state.ErrNotFound) {
				plugin.ID = uuid()
			} else if err != nil {
				return err
			} else {
				plugin.ID = kong.String(*current.ID)
			}
		}
		b.defaulter.MustSet(plugin)
		cgo.Plugins = append(cgo.Plugins, plugin)
	}
	return nil
}

func (b *stateBuilder) consumerGroups() {
	if b.err != nil {
		return
	}

	for _, cg := range b.targetContent.ConsumerGroups {
		cg := cg
		if utils.Empty(cg.ID) {
			current, err := b.currentState.ConsumerGroups.Get(*cg.Name)
			if errors.Is(err, state.ErrNotFound) {
				cg.ID = uuid()
			} else if err != nil {
				b.err = err
				return
			} else {
				cg.ID = kong.String(*current.ID)
			}
		}
		utils.MustMergeTags(&cg.ConsumerGroup, b.selectTags)

		cgo := kong.ConsumerGroupObject{
			ConsumerGroup: &cg.ConsumerGroup,
		}

		err := b.intermediate.ConsumerGroups.Add(state.ConsumerGroup{ConsumerGroup: cg.ConsumerGroup})
		if err != nil {
			b.err = err
			return
		}

		// Plugins and Consumer Groups can be handled in two ways:
		//   1. directly in the ConsumerGroup object
		//   2. by scoping the plugin to the ConsumerGroup (Kong >= 3.4.0)
		//
		// The first method is deprecated and will be removed in the future, but
		// we still need to support it for now. The isConsumerGroupScopedPluginSupported
		// flag is used to determine which method to use based on the Kong version.
		if b.isConsumerGroupScopedPluginSupported {
			if err := b.ingestConsumerGroupScopedPlugins(cg); err != nil {
				b.err = err
				return
			}
		} else {
			if err := b.addConsumerGroupPlugins(cg, &cgo); err != nil {
				b.err = err
				return
			}
		}
		b.rawState.ConsumerGroups = append(b.rawState.ConsumerGroups, &cgo)
	}
}

func (b *stateBuilder) certificates() {
	if b.err != nil {
		return
	}

	for i := range b.targetContent.Certificates {
		c := b.targetContent.Certificates[i]
		if utils.Empty(c.ID) {
			cert, err := b.currentState.Certificates.GetByCertKey(*c.Cert,
				*c.Key)
			if errors.Is(err, state.ErrNotFound) {
				c.ID = uuid()
			} else if err != nil {
				b.err = err
				return
			} else {
				c.ID = kong.String(*cert.ID)
			}
		}
		utils.MustMergeTags(&c, b.selectTags)

		snisFromCert := c.SNIs

		kongCert := kong.Certificate{
			ID:        c.ID,
			Key:       c.Key,
			Cert:      c.Cert,
			Tags:      c.Tags,
			CreatedAt: c.CreatedAt,
		}
		b.rawState.Certificates = append(b.rawState.Certificates, &kongCert)

		// snis associated with the certificate
		var snis []kong.SNI
		for _, sni := range snisFromCert {
			sni.Certificate = &kong.Certificate{ID: kong.String(*c.ID)}
			snis = append(snis, sni)
		}
		if err := b.ingestSNIs(snis); err != nil {
			b.err = err
			return
		}
	}
}

func (b *stateBuilder) ingestSNIs(snis []kong.SNI) error {
	for _, sni := range snis {
		sni := sni
		if utils.Empty(sni.ID) {
			currentSNI, err := b.currentState.SNIs.Get(*sni.Name)
			if errors.Is(err, state.ErrNotFound) {
				sni.ID = uuid()
			} else if err != nil {
				return err
			} else {
				sni.ID = kong.String(*currentSNI.ID)
			}
		}
		utils.MustMergeTags(&sni, b.selectTags)
		b.rawState.SNIs = append(b.rawState.SNIs, &sni)
	}
	return nil
}

func (b *stateBuilder) caCertificates() {
	if b.err != nil {
		return
	}

	for _, c := range b.targetContent.CACertificates {
		c := c
		if utils.Empty(c.ID) {
			cert, err := b.currentState.CACertificates.Get(*c.Cert)
			if errors.Is(err, state.ErrNotFound) {
				c.ID = uuid()
			} else if err != nil {
				b.err = err
				return
			} else {
				c.ID = kong.String(*cert.ID)
			}
		}
		utils.MustMergeTags(&c.CACertificate, b.selectTags)

		b.rawState.CACertificates = append(b.rawState.CACertificates,
			&c.CACertificate)
	}
}

func (b *stateBuilder) consumers() {
	if b.err != nil {
		return
	}

	for _, c := range b.targetContent.Consumers {
		c := c
		if utils.Empty(c.ID) {
			var (
				consumer *state.Consumer
				err      error
			)
			if c.Username != nil {
				consumer, err = b.currentState.Consumers.GetByIDOrUsername(*c.Username)
			}
			if errors.Is(err, state.ErrNotFound) || consumer == nil {
				if c.CustomID != nil {
					consumer, err = b.currentState.Consumers.GetByCustomID(*c.CustomID)
					if err == nil {
						c.ID = kong.String(*consumer.ID)
					}
				}
				if c.ID == nil {
					c.ID = uuid()
				}
			} else if err != nil {
				b.err = err
				return
			} else {
				c.ID = kong.String(*consumer.ID)
			}
		}
		utils.MustMergeTags(&c.Consumer, b.selectTags)

		b.rawState.Consumers = append(b.rawState.Consumers, &c.Consumer)
		err := b.intermediate.Consumers.Add(state.Consumer{Consumer: c.Consumer})
		if err != nil {
			b.err = err
			return
		}

		// ingest consumer into consumer group
		if err := b.ingestIntoConsumerGroup(c); err != nil {
			b.err = err
			return
		}

		// plugins for the Consumer
		var plugins []FPlugin
		for _, p := range c.Plugins {
			p.Consumer = utils.GetConsumerReference(c.Consumer)
			plugins = append(plugins, *p)
		}
		if err := b.ingestPlugins(plugins); err != nil {
			b.err = err
			return
		}

		var keyAuths []kong.KeyAuth
		for _, cred := range c.KeyAuths {
			cred.Consumer = utils.GetConsumerReference(c.Consumer)
			keyAuths = append(keyAuths, *cred)
		}
		if err := b.ingestKeyAuths(keyAuths); err != nil {
			b.err = err
			return
		}

		var basicAuths []kong.BasicAuth
		for _, cred := range c.BasicAuths {
			cred.Consumer = utils.GetConsumerReference(c.Consumer)
			basicAuths = append(basicAuths, *cred)
		}
		if err := b.ingestBasicAuths(basicAuths); err != nil {
			b.err = err
			return
		}

		var hmacAuths []kong.HMACAuth
		for _, cred := range c.HMACAuths {
			cred.Consumer = utils.GetConsumerReference(c.Consumer)
			hmacAuths = append(hmacAuths, *cred)
		}
		if err := b.ingestHMACAuths(hmacAuths); err != nil {
			b.err = err
			return
		}

		var jwtAuths []kong.JWTAuth
		for _, cred := range c.JWTAuths {
			cred.Consumer = utils.GetConsumerReference(c.Consumer)
			jwtAuths = append(jwtAuths, *cred)
		}
		if err := b.ingestJWTAuths(jwtAuths); err != nil {
			b.err = err
			return
		}

		var oauth2Creds []kong.Oauth2Credential
		for _, cred := range c.Oauth2Creds {
			cred.Consumer = utils.GetConsumerReference(c.Consumer)
			oauth2Creds = append(oauth2Creds, *cred)
		}
		if err := b.ingestOauth2Creds(oauth2Creds); err != nil {
			b.err = err
			return
		}

		var aclGroups []kong.ACLGroup
		for _, cred := range c.ACLGroups {
			cred.Consumer = utils.GetConsumerReference(c.Consumer)
			aclGroups = append(aclGroups, *cred)
		}
		if err := b.ingestACLGroups(aclGroups); err != nil {
			b.err = err
			return
		}

		var mtlsAuths []kong.MTLSAuth
		for _, cred := range c.MTLSAuths {
			cred.Consumer = utils.GetConsumerReference(c.Consumer)
			mtlsAuths = append(mtlsAuths, *cred)
		}

		b.ingestMTLSAuths(mtlsAuths)
	}
}

func (b *stateBuilder) ingestIntoConsumerGroup(consumer FConsumer) error {
	for _, group := range consumer.Groups {
		found := false
		for _, cg := range b.rawState.ConsumerGroups {
			if group.ID != nil && *cg.ConsumerGroup.ID == *group.ID {
				cg.Consumers = append(cg.Consumers, &consumer.Consumer)
				found = true
				break

			}
			if group.Name != nil && *cg.ConsumerGroup.Name == *group.Name {
				cg.Consumers = append(cg.Consumers, &consumer.Consumer)
				found = true
				break
			}
		}
		if !found {
			var groupIdentifier string
			if group.Name != nil {
				groupIdentifier = *group.Name
			} else {
				groupIdentifier = *group.ID
			}
			return fmt.Errorf(
				"consumer-group '%s' not found for consumer '%s'", groupIdentifier, *consumer.ID,
			)
		}
	}
	return nil
}

func (b *stateBuilder) ingestKeyAuths(creds []kong.KeyAuth) error {
	for _, cred := range creds {
		cred := cred
		if utils.Empty(cred.ID) {
			existingCred, err := b.currentState.KeyAuths.Get(*cred.Key)
			if errors.Is(err, state.ErrNotFound) {
				cred.ID = uuid()
			} else if err != nil {
				return err
			} else {
				cred.ID = kong.String(*existingCred.ID)
			}
		}
		if b.kongVersion.GTE(utils.Kong140Version) {
			utils.MustMergeTags(&cred, b.selectTags)
		}
		b.rawState.KeyAuths = append(b.rawState.KeyAuths, &cred)
	}
	return nil
}

func (b *stateBuilder) ingestBasicAuths(creds []kong.BasicAuth) error {
	for _, cred := range creds {
		cred := cred
		if utils.Empty(cred.ID) {
			existingCred, err := b.currentState.BasicAuths.Get(*cred.Username)
			if errors.Is(err, state.ErrNotFound) {
				cred.ID = uuid()
			} else if err != nil {
				return err
			} else {
				cred.ID = kong.String(*existingCred.ID)
			}
		}
		if b.kongVersion.GTE(utils.Kong140Version) {
			utils.MustMergeTags(&cred, b.selectTags)
		}
		b.rawState.BasicAuths = append(b.rawState.BasicAuths, &cred)
	}
	return nil
}

func (b *stateBuilder) ingestHMACAuths(creds []kong.HMACAuth) error {
	for _, cred := range creds {
		cred := cred
		if utils.Empty(cred.ID) {
			existingCred, err := b.currentState.HMACAuths.Get(*cred.Username)
			if errors.Is(err, state.ErrNotFound) {
				cred.ID = uuid()
			} else if err != nil {
				return err
			} else {
				cred.ID = kong.String(*existingCred.ID)
			}
		}
		if b.kongVersion.GTE(utils.Kong140Version) {
			utils.MustMergeTags(&cred, b.selectTags)
		}
		b.rawState.HMACAuths = append(b.rawState.HMACAuths, &cred)
	}
	return nil
}

func (b *stateBuilder) ingestJWTAuths(creds []kong.JWTAuth) error {
	for _, cred := range creds {
		cred := cred
		if utils.Empty(cred.ID) {
			existingCred, err := b.currentState.JWTAuths.Get(*cred.Key)
			if errors.Is(err, state.ErrNotFound) {
				cred.ID = uuid()
			} else if err != nil {
				return err
			} else {
				cred.ID = kong.String(*existingCred.ID)
			}
		}
		if b.kongVersion.GTE(utils.Kong140Version) {
			utils.MustMergeTags(&cred, b.selectTags)
		}
		b.rawState.JWTAuths = append(b.rawState.JWTAuths, &cred)
	}
	return nil
}

func (b *stateBuilder) ingestOauth2Creds(creds []kong.Oauth2Credential) error {
	for _, cred := range creds {
		cred := cred
		if utils.Empty(cred.ID) {
			existingCred, err := b.currentState.Oauth2Creds.Get(*cred.ClientID)
			if errors.Is(err, state.ErrNotFound) {
				cred.ID = uuid()
			} else if err != nil {
				return err
			} else {
				cred.ID = kong.String(*existingCred.ID)
			}
		}
		if b.kongVersion.GTE(utils.Kong140Version) {
			utils.MustMergeTags(&cred, b.selectTags)
		}
		b.rawState.Oauth2Creds = append(b.rawState.Oauth2Creds, &cred)
	}
	return nil
}

func (b *stateBuilder) ingestACLGroups(creds []kong.ACLGroup) error {
	for _, cred := range creds {
		cred := cred
		if utils.Empty(cred.ID) {
			existingCred, err := b.currentState.ACLGroups.Get(
				*cred.Consumer.ID,
				*cred.Group)
			if errors.Is(err, state.ErrNotFound) {
				cred.ID = uuid()
			} else if err != nil {
				return err
			} else {
				cred.ID = kong.String(*existingCred.ID)
			}
		}
		if b.kongVersion.GTE(utils.Kong140Version) {
			utils.MustMergeTags(&cred, b.selectTags)
		}
		b.rawState.ACLGroups = append(b.rawState.ACLGroups, &cred)
	}
	return nil
}

func (b *stateBuilder) ingestMTLSAuths(creds []kong.MTLSAuth) {
	kong230Version := semver.MustParse("2.3.0")
	for _, cred := range creds {
		cred := cred
		// normally, we'd want to look up existing resources in this case
		// however, this is impossible here: mtls-auth simply has no unique fields other than ID,
		// so we don't--schema validation requires the ID
		// there's nothing more to do here

		if b.kongVersion.GTE(kong230Version) {
			utils.MustMergeTags(&cred, b.selectTags)
		}
		b.rawState.MTLSAuths = append(b.rawState.MTLSAuths, &cred)
	}
}

func (b *stateBuilder) konnect() {
	if b.err != nil {
		return
	}

	for i := range b.targetContent.ServicePackages {
		targetSP := b.targetContent.ServicePackages[i]
		if utils.Empty(targetSP.ID) {
			currentSP, err := b.currentState.ServicePackages.Get(*targetSP.Name)
			if errors.Is(err, state.ErrNotFound) {
				targetSP.ID = uuid()
			} else if err != nil {
				b.err = err
				return
			} else {
				targetSP.ID = kong.String(*currentSP.ID)
			}
		}

		targetKonnectSP := konnect.ServicePackage{
			ID:          targetSP.ID,
			Name:        targetSP.Name,
			Description: targetSP.Description,
		}

		if targetSP.Document != nil {
			targetKonnectDoc := konnect.Document{
				ID:        targetSP.Document.ID,
				Path:      targetSP.Document.Path,
				Published: targetSP.Document.Published,
				Content:   targetSP.Document.Content,
				Parent:    &targetKonnectSP,
			}
			if utils.Empty(targetKonnectDoc.ID) {
				currentDoc, err := b.currentState.Documents.GetByParent(&targetKonnectSP, *targetKonnectDoc.Path)
				if errors.Is(err, state.ErrNotFound) {
					targetKonnectDoc.ID = uuid()
				} else if err != nil {
					b.err = err
					return
				} else {
					targetKonnectDoc.ID = kong.String(*currentDoc.ID)
				}
			}
			b.konnectRawState.Documents = append(b.konnectRawState.Documents, &targetKonnectDoc)
		}

		// versions associated with the package
		for _, targetSV := range targetSP.Versions {
			targetKonnectSV := konnect.ServiceVersion{
				ID:      targetSV.ID,
				Version: targetSV.Version,
			}
			targetRelationID := ""
			if utils.Empty(targetKonnectSV.ID) {
				currentSV, err := b.currentState.ServiceVersions.Get(*targetKonnectSP.ID, *targetKonnectSV.Version)
				if errors.Is(err, state.ErrNotFound) {
					targetKonnectSV.ID = uuid()
				} else if err != nil {
					b.err = err
					return
				} else {
					targetKonnectSV.ID = kong.String(*currentSV.ID)
					if currentSV.ControlPlaneServiceRelation != nil {
						targetRelationID = *currentSV.ControlPlaneServiceRelation.ID
					}
				}
			}
			if targetSV.Implementation != nil &&
				targetSV.Implementation.Kong != nil {
				err := b.ingestService(targetSV.Implementation.Kong.Service)
				if err != nil {
					b.err = err
					return
				}
				targetKonnectSV.ControlPlaneServiceRelation = &konnect.ControlPlaneServiceRelation{
					ControlPlaneEntityID: targetSV.Implementation.Kong.Service.ID,
				}
				if targetRelationID != "" {
					targetKonnectSV.ControlPlaneServiceRelation.ID = &targetRelationID
				}
			}
			if targetSV.Document != nil {
				targetKonnectDoc := konnect.Document{
					ID:        targetSV.Document.ID,
					Path:      targetSV.Document.Path,
					Published: targetSV.Document.Published,
					Content:   targetSV.Document.Content,
					Parent:    &targetKonnectSV,
				}
				if utils.Empty(targetKonnectDoc.ID) {
					currentDoc, err := b.currentState.Documents.GetByParent(&targetKonnectSV, *targetKonnectDoc.Path)
					if errors.Is(err, state.ErrNotFound) {
						targetKonnectDoc.ID = uuid()
					} else if err != nil {
						b.err = err
						return
					} else {
						targetKonnectDoc.ID = kong.String(*currentDoc.ID)
					}
				}
				b.konnectRawState.Documents = append(b.konnectRawState.Documents, &targetKonnectDoc)
			}
			targetKonnectSP.Versions = append(targetKonnectSP.Versions, targetKonnectSV)
		}

		b.konnectRawState.ServicePackages = append(b.konnectRawState.ServicePackages,
			&targetKonnectSP)
	}
}

func (b *stateBuilder) services() {
	if b.err != nil {
		return
	}

	for _, s := range b.targetContent.Services {
		s := s
		err := b.ingestService(&s)
		if err != nil {
			b.err = err
			return
		}
	}
}

func (b *stateBuilder) ingestService(s *FService) error {
	if utils.Empty(s.ID) {
		svc, err := b.currentState.Services.Get(*s.Name)
		if errors.Is(err, state.ErrNotFound) {
			s.ID = uuid()
		} else if err != nil {
			return err
		} else {
			s.ID = kong.String(*svc.ID)
		}
	}
	utils.MustMergeTags(&s.Service, b.selectTags)
	b.defaulter.MustSet(&s.Service)

	b.rawState.Services = append(b.rawState.Services, &s.Service)
	err := b.intermediate.Services.Add(state.Service{Service: s.Service})
	if err != nil {
		return err
	}

	// plugins for the service
	var plugins []FPlugin
	for _, p := range s.Plugins {
		p.Service = utils.GetServiceReference(s.Service)
		plugins = append(plugins, *p)
	}
	if err := b.ingestPlugins(plugins); err != nil {
		return err
	}

	// routes for the service
	for _, r := range s.Routes {
		r := r
		r.Service = utils.GetServiceReference(s.Service)
		if err := b.ingestRoute(*r); err != nil {
			return err
		}
	}
	return nil
}

func (b *stateBuilder) routes() {
	if b.err != nil {
		return
	}

	for _, r := range b.targetContent.Routes {
		r := r
		if err := b.ingestRoute(r); err != nil {
			b.err = err
			return
		}
	}

	// check routes' paths format
	if b.checkRoutePaths {
		unsupportedRoutes := []string{}
		allRoutes, err := b.intermediate.Routes.GetAll()
		if err != nil {
			b.err = err
			return
		}
		for _, r := range allRoutes {
			if utils.HasPathsWithRegex300AndAbove(r.Route) {
				unsupportedRoutes = append(unsupportedRoutes, *r.Route.ID)
			}
		}
		if len(unsupportedRoutes) > 0 {
			utils.PrintRouteRegexWarning(unsupportedRoutes)
		}
	}
}

func (b *stateBuilder) enterprise() {
	b.rbacRoles()
	b.vaults()
}

func (b *stateBuilder) vaults() {
	if b.err != nil {
		return
	}

	for _, v := range b.targetContent.Vaults {
		v := v
		if utils.Empty(v.ID) {
			vault, err := b.currentState.Vaults.Get(*v.Prefix)
			if errors.Is(err, state.ErrNotFound) {
				v.ID = uuid()
			} else if err != nil {
				b.err = err
				return
			} else {
				v.ID = kong.String(*vault.ID)
			}
		}
		utils.MustMergeTags(&v.Vault, b.selectTags)

		b.rawState.Vaults = append(b.rawState.Vaults, &v.Vault)
	}
}

func (b *stateBuilder) rbacRoles() {
	if b.err != nil {
		return
	}

	for _, r := range b.targetContent.RBACRoles {
		r := r
		if utils.Empty(r.ID) {
			role, err := b.currentState.RBACRoles.Get(*r.Name)
			if errors.Is(err, state.ErrNotFound) {
				r.ID = uuid()
			} else if err != nil {
				b.err = err
				return
			} else {
				r.ID = kong.String(*role.ID)
			}
		}
		b.rawState.RBACRoles = append(b.rawState.RBACRoles, &r.RBACRole)
		// rbac endpoint permissions for the role
		for _, ep := range r.EndpointPermissions {
			ep := ep
			ep.Role = &kong.RBACRole{ID: kong.String(*r.ID)}
			b.rawState.RBACEndpointPermissions = append(b.rawState.RBACEndpointPermissions, &ep.RBACEndpointPermission)
		}
	}
}

func (b *stateBuilder) upstreams() {
	if b.err != nil {
		return
	}

	for _, u := range b.targetContent.Upstreams {
		u := u
		if utils.Empty(u.ID) {
			ups, err := b.currentState.Upstreams.Get(*u.Name)
			if errors.Is(err, state.ErrNotFound) {
				u.ID = uuid()
			} else if err != nil {
				b.err = err
				return
			} else {
				u.ID = kong.String(*ups.ID)
			}
		}
		utils.MustMergeTags(&u.Upstream, b.selectTags)
		b.defaulter.MustSet(&u.Upstream)

		b.rawState.Upstreams = append(b.rawState.Upstreams, &u.Upstream)

		// targets for the upstream
		var targets []kong.Target
		for _, t := range u.Targets {
			t.Upstream = &kong.Upstream{ID: kong.String(*u.ID)}
			targets = append(targets, t.Target)
		}
		if err := b.ingestTargets(targets); err != nil {
			b.err = err
			return
		}
	}
}

func (b *stateBuilder) ingestTargets(targets []kong.Target) error {
	for _, t := range targets {
		t := t
		if utils.Empty(t.ID) {
			target, err := b.currentState.Targets.Get(*t.Upstream.ID, *t.Target)
			if errors.Is(err, state.ErrNotFound) {
				t.ID = uuid()
			} else if err != nil {
				return err
			} else {
				t.ID = kong.String(*target.ID)
			}
		}
		utils.MustMergeTags(&t, b.selectTags)
		b.defaulter.MustSet(&t)
		b.rawState.Targets = append(b.rawState.Targets, &t)
	}
	return nil
}

func (b *stateBuilder) plugins() {
	if b.err != nil {
		return
	}

	var plugins []FPlugin
	for _, p := range b.targetContent.Plugins {
		p := p
		if p.Consumer != nil && !utils.Empty(p.Consumer.ID) {
			c, err := b.intermediate.Consumers.GetByIDOrUsername(*p.Consumer.ID)
			if errors.Is(err, state.ErrNotFound) {
				b.err = fmt.Errorf("consumer %v for plugin %v: %w",
					p.Consumer.FriendlyName(), *p.Name, err)

				return
			} else if err != nil {
				b.err = err
				return
			}
			p.Consumer = utils.GetConsumerReference(c.Consumer)
		}
		if p.Service != nil && !utils.Empty(p.Service.ID) {
			s, err := b.intermediate.Services.Get(*p.Service.ID)
			if errors.Is(err, state.ErrNotFound) {
				b.err = fmt.Errorf("service %v for plugin %v: %w",
					p.Service.FriendlyName(), *p.Name, err)

				return
			} else if err != nil {
				b.err = err
				return
			}
			p.Service = utils.GetServiceReference(s.Service)
		}
		if p.Route != nil && !utils.Empty(p.Route.ID) {
			r, err := b.intermediate.Routes.Get(*p.Route.ID)
			if errors.Is(err, state.ErrNotFound) {
				b.err = fmt.Errorf("route %v for plugin %v: %w",
					p.Route.FriendlyName(), *p.Name, err)

				return
			} else if err != nil {
				b.err = err
				return
			}
			p.Route = utils.GetRouteReference(r.Route)
		}
		if p.ConsumerGroup != nil && !utils.Empty(p.ConsumerGroup.ID) {
			cg, err := b.intermediate.ConsumerGroups.Get(*p.ConsumerGroup.ID)
			if errors.Is(err, state.ErrNotFound) {
				b.err = fmt.Errorf("consumer-group %v for plugin %v: %w",
					p.ConsumerGroup.FriendlyName(), *p.Name, err)
				return
			} else if err != nil {
				b.err = err
				return
			}
			p.ConsumerGroup = utils.GetConsumerGroupReference(cg.ConsumerGroup)
		}

		if err := b.validatePlugin(p); err != nil {
			b.err = err
			return
		}
		plugins = append(plugins, p)
	}
	if err := b.ingestPlugins(plugins); err != nil {
		b.err = err
		return
	}
}

func (b *stateBuilder) validatePlugin(p FPlugin) error {
	if b.isConsumerGroupScopedPluginSupported && *p.Name == ratelimitingAdvancedPluginName {
		// check if deprecated consumer-groups configuration is present in the config
		var consumerGroupsFound bool
		if groups, ok := p.Config["consumer_groups"]; ok {
			// if groups is an array of length > 0, then consumer_groups is set
			if groupsArray, ok := groups.([]interface{}); ok && len(groupsArray) > 0 {
				consumerGroupsFound = true
			}
		}
		var enforceConsumerGroupsFound bool
		if enforceConsumerGroups, ok := p.Config["enforce_consumer_groups"]; ok {
			if enforceConsumerGroupsBool, ok := enforceConsumerGroups.(bool); ok && enforceConsumerGroupsBool {
				enforceConsumerGroupsFound = true
			}
		}
		if consumerGroupsFound || enforceConsumerGroupsFound {
			return utils.ErrorConsumerGroupUpgrade
		}
	}
	return nil
}

// strip_path schema default value is 'true', but it cannot be set when
// protocols include 'grpc' and/or 'grpcs'. When users explicitly set
// strip_path to 'true' with grpc/s protocols, deck returns a schema violation error.
// When strip_path is not set and protocols include grpc/s, deck sets strip_path to 'false',
// despite its default value would be 'true' under normal circumstances.
func getStripPathBasedOnProtocols(route kong.Route) (*bool, error) {
	for _, p := range route.Protocols {
		if *p == "grpc" || *p == "grpcs" {
			if route.StripPath != nil && *route.StripPath {
				return nil, fmt.Errorf("schema violation (strip_path: cannot set " +
					"'strip_path' when 'protocols' is 'grpc' or 'grpcs')")
			}
			return kong.Bool(false), nil
		}
	}
	return route.StripPath, nil
}

func (b *stateBuilder) ingestRoute(r FRoute) error {
	if utils.Empty(r.ID) {
		route, err := b.currentState.Routes.Get(*r.Name)
		if errors.Is(err, state.ErrNotFound) {
			r.ID = uuid()
		} else if err != nil {
			return err
		} else {
			r.ID = kong.String(*route.ID)
		}
	}

	utils.MustMergeTags(&r, b.selectTags)

	stripPath, err := getStripPathBasedOnProtocols(r.Route)
	if err != nil {
		return err
	}
	r.Route.StripPath = stripPath
	b.defaulter.MustSet(&r.Route)

	b.rawState.Routes = append(b.rawState.Routes, &r.Route)
	err = b.intermediate.Routes.Add(state.Route{Route: r.Route})
	if err != nil {
		return err
	}

	// plugins for the route
	var plugins []FPlugin
	for _, p := range r.Plugins {
		p.Route = utils.GetRouteReference(r.Route)
		plugins = append(plugins, *p)
	}
	if err := b.ingestPlugins(plugins); err != nil {
		return err
	}
	if r.Service != nil && utils.Empty(r.Service.ID) && !utils.Empty(r.Service.Name) {
		s, err := b.intermediate.Services.Get(*r.Service.Name)
		if err != nil {
			return fmt.Errorf("retrieve intermediate services (%s): %w", *r.Service.Name, err)
		}
		r.Service.ID = s.ID
		r.Service.Name = nil
	}
	return nil
}

func (b *stateBuilder) getPluginSchema(pluginName string) (map[string]interface{}, error) {
	var schema map[string]interface{}

	// lookup in cache
	if schema, ok := b.schemasCache[pluginName]; ok {
		return schema, nil
	}

	exists, err := utils.WorkspaceExists(b.ctx, b.client)
	if err != nil {
		return nil, fmt.Errorf("ensure workspace exists: %w", err)
	}
	if !exists {
		return schema, ErrWorkspaceNotFound
	}

	schema, err = b.client.Plugins.GetFullSchema(b.ctx, &pluginName)
	if err != nil {
		return schema, err
	}
	b.schemasCache[pluginName] = schema
	return schema, nil
}

func (b *stateBuilder) addPluginDefaults(plugin *FPlugin) error {
	if b.client == nil {
		return nil
	}
	schema, err := b.getPluginSchema(*plugin.Name)
	if err != nil {
		if errors.Is(err, ErrWorkspaceNotFound) {
			return nil
		}
		return fmt.Errorf("retrieve schema for %v from Kong: %w", *plugin.Name, err)
	}
	return kong.FillPluginsDefaults(&plugin.Plugin, schema)
}

func (b *stateBuilder) ingestPlugins(plugins []FPlugin) error {
	for _, p := range plugins {
		p := p
		if utils.Empty(p.ID) {
			cID, rID, sID, cgID := pluginRelations(&p.Plugin)
			plugin, err := b.currentState.Plugins.GetByProp(*p.Name,
				sID, rID, cID, cgID)
			if errors.Is(err, state.ErrNotFound) {
				p.ID = uuid()
			} else if err != nil {
				return err
			} else {
				p.ID = kong.String(*plugin.ID)
			}
		}
		if p.Config == nil {
			p.Config = make(map[string]interface{})
		}
		p.Config = ensureJSON(p.Config)
		err := b.fillPluginConfig(&p)
		if err != nil {
			return err
		}
		if err := b.addPluginDefaults(&p); err != nil {
			return fmt.Errorf("add defaults to plugin '%v': %w", *p.Name, err)
		}
		utils.MustMergeTags(&p, b.selectTags)
		b.rawState.Plugins = append(b.rawState.Plugins, &p.Plugin)
	}
	return nil
}

func (b *stateBuilder) fillPluginConfig(plugin *FPlugin) error {
	if plugin == nil {
		return fmt.Errorf("plugin is nil")
	}
	if !utils.Empty(plugin.ConfigSource) {
		conf, ok := b.targetContent.PluginConfigs[*plugin.ConfigSource]
		if !ok {
			return fmt.Errorf("_plugin_config %q not found",
				*plugin.ConfigSource)
		}
		for k, v := range conf {
			if _, ok := plugin.Config[k]; !ok {
				plugin.Config[k] = v
			}
		}
	}
	return nil
}

func pluginRelations(plugin *kong.Plugin) (cID, rID, sID, cgID string) {
	if plugin.Consumer != nil && !utils.Empty(plugin.Consumer.ID) {
		cID = *plugin.Consumer.ID
	}
	if plugin.Route != nil && !utils.Empty(plugin.Route.ID) {
		rID = *plugin.Route.ID
	}
	if plugin.Service != nil && !utils.Empty(plugin.Service.ID) {
		sID = *plugin.Service.ID
	}
	if plugin.ConsumerGroup != nil && !utils.Empty(plugin.ConsumerGroup.ID) {
		cgID = *plugin.ConsumerGroup.ID
	}
	return
}

func defaulter(
	ctx context.Context, client *kong.Client, fileContent *Content, disableDynamicDefaults, isKonnect bool,
) (*utils.Defaulter, error) {
	var kongDefaults KongDefaults
	if fileContent.Info != nil {
		kongDefaults = fileContent.Info.Defaults
	}
	opts := utils.DefaulterOpts{
		Client:                 client,
		KongDefaults:           kongDefaults,
		DisableDynamicDefaults: disableDynamicDefaults,
		IsKonnect:              isKonnect,
	}
	defaulter, err := utils.GetDefaulter(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("creating defaulter: %w", err)
	}
	return defaulter, nil
}
