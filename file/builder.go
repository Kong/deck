package file

import (
	"fmt"

	"github.com/blang/semver/v4"
	"github.com/kong/deck/konnect"
	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
)

type stateBuilder struct {
	targetContent   *Content
	rawState        *utils.KongRawState
	konnectRawState *utils.KonnectRawState
	currentState    *state.KongState
	defaulter       *utils.Defaulter
	kongVersion     semver.Version

	selectTags   []string
	intermediate *state.KongState

	err error
}

var kong140Version = semver.MustParse("1.4.0")

// uuid generates a UUID string and returns a pointer to it.
// It is a variable for testing purpose, to override and supply
// a deterministic UUID generator.
var uuid = func() *string {
	return kong.String(utils.UUID())
}

func (b *stateBuilder) build() (*utils.KongRawState, *utils.KonnectRawState, error) {
	// setup
	var err error
	b.rawState = &utils.KongRawState{}
	b.konnectRawState = &utils.KonnectRawState{}

	if b.targetContent.Info != nil {
		b.selectTags = b.targetContent.Info.SelectorTags
	}
	b.intermediate, err = state.NewKongState()
	if err != nil {
		return nil, nil, err
	}

	// defaulter
	var kongDefaults KongDefaults
	if b.targetContent.Info != nil {
		kongDefaults = b.targetContent.Info.Defaults
	}
	b.defaulter, err = defaulter(kongDefaults)
	if err != nil {
		return nil, nil, fmt.Errorf("creating defaulter: %w", err)
	}

	// build
	b.certificates()
	b.caCertificates()
	b.services()
	b.routes()
	b.upstreams()
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

func (b *stateBuilder) certificates() {
	if b.err != nil {
		return
	}

	for i := range b.targetContent.Certificates {
		c := b.targetContent.Certificates[i]
		if utils.Empty(c.ID) {
			cert, err := b.currentState.Certificates.GetByCertKey(*c.Cert,
				*c.Key)
			if err == state.ErrNotFound {
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
			if err == state.ErrNotFound {
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
			if err == state.ErrNotFound {
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
			consumer, err := b.currentState.Consumers.Get(*c.Username)
			if err == state.ErrNotFound {
				c.ID = uuid()
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

		// plugins for the Consumer
		var plugins []FPlugin
		for _, p := range c.Plugins {
			p.Consumer = &kong.Consumer{ID: kong.String(*c.ID)}
			plugins = append(plugins, *p)
		}
		if err := b.ingestPlugins(plugins); err != nil {
			b.err = err
			return
		}

		var keyAuths []kong.KeyAuth
		for _, cred := range c.KeyAuths {
			cred.Consumer = &kong.Consumer{ID: kong.String(*c.ID)}
			keyAuths = append(keyAuths, *cred)
		}
		if err := b.ingestKeyAuths(keyAuths); err != nil {
			b.err = err
			return
		}

		var basicAuths []kong.BasicAuth
		for _, cred := range c.BasicAuths {
			cred.Consumer = &kong.Consumer{ID: kong.String(*c.ID)}
			basicAuths = append(basicAuths, *cred)
		}
		if err := b.ingestBasicAuths(basicAuths); err != nil {
			b.err = err
			return
		}

		var hmacAuths []kong.HMACAuth
		for _, cred := range c.HMACAuths {
			cred.Consumer = &kong.Consumer{ID: kong.String(*c.ID)}
			hmacAuths = append(hmacAuths, *cred)
		}
		if err := b.ingestHMACAuths(hmacAuths); err != nil {
			b.err = err
			return
		}

		var jwtAuths []kong.JWTAuth
		for _, cred := range c.JWTAuths {
			cred.Consumer = &kong.Consumer{ID: kong.String(*c.ID)}
			jwtAuths = append(jwtAuths, *cred)
		}
		if err := b.ingestJWTAuths(jwtAuths); err != nil {
			b.err = err
			return
		}

		var oauth2Creds []kong.Oauth2Credential
		for _, cred := range c.Oauth2Creds {
			cred.Consumer = &kong.Consumer{ID: kong.String(*c.ID)}
			oauth2Creds = append(oauth2Creds, *cred)
		}
		if err := b.ingestOauth2Creds(oauth2Creds); err != nil {
			b.err = err
			return
		}

		var aclGroups []kong.ACLGroup
		for _, cred := range c.ACLGroups {
			cred.Consumer = &kong.Consumer{ID: kong.String(*c.ID)}
			aclGroups = append(aclGroups, *cred)
		}
		if err := b.ingestACLGroups(aclGroups); err != nil {
			b.err = err
			return
		}

		var mtlsAuths []kong.MTLSAuth
		for _, cred := range c.MTLSAuths {
			cred.Consumer = &kong.Consumer{
				ID: kong.String(*c.ID),
			}
			mtlsAuths = append(mtlsAuths, *cred)
		}

		b.ingestMTLSAuths(mtlsAuths)
	}
}

func (b *stateBuilder) ingestKeyAuths(creds []kong.KeyAuth) error {
	for _, cred := range creds {
		cred := cred
		if utils.Empty(cred.ID) {
			existingCred, err := b.currentState.KeyAuths.Get(*cred.Key)
			if err == state.ErrNotFound {
				cred.ID = uuid()
			} else if err != nil {
				return err
			} else {
				cred.ID = kong.String(*existingCred.ID)
			}
		}
		if b.kongVersion.GTE(kong140Version) {
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
			if err == state.ErrNotFound {
				cred.ID = uuid()
			} else if err != nil {
				return err
			} else {
				cred.ID = kong.String(*existingCred.ID)
			}
		}
		if b.kongVersion.GTE(kong140Version) {
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
			if err == state.ErrNotFound {
				cred.ID = uuid()
			} else if err != nil {
				return err
			} else {
				cred.ID = kong.String(*existingCred.ID)
			}
		}
		if b.kongVersion.GTE(kong140Version) {
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
			if err == state.ErrNotFound {
				cred.ID = uuid()
			} else if err != nil {
				return err
			} else {
				cred.ID = kong.String(*existingCred.ID)
			}
		}
		if b.kongVersion.GTE(kong140Version) {
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
			if err == state.ErrNotFound {
				cred.ID = uuid()
			} else if err != nil {
				return err
			} else {
				cred.ID = kong.String(*existingCred.ID)
			}
		}
		if b.kongVersion.GTE(kong140Version) {
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
			if err == state.ErrNotFound {
				cred.ID = uuid()
			} else if err != nil {
				return err
			} else {
				cred.ID = kong.String(*existingCred.ID)
			}
		}
		if b.kongVersion.GTE(kong140Version) {
			utils.MustMergeTags(&cred, b.selectTags)
		}
		b.rawState.ACLGroups = append(b.rawState.ACLGroups, &cred)
	}
	return nil
}

func (b *stateBuilder) ingestMTLSAuths(creds []kong.MTLSAuth) {
	for _, cred := range creds {
		cred := cred
		// normally, we'd want to look up existing resources in this case
		// however, this is impossible here: mtls-auth simply has no unique fields other than ID,
		// so we don't--schema validation requires the ID
		// there's nothing more to do here

		// TODO: this is stub code, since mtls-auth doesn't actually have tag support yet
		// They probably should, FTI-1706 tracks that request with the Kong Enterprise team
		//if b.kongVersion.GTE(kong220Version) {
		//	utils.MustMergeTags(&cred, b.selectTags)
		//}
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
			if err == state.ErrNotFound {
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
				if err == state.ErrNotFound {
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
				if err == state.ErrNotFound {
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
					if err == state.ErrNotFound {
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
		if err == state.ErrNotFound {
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
		p.Service = &kong.Service{ID: kong.String(*s.ID)}
		plugins = append(plugins, *p)
	}
	if err := b.ingestPlugins(plugins); err != nil {
		return err
	}

	// routes for the service
	for _, r := range s.Routes {
		r := r
		r.Service = &kong.Service{ID: kong.String(*s.ID)}
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
}

func (b *stateBuilder) enterprise() {
	b.rbacRoles()
}

func (b *stateBuilder) rbacRoles() {
	if b.err != nil {
		return
	}

	for _, r := range b.targetContent.RBACRoles {
		r := r
		if utils.Empty(r.ID) {
			role, err := b.currentState.RBACRoles.Get(*r.Name)
			if err == state.ErrNotFound {
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
			if err == state.ErrNotFound {
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
			if err == state.ErrNotFound {
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
			c, err := b.intermediate.Consumers.Get(*p.Consumer.ID)
			if err == state.ErrNotFound {
				b.err = fmt.Errorf("consumer %v for plugin %v: %w",
					*p.Consumer.ID, *p.Name, err)

				return
			} else if err != nil {
				b.err = err
				return
			}
			p.Consumer = &kong.Consumer{ID: kong.String(*c.ID)}
		}
		if p.Service != nil && !utils.Empty(p.Service.ID) {
			s, err := b.intermediate.Services.Get(*p.Service.ID)
			if err == state.ErrNotFound {
				b.err = fmt.Errorf("service %v for plugin %v: %w",
					*p.Service.ID, *p.Name, err)

				return
			} else if err != nil {
				b.err = err
				return
			}
			p.Service = &kong.Service{ID: kong.String(*s.ID)}
		}
		if p.Route != nil && !utils.Empty(p.Route.ID) {
			s, err := b.intermediate.Routes.Get(*p.Route.ID)
			if err == state.ErrNotFound {
				b.err = fmt.Errorf("route %v for plugin %v: %w",
					*p.Route.ID, *p.Name, err)

				return
			} else if err != nil {
				b.err = err
				return
			}
			p.Route = &kong.Route{ID: kong.String(*s.ID)}
		}
		plugins = append(plugins, p)
	}
	if err := b.ingestPlugins(plugins); err != nil {
		b.err = err
		return
	}
}

func (b *stateBuilder) ingestRoute(r FRoute) error {
	if utils.Empty(r.ID) {
		route, err := b.currentState.Routes.Get(*r.Name)
		if err == state.ErrNotFound {
			r.ID = uuid()
		} else if err != nil {
			return err
		} else {
			r.ID = kong.String(*route.ID)
		}
	}

	utils.MustMergeTags(&r, b.selectTags)
	b.defaulter.MustSet(&r.Route)

	b.rawState.Routes = append(b.rawState.Routes, &r.Route)
	err := b.intermediate.Routes.Add(state.Route{Route: r.Route})
	if err != nil {
		return err
	}

	// plugins for the route
	var plugins []FPlugin
	for _, p := range r.Plugins {
		p.Route = &kong.Route{ID: kong.String(*r.ID)}
		plugins = append(plugins, *p)
	}
	if err := b.ingestPlugins(plugins); err != nil {
		return err
	}
	return nil
}

func (b *stateBuilder) ingestPlugins(plugins []FPlugin) error {
	for _, p := range plugins {
		p := p
		if utils.Empty(p.ID) {
			cID, rID, sID := pluginRelations(&p.Plugin)
			plugin, err := b.currentState.Plugins.GetByProp(*p.Name,
				sID, rID, cID)
			if err == state.ErrNotFound {
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

func pluginRelations(plugin *kong.Plugin) (cID, rID, sID string) {
	if plugin.Consumer != nil && !utils.Empty(plugin.Consumer.ID) {
		cID = *plugin.Consumer.ID
	}
	if plugin.Route != nil && !utils.Empty(plugin.Route.ID) {
		rID = *plugin.Route.ID
	}
	if plugin.Service != nil && !utils.Empty(plugin.Service.ID) {
		sID = *plugin.Service.ID
	}
	return
}

func defaulter(defaults KongDefaults) (*utils.Defaulter, error) {
	d, err := utils.GetKongDefaulter()
	if err != nil {
		return nil, err
	}
	if defaults.Route != nil {
		if err = d.Register(defaults.Route); err != nil {
			return nil, err
		}
	}
	if defaults.Service != nil {
		if err = d.Register(defaults.Service); err != nil {
			return nil, err
		}
	}
	if defaults.Upstream != nil {
		if err = d.Register(defaults.Upstream); err != nil {
			return nil, err
		}
	}
	if defaults.Target != nil {
		if err = d.Register(defaults.Target); err != nil {
			return nil, err
		}
	}
	return d, nil
}
