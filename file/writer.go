package file

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// WriteConfig holds settings to use to write the state file.
type WriteConfig struct {
	Workspace  string
	SelectTags []string
	Filename   string
	FileFormat Format
	WithID     bool
}

func compareID(obj1, obj2 id) bool {
	return strings.Compare(obj1.id(), obj2.id()) < 0
}

// KongStateToFile writes a state object to file with filename.
// It will omit timestamps and IDs while writing.
func KongStateToFile(kongState *state.KongState, config WriteConfig) error {
	// TODO break-down this giant function
	var file Content

	file.Workspace = config.Workspace
	// hardcoded as only one version exists currently
	file.FormatVersion = "1.1"

	selectTags := config.SelectTags
	if len(selectTags) > 0 {
		file.Info = &Info{
			SelectorTags: selectTags,
		}
	}

	services, err := kongState.Services.GetAll()
	if err != nil {
		return err
	}
	for _, s := range services {
		s := FService{Service: s.Service}
		routes, err := kongState.Routes.GetAllByServiceID(*s.ID)
		if err != nil {
			return err
		}
		plugins, err := kongState.Plugins.GetAllByServiceID(*s.ID)
		if err != nil {
			return err
		}
		for _, p := range plugins {
			if p.Route != nil || p.Consumer != nil {
				continue
			}
			p.Service = nil
			zeroOutID(p, p.Name, config.WithID)
			zeroOutTimestamps(p)
			utils.MustRemoveTags(&p.Plugin, selectTags)
			s.Plugins = append(s.Plugins, &FPlugin{Plugin: p.Plugin})
		}
		sort.SliceStable(s.Plugins, func(i, j int) bool {
			return compareID(s.Plugins[i], s.Plugins[j])
		})
		for _, r := range routes {
			plugins, err := kongState.Plugins.GetAllByRouteID(*r.ID)
			if err != nil {
				return err
			}
			r.Service = nil
			zeroOutID(r, r.Name, config.WithID)
			zeroOutTimestamps(r)
			utils.MustRemoveTags(&r.Route, selectTags)
			route := &FRoute{Route: r.Route}
			for _, p := range plugins {
				if p.Service != nil || p.Consumer != nil {
					continue
				}
				p.Route = nil
				zeroOutID(p, p.Name, config.WithID)
				zeroOutTimestamps(p)
				utils.MustRemoveTags(&p.Plugin, selectTags)
				route.Plugins = append(route.Plugins, &FPlugin{Plugin: p.Plugin})
			}
			sort.SliceStable(route.Plugins, func(i, j int) bool {
				return compareID(route.Plugins[i], route.Plugins[j])
			})
			s.Routes = append(s.Routes, route)
		}
		sort.SliceStable(s.Routes, func(i, j int) bool {
			return compareID(s.Routes[i], s.Routes[j])
		})
		zeroOutID(&s, s.Name, config.WithID)
		zeroOutTimestamps(&s)
		utils.MustRemoveTags(&s.Service, selectTags)
		file.Services = append(file.Services, s)
	}
	sort.SliceStable(file.Services, func(i, j int) bool {
		return compareID(file.Services[i], file.Services[j])
	})
	// service-less routes
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
		zeroOutID(r, r.Name, config.WithID)
		zeroOutTimestamps(r)
		utils.MustRemoveTags(&r.Route, selectTags)
		route := &FRoute{Route: r.Route}
		for _, p := range plugins {
			if p.Service != nil || p.Consumer != nil {
				continue
			}
			p.Route = nil
			zeroOutID(p, p.Name, config.WithID)
			zeroOutTimestamps(p)
			utils.MustRemoveTags(&p.Plugin, selectTags)
			route.Plugins = append(route.Plugins, &FPlugin{Plugin: p.Plugin})
		}
		sort.SliceStable(route.Plugins, func(i, j int) bool {
			return compareID(route.Plugins[i], route.Plugins[j])
		})
		file.Routes = append(file.Routes, *route)
	}
	sort.SliceStable(file.Routes, func(i, j int) bool {
		return compareID(file.Routes[i], file.Routes[j])
	})

	// Add global and multi-relational plugins
	plugins, err := kongState.Plugins.GetAll()
	if err != nil {
		return err
	}
	for _, p := range plugins {
		associations := 0
		if p.Consumer != nil {
			associations++
			cID := *p.Consumer.ID
			consumer, err := kongState.Consumers.Get(cID)
			if err != nil {
				return err
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
				return err
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
				return err
			}
			if !utils.Empty(route.Name) {
				rID = *route.Name
			}
			p.Route.ID = &rID
		}
		if associations == 0 || associations > 1 {
			zeroOutID(p, p.Name, config.WithID)
			zeroOutTimestamps(p)
			utils.MustRemoveTags(&p.Plugin, selectTags)
			p := FPlugin{Plugin: p.Plugin}
			file.Plugins = append(file.Plugins, p)
		}
	}
	sort.SliceStable(file.Plugins, func(i, j int) bool {
		return compareID(file.Plugins[i], file.Plugins[j])
	})

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
			zeroOutID(t, t.Target.Target, config.WithID)
			zeroOutTimestamps(t)
			utils.MustRemoveTags(&t.Target, selectTags)
			u.Targets = append(u.Targets, &FTarget{Target: t.Target})
		}
		sort.SliceStable(u.Targets, func(i, j int) bool {
			return compareID(u.Targets[i], u.Targets[j])
		})
		zeroOutID(&u, u.Name, config.WithID)
		zeroOutTimestamps(&u)
		utils.MustRemoveTags(&u.Upstream, selectTags)
		file.Upstreams = append(file.Upstreams, u)
	}
	sort.SliceStable(file.Upstreams, func(i, j int) bool {
		return compareID(file.Upstreams[i], file.Upstreams[j])
	})

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
			zeroOutID(s, s.Name, config.WithID)
			zeroOutTimestamps(s)
			utils.MustRemoveTags(&s.SNI, selectTags)
			c.SNIs = append(c.SNIs, s.SNI)
		}
		sort.SliceStable(c.SNIs, func(i, j int) bool {
			return strings.Compare(*c.SNIs[i].Name, *c.SNIs[j].Name) < 0
		})
		zeroOutTimestamps(&c)
		utils.MustRemoveTags(&c, selectTags)
		file.Certificates = append(file.Certificates, c)
	}
	sort.SliceStable(file.Certificates, func(i, j int) bool {
		return compareID(file.Certificates[i], file.Certificates[j])
	})

	caCertificates, err := kongState.CACertificates.GetAll()
	if err != nil {
		return err
	}
	for _, c := range caCertificates {
		c := FCACertificate{CACertificate: c.CACertificate}
		zeroOutID(&c, c.Cert, config.WithID)
		zeroOutTimestamps(&c)
		utils.MustRemoveTags(&c.CACertificate, selectTags)
		file.CACertificates = append(file.CACertificates, c)
	}
	sort.SliceStable(file.CACertificates, func(i, j int) bool {
		return compareID(file.CACertificates[i], file.CACertificates[j])
	})

	consumers, err := kongState.Consumers.GetAll()
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
			zeroOutID(p, p.Name, config.WithID)
			zeroOutTimestamps(p)
			p.Consumer = nil
			utils.MustRemoveTags(&p.Plugin, selectTags)
			c.Plugins = append(c.Plugins, &FPlugin{Plugin: p.Plugin})
		}
		sort.SliceStable(c.Plugins, func(i, j int) bool {
			return compareID(c.Plugins[i], c.Plugins[j])
		})
		// custom-entities associated with Consumer
		keyAuths, err := kongState.KeyAuths.GetAllByConsumerID(*c.ID)
		if err != nil {
			return err
		}
		for _, k := range keyAuths {
			zeroOutID(k, k.Key, config.WithID)
			zeroOutTimestamps(k)
			k.Consumer = nil
			c.KeyAuths = append(c.KeyAuths, &k.KeyAuth)
		}
		hmacAuth, err := kongState.HMACAuths.GetAllByConsumerID(*c.ID)
		if err != nil {
			return err
		}
		for _, k := range hmacAuth {
			k.Consumer = nil
			zeroOutID(k, k.Username, config.WithID)
			zeroOutTimestamps(k)
			c.HMACAuths = append(c.HMACAuths, &k.HMACAuth)
		}
		jwtSecrets, err := kongState.JWTAuths.GetAllByConsumerID(*c.ID)
		if err != nil {
			return err
		}
		for _, k := range jwtSecrets {
			k.Consumer = nil
			zeroOutID(k, k.Key, config.WithID)
			zeroOutTimestamps(k)
			c.JWTAuths = append(c.JWTAuths, &k.JWTAuth)
		}
		basicAuths, err := kongState.BasicAuths.GetAllByConsumerID(*c.ID)
		if err != nil {
			return err
		}
		for _, k := range basicAuths {
			k.Consumer = nil
			zeroOutID(k, k.Username, config.WithID)
			zeroOutTimestamps(k)
			c.BasicAuths = append(c.BasicAuths, &k.BasicAuth)
		}
		oauth2Creds, err := kongState.Oauth2Creds.GetAllByConsumerID(*c.ID)
		if err != nil {
			return err
		}
		for _, k := range oauth2Creds {
			k.Consumer = nil
			zeroOutID(k, k.ClientID, config.WithID)
			zeroOutTimestamps(k)
			c.Oauth2Creds = append(c.Oauth2Creds, &k.Oauth2Credential)
		}
		aclGroups, err := kongState.ACLGroups.GetAllByConsumerID(*c.ID)
		if err != nil {
			return err
		}
		for _, k := range aclGroups {
			k.Consumer = nil
			zeroOutID(k, k.Group, config.WithID)
			zeroOutTimestamps(k)
			c.ACLGroups = append(c.ACLGroups, &k.ACLGroup)
		}
		zeroOutID(&c, c.Username, config.WithID)
		zeroOutTimestamps(&c)
		utils.MustRemoveTags(&c.Consumer, selectTags)
		file.Consumers = append(file.Consumers, c)
	}
	sort.SliceStable(file.Consumers, func(i, j int) bool {
		return compareID(file.Consumers[i], file.Consumers[j])
	})

	return writeFile(file, config.Filename, config.FileFormat)
}

func writeFile(content Content, filename string, format Format) error {
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
	default:
		return errors.New("unknown file format: " + string(format))
	}

	if filename == "-" {
		_, err = fmt.Print(string(c))
	} else {
		filename = addExtToFilename(filename, string(format))
		err = ioutil.WriteFile(filename, c, 0600)
	}
	if err != nil {
		return errors.Wrap(err, "writing file")
	}
	return nil
}

func addExtToFilename(filename, format string) string {
	if filepath.Ext(filename) == "" {
		filename = filename + "." + strings.ToLower(format)
	}
	return filename
}

func zeroOutTimestamps(obj interface{}) {
	zeroOutField(obj, "CreatedAt")
	zeroOutField(obj, "UpdatedAt")
}

var zero reflect.Value

func zeroOutField(obj interface{}, field string) {
	ptr := reflect.ValueOf(obj)
	if ptr.Kind() != reflect.Ptr {
		return
	}
	v := reflect.Indirect(ptr)
	ts := v.FieldByName(field)
	if ts == zero {
		return
	}
	ts.Set(reflect.Zero(ts.Type()))
}

func zeroOutID(obj interface{}, altName *string, withID bool) {
	// withID is set, export the ID
	if withID {
		return
	}
	// altName is not set, export the ID
	if utils.Empty(altName) {
		return
	}
	// zero the ID field
	zeroOutField(obj, "ID")
}
