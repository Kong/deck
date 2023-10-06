package file

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
	"sigs.k8s.io/yaml"
)

// WriteConfig holds settings to use to write the state file.
type WriteConfig struct {
	Workspace        string
	SelectTags       []string
	Filename         string
	FileFormat       Format
	WithID           bool
	RuntimeGroupName string
	KongVersion      string
}

func compareOrder(obj1, obj2 sortable) bool {
	return strings.Compare(obj1.sortKey(), obj2.sortKey()) < 0
}

func getFormatVersion(kongVersion string) (string, error) {
	parsedKongVersion, err := utils.ParseKongVersion(kongVersion)
	if err != nil {
		return "", fmt.Errorf("parsing Kong version: %w", err)
	}
	formatVersion := "1.1"
	if parsedKongVersion.GTE(utils.Kong300Version) {
		formatVersion = "3.0"
	}
	return formatVersion, nil
}

// KongStateToFile generates a state object to file.Content.
// It will omit timestamps and IDs while writing.
func KongStateToContent(kongState *state.KongState, config WriteConfig) (*Content, error) {
	file := &Content{}
	var err error

	file.Workspace = config.Workspace
	formatVersion, err := getFormatVersion(config.KongVersion)
	if err != nil {
		return nil, fmt.Errorf("get format version: %w", err)
	}
	file.FormatVersion = formatVersion
	if config.RuntimeGroupName != "" {
		file.Konnect = &Konnect{
			RuntimeGroupName: config.RuntimeGroupName,
		}
	}

	selectTags := config.SelectTags
	if len(selectTags) > 0 {
		file.Info = &Info{
			SelectorTags: selectTags,
		}
	}

	err = populateServices(kongState, file, config)
	if err != nil {
		return nil, err
	}

	err = populateServicelessRoutes(kongState, file, config)
	if err != nil {
		return nil, err
	}

	err = populatePlugins(kongState, file, config)
	if err != nil {
		return nil, err
	}

	err = populateUpstreams(kongState, file, config)
	if err != nil {
		return nil, err
	}

	err = populateCertificates(kongState, file, config)
	if err != nil {
		return nil, err
	}

	err = populateCACertificates(kongState, file, config)
	if err != nil {
		return nil, err
	}

	err = populateConsumers(kongState, file, config)
	if err != nil {
		return nil, err
	}

	err = populateVaults(kongState, file, config)
	if err != nil {
		return nil, err
	}

	err = populateConsumerGroups(kongState, file, config)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// KongStateToFile writes a state object to file with filename.
// See KongStateToContent for the State generation
func KongStateToFile(kongState *state.KongState, config WriteConfig) error {
	file, err := KongStateToContent(kongState, config)
	if err != nil {
		return err
	}
	return WriteContentToFile(file, config.Filename, config.FileFormat)
}

func KonnectStateToFile(kongState *state.KongState, config WriteConfig) error {
	file := &Content{}
	file.FormatVersion = "0.1"
	var err error

	err = populateServicePackages(kongState, file, config)
	if err != nil {
		return err
	}

	// do not populate service-less routes
	// we do not know if konnect supports these or not

	err = populatePlugins(kongState, file, config)
	if err != nil {
		return err
	}

	err = populateUpstreams(kongState, file, config)
	if err != nil {
		return err
	}

	err = populateCertificates(kongState, file, config)
	if err != nil {
		return err
	}

	err = populateCACertificates(kongState, file, config)
	if err != nil {
		return err
	}

	err = populateConsumers(kongState, file, config)
	if err != nil {
		return err
	}

	return WriteContentToFile(file, config.Filename, config.FileFormat)
}

func populateServicePackages(kongState *state.KongState, file *Content,
	config WriteConfig,
) error {
	packages, err := kongState.ServicePackages.GetAll()
	if err != nil {
		return err
	}

	for _, sp := range packages {
		safePackageName := utils.NameToFilename(*sp.Name)
		p := FServicePackage{
			ID:          sp.ID,
			Name:        sp.Name,
			Description: sp.Description,
		}
		versions, err := kongState.ServiceVersions.GetAllByServicePackageID(*p.ID)
		if err != nil {
			return err
		}
		documents, err := kongState.Documents.GetAllByParent(sp)
		if err != nil {
			return err
		}

		for _, d := range documents {
			safeDocPath := utils.NameToFilename(*d.Path)
			fDocument := FDocument{
				ID:        d.ID,
				Path:      kong.String(filepath.Join(safePackageName, safeDocPath)),
				Published: d.Published,
				Content:   d.Content,
			}
			utils.ZeroOutID(&fDocument, fDocument.Path, config.WithID)
			p.Document = &fDocument
			// Although the documents API returns a list of documents and does support multiple documents,
			// we pretend there's only one because that's all the web UI allows.
			break
		}

		for _, v := range versions {
			safeVersionName := utils.NameToFilename(*v.Version)
			fVersion := FServiceVersion{
				ID:      v.ID,
				Version: v.Version,
			}
			if v.ControlPlaneServiceRelation != nil &&
				!utils.Empty(v.ControlPlaneServiceRelation.ControlPlaneEntityID) {
				kongServiceID := *v.ControlPlaneServiceRelation.ControlPlaneEntityID

				s, err := fetchService(kongServiceID, kongState, config)
				if err != nil {
					return err
				}
				fVersion.Implementation = &Implementation{
					Type: utils.ImplementationTypeKongGateway,
					Kong: &Kong{
						Service: s,
					},
				}
			}
			documents, err := kongState.Documents.GetAllByParent(v)
			if err != nil {
				return err
			}

			for _, d := range documents {
				safeDocPath := utils.NameToFilename(*d.Path)
				fDocument := FDocument{
					ID:        d.ID,
					Path:      kong.String(filepath.Join(safePackageName, safeVersionName, safeDocPath)),
					Published: d.Published,
					Content:   d.Content,
				}
				utils.ZeroOutID(&fDocument, fDocument.Path, config.WithID)
				fVersion.Document = &fDocument
				break
			}
			utils.ZeroOutID(&fVersion, fVersion.Version, config.WithID)
			p.Versions = append(p.Versions, fVersion)
		}
		sort.SliceStable(p.Versions, func(i, j int) bool {
			return compareOrder(p.Versions[i], p.Versions[j])
		})
		utils.ZeroOutID(&p, p.Name, config.WithID)
		file.ServicePackages = append(file.ServicePackages, p)
	}
	sort.SliceStable(file.ServicePackages, func(i, j int) bool {
		return compareOrder(file.ServicePackages[i], file.ServicePackages[j])
	})
	return nil
}

func populateServices(kongState *state.KongState, file *Content,
	config WriteConfig,
) error {
	services, err := kongState.Services.GetAll()
	if err != nil {
		return err
	}
	for _, s := range services {
		s, err := fetchService(*s.ID, kongState, config)
		if err != nil {
			return err
		}
		file.Services = append(file.Services, *s)
	}
	sort.SliceStable(file.Services, func(i, j int) bool {
		return compareOrder(file.Services[i], file.Services[j])
	})
	return nil
}

func fetchService(id string, kongState *state.KongState, config WriteConfig) (*FService, error) {
	kongService, err := kongState.Services.Get(id)
	if err != nil {
		return nil, err
	}
	s := FService{Service: kongService.Service}
	routes, err := kongState.Routes.GetAllByServiceID(*s.ID)
	if err != nil {
		return nil, err
	}
	plugins, err := kongState.Plugins.GetAllByServiceID(*s.ID)
	if err != nil {
		return nil, err
	}
	for _, p := range plugins {
		if p.Route != nil || p.Consumer != nil {
			continue
		}
		p.Service = nil
		utils.ZeroOutID(p, p.Name, config.WithID)
		utils.ZeroOutTimestamps(p)
		utils.MustRemoveTags(&p.Plugin, config.SelectTags)
		s.Plugins = append(s.Plugins, &FPlugin{Plugin: p.Plugin})
	}
	sort.SliceStable(s.Plugins, func(i, j int) bool {
		return compareOrder(s.Plugins[i], s.Plugins[j])
	})
	for _, r := range routes {
		plugins, err := kongState.Plugins.GetAllByRouteID(*r.ID)
		if err != nil {
			return nil, err
		}
		r.Service = nil
		utils.ZeroOutID(r, r.Name, config.WithID)
		utils.ZeroOutTimestamps(r)
		utils.MustRemoveTags(&r.Route, config.SelectTags)
		route := &FRoute{Route: r.Route}
		for _, p := range plugins {
			if p.Service != nil || p.Consumer != nil {
				continue
			}
			p.Route = nil
			utils.ZeroOutID(p, p.Name, config.WithID)
			utils.ZeroOutTimestamps(p)
			utils.MustRemoveTags(&p.Plugin, config.SelectTags)
			route.Plugins = append(route.Plugins, &FPlugin{Plugin: p.Plugin})
		}
		sort.SliceStable(route.Plugins, func(i, j int) bool {
			return compareOrder(route.Plugins[i], route.Plugins[j])
		})
		s.Routes = append(s.Routes, route)
	}
	sort.SliceStable(s.Routes, func(i, j int) bool {
		return compareOrder(s.Routes[i], s.Routes[j])
	})
	utils.ZeroOutID(&s, s.Name, config.WithID)
	utils.ZeroOutTimestamps(&s)
	utils.MustRemoveTags(&s, config.SelectTags)
	return &s, nil
}

func populateServicelessRoutes(kongState *state.KongState, file *Content,
	config WriteConfig,
) error {
	routes, err := kongState.Routes.GetAll()
	if err != nil {
		return err
	}
	for _, r := range routes {
		if r.Service != nil {
			continue
		}
		plugins, err := kongState.Plugins.GetAllByRouteID(*r.ID)
		if err != nil {
			return err
		}
		utils.ZeroOutID(r, r.Name, config.WithID)
		utils.ZeroOutTimestamps(r)
		utils.MustRemoveTags(&r.Route, config.SelectTags)
		route := &FRoute{Route: r.Route}
		for _, p := range plugins {
			if p.Service != nil || p.Consumer != nil {
				continue
			}
			p.Route = nil
			utils.ZeroOutID(p, p.Name, config.WithID)
			utils.ZeroOutTimestamps(p)
			utils.MustRemoveTags(&p.Plugin, config.SelectTags)
			route.Plugins = append(route.Plugins, &FPlugin{Plugin: p.Plugin})
		}
		sort.SliceStable(route.Plugins, func(i, j int) bool {
			return compareOrder(route.Plugins[i], route.Plugins[j])
		})
		file.Routes = append(file.Routes, *route)
	}
	sort.SliceStable(file.Routes, func(i, j int) bool {
		return compareOrder(file.Routes[i], file.Routes[j])
	})
	return nil
}

func populatePlugins(kongState *state.KongState, file *Content,
	config WriteConfig,
) error {
	plugins, err := kongState.Plugins.GetAll()
	if err != nil {
		return err
	}
	for _, p := range plugins {
		associations := 0
		if p.Consumer != nil {
			associations++
			cID := *p.Consumer.ID
			consumer, err := kongState.Consumers.GetByIDOrUsername(cID)
			if err != nil {
				return fmt.Errorf("unable to get consumer %s for plugin %s [%s]: %w", cID, *p.Name, *p.ID, err)
			}
			if !utils.Empty(consumer.Username) {
				cID = *consumer.Username
			}
			p.Consumer.ID = &cID
		}
		if p.Service != nil {
			associations++
			sID := *p.Service.ID
			service, err := kongState.Services.Get(sID)
			if err != nil {
				return fmt.Errorf("unable to get service %s for plugin %s [%s]: %w", sID, *p.Name, *p.ID, err)
			}
			if !utils.Empty(service.Name) {
				sID = *service.Name
			}
			p.Service.ID = &sID
		}
		if p.Route != nil {
			associations++
			rID := *p.Route.ID
			route, err := kongState.Routes.Get(rID)
			if err != nil {
				return fmt.Errorf("unable to get route %s for plugin %s [%s]: %w", rID, *p.Name, *p.ID, err)
			}
			if !utils.Empty(route.Name) {
				rID = *route.Name
			}
			p.Route.ID = &rID
		}
		if p.ConsumerGroup != nil {
			associations++
			cgID := *p.ConsumerGroup.ID
			cg, err := kongState.ConsumerGroups.Get(cgID)
			if err != nil {
				return fmt.Errorf("unable to get consumer-group %s for plugin %s [%s]: %w", cgID, *p.Name, *p.ID, err)
			}
			if !utils.Empty(cg.Name) {
				cgID = *cg.Name
			}
			p.ConsumerGroup.ID = &cgID
		}
		if associations == 0 || associations > 1 {
			utils.ZeroOutID(p, p.Name, config.WithID)
			utils.ZeroOutTimestamps(p)
			utils.MustRemoveTags(&p.Plugin, config.SelectTags)
			p := FPlugin{Plugin: p.Plugin}
			file.Plugins = append(file.Plugins, p)
		}
	}
	sort.SliceStable(file.Plugins, func(i, j int) bool {
		return compareOrder(file.Plugins[i], file.Plugins[j])
	})
	return nil
}

func populateUpstreams(kongState *state.KongState, file *Content,
	config WriteConfig,
) error {
	upstreams, err := kongState.Upstreams.GetAll()
	if err != nil {
		return err
	}
	for _, u := range upstreams {
		u := FUpstream{Upstream: u.Upstream}
		targets, err := kongState.Targets.GetAllByUpstreamID(*u.ID)
		if err != nil {
			return err
		}
		for _, t := range targets {
			t.Upstream = nil
			utils.ZeroOutID(t, t.Target.Target, config.WithID)
			utils.ZeroOutTimestamps(t)
			utils.MustRemoveTags(&t.Target, config.SelectTags)
			u.Targets = append(u.Targets, &FTarget{Target: t.Target})
		}
		sort.SliceStable(u.Targets, func(i, j int) bool {
			return compareOrder(u.Targets[i], u.Targets[j])
		})
		utils.ZeroOutID(&u, u.Name, config.WithID)
		utils.ZeroOutTimestamps(&u)
		utils.MustRemoveTags(&u.Upstream, config.SelectTags)
		file.Upstreams = append(file.Upstreams, u)
	}
	sort.SliceStable(file.Upstreams, func(i, j int) bool {
		return compareOrder(file.Upstreams[i], file.Upstreams[j])
	})
	return nil
}

func populateVaults(kongState *state.KongState, file *Content,
	config WriteConfig,
) error {
	vaults, err := kongState.Vaults.GetAll()
	if err != nil {
		return err
	}
	for _, v := range vaults {
		v := FVault{Vault: v.Vault}
		utils.ZeroOutID(&v, v.Prefix, config.WithID)
		utils.ZeroOutTimestamps(&v)
		utils.MustRemoveTags(&v.Vault, config.SelectTags)
		file.Vaults = append(file.Vaults, v)
	}
	sort.SliceStable(file.Vaults, func(i, j int) bool {
		return compareOrder(file.Vaults[i], file.Vaults[j])
	})
	return nil
}

func populateCertificates(kongState *state.KongState, file *Content,
	config WriteConfig,
) error {
	certificates, err := kongState.Certificates.GetAll()
	if err != nil {
		return err
	}
	for _, c := range certificates {
		c := FCertificate{
			ID:   c.ID,
			Cert: c.Cert,
			Key:  c.Key,
			Tags: c.Tags,
		}
		snis, err := kongState.SNIs.GetAllByCertID(*c.ID)
		if err != nil {
			return err
		}
		for _, s := range snis {
			s.Certificate = nil
			utils.ZeroOutID(s, s.Name, config.WithID)
			utils.ZeroOutTimestamps(s)
			utils.MustRemoveTags(&s.SNI, config.SelectTags)
			c.SNIs = append(c.SNIs, s.SNI)
		}
		sort.SliceStable(c.SNIs, func(i, j int) bool {
			return strings.Compare(*c.SNIs[i].Name, *c.SNIs[j].Name) < 0
		})
		utils.ZeroOutTimestamps(&c)
		utils.MustRemoveTags(&c, config.SelectTags)
		file.Certificates = append(file.Certificates, c)
	}
	sort.SliceStable(file.Certificates, func(i, j int) bool {
		return compareOrder(file.Certificates[i], file.Certificates[j])
	})
	return nil
}

func populateCACertificates(kongState *state.KongState, file *Content,
	config WriteConfig,
) error {
	caCertificates, err := kongState.CACertificates.GetAll()
	if err != nil {
		return err
	}
	for _, c := range caCertificates {
		c := FCACertificate{CACertificate: c.CACertificate}
		utils.ZeroOutTimestamps(&c)
		utils.MustRemoveTags(&c.CACertificate, config.SelectTags)
		file.CACertificates = append(file.CACertificates, c)
	}
	sort.SliceStable(file.CACertificates, func(i, j int) bool {
		return compareOrder(file.CACertificates[i], file.CACertificates[j])
	})
	return nil
}

func populateConsumers(kongState *state.KongState, file *Content,
	config WriteConfig,
) error {
	consumers, err := kongState.Consumers.GetAll()
	if err != nil {
		return err
	}
	consumerGroups, err := kongState.ConsumerGroups.GetAll()
	if err != nil {
		return err
	}
	for _, c := range consumers {
		c := FConsumer{Consumer: c.Consumer}
		plugins, err := kongState.Plugins.GetAllByConsumerID(*c.ID)
		if err != nil {
			return err
		}
		for _, p := range plugins {
			if p.Service != nil || p.Route != nil {
				continue
			}
			utils.ZeroOutID(p, p.Name, config.WithID)
			utils.ZeroOutTimestamps(p)
			p.Consumer = nil
			utils.MustRemoveTags(&p.Plugin, config.SelectTags)
			c.Plugins = append(c.Plugins, &FPlugin{Plugin: p.Plugin})
		}
		sort.SliceStable(c.Plugins, func(i, j int) bool {
			return compareOrder(c.Plugins[i], c.Plugins[j])
		})
		// custom-entities associated with Consumer
		keyAuths, err := kongState.KeyAuths.GetAllByConsumerID(*c.ID)
		if err != nil {
			return err
		}
		for _, k := range keyAuths {
			utils.ZeroOutID(k, k.Key, config.WithID)
			utils.ZeroOutTimestamps(k)
			utils.MustRemoveTags(k, config.SelectTags)
			k.Consumer = nil
			c.KeyAuths = append(c.KeyAuths, &k.KeyAuth)
		}
		hmacAuth, err := kongState.HMACAuths.GetAllByConsumerID(*c.ID)
		if err != nil {
			return err
		}
		for _, k := range hmacAuth {
			k.Consumer = nil
			utils.ZeroOutID(k, k.Username, config.WithID)
			utils.ZeroOutTimestamps(k)
			utils.MustRemoveTags(k, config.SelectTags)
			c.HMACAuths = append(c.HMACAuths, &k.HMACAuth)
		}
		jwtSecrets, err := kongState.JWTAuths.GetAllByConsumerID(*c.ID)
		if err != nil {
			return err
		}
		for _, k := range jwtSecrets {
			k.Consumer = nil
			utils.ZeroOutID(k, k.Key, config.WithID)
			utils.ZeroOutTimestamps(k)
			utils.MustRemoveTags(k, config.SelectTags)
			c.JWTAuths = append(c.JWTAuths, &k.JWTAuth)
		}
		basicAuths, err := kongState.BasicAuths.GetAllByConsumerID(*c.ID)
		if err != nil {
			return err
		}
		for _, k := range basicAuths {
			k.Consumer = nil
			utils.ZeroOutID(k, k.Username, config.WithID)
			utils.ZeroOutTimestamps(k)
			utils.MustRemoveTags(k, config.SelectTags)
			c.BasicAuths = append(c.BasicAuths, &k.BasicAuth)
		}
		oauth2Creds, err := kongState.Oauth2Creds.GetAllByConsumerID(*c.ID)
		if err != nil {
			return err
		}
		for _, k := range oauth2Creds {
			k.Consumer = nil
			utils.ZeroOutID(k, k.ClientID, config.WithID)
			utils.ZeroOutTimestamps(k)
			utils.MustRemoveTags(k, config.SelectTags)
			c.Oauth2Creds = append(c.Oauth2Creds, &k.Oauth2Credential)
		}
		aclGroups, err := kongState.ACLGroups.GetAllByConsumerID(*c.ID)
		if err != nil {
			return err
		}
		for _, k := range aclGroups {
			k.Consumer = nil
			utils.ZeroOutID(k, k.Group, config.WithID)
			utils.ZeroOutTimestamps(k)
			utils.MustRemoveTags(k, config.SelectTags)
			c.ACLGroups = append(c.ACLGroups, &k.ACLGroup)
		}
		mtlsAuths, err := kongState.MTLSAuths.GetAllByConsumerID(*c.ID)
		if err != nil {
			return err
		}
		for _, k := range mtlsAuths {
			utils.ZeroOutTimestamps(k)
			utils.MustRemoveTags(k, config.SelectTags)
			k.Consumer = nil
			c.MTLSAuths = append(c.MTLSAuths, &k.MTLSAuth)
		}
		// populate groups
		for _, cg := range consumerGroups {
			cg := *cg
			_, err := kongState.ConsumerGroupConsumers.Get(*c.ID, *cg.ID)
			if err != nil {
				if !errors.Is(err, state.ErrNotFound) {
					return err
				}
				continue
			}
			utils.ZeroOutID(&cg, cg.Name, config.WithID)
			utils.ZeroOutTimestamps(&cg)
			utils.MustRemoveTags(&cg.ConsumerGroup, config.SelectTags)
			c.Groups = append(c.Groups, cg.DeepCopy())
		}
		sort.SliceStable(c.Plugins, func(i, j int) bool {
			return compareOrder(c.Plugins[i], c.Plugins[j])
		})
		utils.ZeroOutID(&c, c.Username, config.WithID)
		utils.ZeroOutTimestamps(&c)
		utils.MustRemoveTags(&c.Consumer, config.SelectTags)
		file.Consumers = append(file.Consumers, c)
	}
	rbacRoles, err := kongState.RBACRoles.GetAll()
	if err != nil {
		return err
	}
	for _, r := range rbacRoles {
		r := FRBACRole{RBACRole: r.RBACRole}
		eps, err := kongState.RBACEndpointPermissions.GetAllByRoleID(*r.ID)
		if err != nil {
			return err
		}
		for _, ep := range eps {
			ep.Role = nil
			utils.ZeroOutTimestamps(ep)
			r.EndpointPermissions = append(
				r.EndpointPermissions, &FRBACEndpointPermission{RBACEndpointPermission: ep.RBACEndpointPermission})
		}
		utils.ZeroOutID(&r, r.Name, config.WithID)
		utils.ZeroOutTimestamps(&r)
		file.RBACRoles = append(file.RBACRoles, r)
	}
	sort.SliceStable(file.Consumers, func(i, j int) bool {
		return compareOrder(file.Consumers[i], file.Consumers[j])
	})
	return nil
}

func populateConsumerGroups(kongState *state.KongState, file *Content,
	config WriteConfig,
) error {
	consumerGroups, err := kongState.ConsumerGroups.GetAll()
	if err != nil {
		return err
	}
	cgPlugins, err := kongState.ConsumerGroupPlugins.GetAll()
	if err != nil {
		return err
	}
	for _, cg := range consumerGroups {
		group := FConsumerGroupObject{ConsumerGroup: cg.ConsumerGroup}
		for _, plugin := range cgPlugins {
			if plugin.ID != nil && cg.ID != nil {
				if plugin.ConsumerGroup != nil && *plugin.ConsumerGroup.ID == *cg.ID {
					utils.ZeroOutID(plugin, plugin.Name, config.WithID)
					utils.ZeroOutID(plugin.ConsumerGroup, plugin.ConsumerGroup.Name, config.WithID)
					utils.ZeroOutTimestamps(plugin.ConsumerGroupPlugin.ConsumerGroup)
					utils.ZeroOutField(&plugin.ConsumerGroupPlugin, "ConsumerGroup")
					group.Plugins = append(group.Plugins, &plugin.ConsumerGroupPlugin)
				}
			}
		}

		plugins, err := kongState.Plugins.GetAllByConsumerGroupID(*cg.ID)
		if err != nil {
			return err
		}
		for _, plugin := range plugins {
			if plugin.ID != nil && cg.ID != nil {
				if plugin.ConsumerGroup != nil && *plugin.ConsumerGroup.ID == *cg.ID {
					utils.ZeroOutID(plugin, plugin.Name, config.WithID)
					utils.ZeroOutID(plugin.ConsumerGroup, plugin.ConsumerGroup.Name, config.WithID)
					group.Plugins = append(group.Plugins, &kong.ConsumerGroupPlugin{
						ID:     plugin.ID,
						Name:   plugin.Name,
						Config: plugin.Config,
					})
				}
			}
		}

		utils.ZeroOutID(&group, group.Name, config.WithID)
		utils.ZeroOutTimestamps(&group)
		file.ConsumerGroups = append(file.ConsumerGroups, group)
	}
	sort.SliceStable(file.ConsumerGroups, func(i, j int) bool {
		return compareOrder(file.ConsumerGroups[i], file.ConsumerGroups[j])
	})
	return nil
}

func WriteContentToFile(content *Content, filename string, format Format) error {
	var c []byte
	var err error
	switch format {
	case YAML:
		c, err = yaml.Marshal(content)
		if err != nil {
			return err
		}
	case JSON:
		c, err = json.MarshalIndent(content, "", "  ")
		if err != nil {
			return err
		}
	case KIC_JSON_CRD:
		c, err = MarshalKongToKICJson(content, CUSTOM_RESOURCE)
		if err != nil {
			return err
		}
	case KIC_YAML_CRD:
		c, err = MarshalKongToKICYaml(content, CUSTOM_RESOURCE)
		if err != nil {
			return err
		}
	case KIC_JSON_ANNOTATION:
		c, err = MarshalKongToKICJson(content, ANNOTATIONS)
		if err != nil {
			return err
		}
	case KIC_YAML_ANNOTATION:
		c, err = MarshalKongToKICYaml(content, ANNOTATIONS)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown file format: " + string(format))
	}

	if filename == "-" {
		if _, err := fmt.Print(string(c)); err != nil {
			return fmt.Errorf("writing file: %w", err)
		}
	} else {
		filename = utils.AddExtToFilename(filename, strings.ToLower(string(format)))
		prefix, _ := filepath.Split(filename)
		if err := ioutil.WriteFile(filename, c, 0o600); err != nil {
			return fmt.Errorf("writing file: %w", err)
		}
		for _, sp := range content.ServicePackages {
			if sp.Document != nil {
				if err := os.MkdirAll(filepath.Join(prefix, filepath.Dir(*sp.Document.Path)), 0o700); err != nil {
					return fmt.Errorf("creating document directory: %w", err)
				}
				if err := os.WriteFile(filepath.Join(prefix, *sp.Document.Path),
					[]byte(*sp.Document.Content), 0o600); err != nil {
					return fmt.Errorf("writing document file: %w", err)
				}
			}
			for _, v := range sp.Versions {
				if v.Document != nil {
					if err := os.MkdirAll(filepath.Join(prefix, filepath.Dir(*v.Document.Path)), 0o700); err != nil {
						return fmt.Errorf("creating document directory: %w", err)
					}
					if err := os.WriteFile(filepath.Join(prefix, *v.Document.Path),
						[]byte(*v.Document.Content), 0o600); err != nil {
						return fmt.Errorf("writing document file: %w", err)
					}
				}
			}
		}
	}
	return nil
}
