package file

import (
	"fmt"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	yaml "gopkg.in/yaml.v2"
)

// KongStateToFile writes a state object to file with filename.
// It will omit timestamps and IDs while writing.
func KongStateToFile(kongState *state.KongState,
	// TODO break-down this giant function
	selectTags []string, filename string) error {
	var file Content

	services, err := kongState.Services.GetAll()
	if err != nil {
		return err
	}
	for _, s := range services {
		s := Service{Service: s.Service}
		routes, err := kongState.Routes.GetAllByServiceID(*s.ID)
		if err != nil {
			return err
		}
		plugins, err := kongState.Plugins.GetAllByServiceID(*s.ID)
		if err != nil {
			return err
		}
		for _, p := range plugins {
			p.ID = nil
			p.CreatedAt = nil
			p.Service = nil
			utils.RemoveTags(&p.Plugin, selectTags)
			s.Plugins = append(s.Plugins, &Plugin{Plugin: p.Plugin})
		}
		sort.SliceStable(s.Plugins, func(i, j int) bool {
			return strings.Compare(*s.Plugins[i].Name, *s.Plugins[j].Name) < 0
		})
		for _, r := range routes {
			plugins, err := kongState.Plugins.GetAllByRouteID(*r.ID)
			if err != nil {
				return err
			}
			r.Service = nil
			r.ID = nil
			r.CreatedAt = nil
			r.UpdatedAt = nil
			route := &Route{Route: r.Route}
			utils.RemoveTags(&route.Route, selectTags)
			for _, p := range plugins {
				p.ID = nil
				p.CreatedAt = nil
				p.Route = nil
				utils.RemoveTags(&p.Plugin, selectTags)
				route.Plugins = append(route.Plugins, &Plugin{Plugin: p.Plugin})
			}
			sort.SliceStable(route.Plugins, func(i, j int) bool {
				return strings.Compare(*route.Plugins[i].Name, *route.Plugins[j].Name) < 0
			})
			s.Routes = append(s.Routes, route)
		}
		sort.SliceStable(s.Routes, func(i, j int) bool {
			return strings.Compare(*s.Routes[i].Name, *s.Routes[j].Name) < 0
		})
		s.ID = nil
		s.CreatedAt = nil
		s.UpdatedAt = nil
		utils.RemoveTags(&s.Service, selectTags)
		file.Services = append(file.Services, s)
	}
	sort.SliceStable(file.Services, func(i, j int) bool {
		return strings.Compare(*file.Services[i].Name,
			*file.Services[j].Name) < 0
	})

	// Add global plugins
	plugins, err := kongState.Plugins.GetAll()
	if err != nil {
		return err
	}
	for _, p := range plugins {
		if p.Consumer == nil && p.Service == nil && p.Route == nil {
			p.ID = nil
			p.CreatedAt = nil
			p := Plugin{Plugin: p.Plugin}
			utils.RemoveTags(&p.Plugin, selectTags)
			file.Plugins = append(file.Plugins, p)
		}
	}
	sort.SliceStable(file.Plugins, func(i, j int) bool {
		return strings.Compare(*file.Plugins[i].Name,
			*file.Plugins[j].Name) < 0
	})

	upstreams, err := kongState.Upstreams.GetAll()
	if err != nil {
		return err
	}
	for _, u := range upstreams {
		u := Upstream{Upstream: u.Upstream}
		targets, err := kongState.Targets.GetAllByUpstreamID(*u.ID)
		if err != nil {
			return err
		}
		for _, t := range targets {
			t.Upstream = nil
			t.ID = nil
			t.CreatedAt = nil
			utils.RemoveTags(&t.Target, selectTags)
			u.Targets = append(u.Targets, &Target{Target: t.Target})
		}
		sort.SliceStable(u.Targets, func(i, j int) bool {
			return strings.Compare(*u.Targets[i].Target.Target,
				*u.Targets[j].Target.Target) < 0
		})
		u.ID = nil
		u.CreatedAt = nil
		utils.RemoveTags(&u.Upstream, selectTags)
		file.Upstreams = append(file.Upstreams, u)
	}
	sort.SliceStable(file.Upstreams, func(i, j int) bool {
		return strings.Compare(*file.Upstreams[i].Name,
			*file.Upstreams[j].Name) < 0
	})

	certificates, err := kongState.Certificates.GetAll()
	if err != nil {
		return err
	}
	for _, c := range certificates {
		c := Certificate{Certificate: c.Certificate}
		sort.SliceStable(c.SNIs, func(i, j int) bool {
			return strings.Compare(*c.SNIs[i], *c.SNIs[j]) < 0
		})
		c.ID = nil
		c.CreatedAt = nil
		utils.RemoveTags(&c.Certificate, selectTags)
		file.Certificates = append(file.Certificates, c)
	}
	sort.SliceStable(file.Certificates, func(i, j int) bool {
		return strings.Compare(*file.Certificates[i].Cert,
			*file.Certificates[j].Cert) < 0
	})

	consumers, err := kongState.Consumers.GetAll()
	if err != nil {
		return err
	}
	for _, c := range consumers {
		c := Consumer{Consumer: c.Consumer}
		plugins, err := kongState.Plugins.GetAllByConsumerID(*c.ID)
		if err != nil {
			return err
		}
		for _, p := range plugins {
			p.ID = nil
			p.CreatedAt = nil
			p.Consumer = nil
			utils.RemoveTags(&p.Plugin, selectTags)
			c.Plugins = append(c.Plugins, &Plugin{Plugin: p.Plugin})
		}
		sort.SliceStable(c.Plugins, func(i, j int) bool {
			return strings.Compare(*c.Plugins[i].Name, *c.Plugins[j].Name) < 0
		})
		// custom-entities associated with Consumer
		keyAuths, err := kongState.KeyAuths.GetAllByConsumerID(*c.ID)
		if err != nil {
			return err
		}
		for _, k := range keyAuths {
			k.ID = nil
			k.CreatedAt = nil
			k.Consumer = nil
			c.KeyAuths = append(c.KeyAuths, &k.KeyAuth)
		}
		hmacAuth, err := kongState.HMACAuths.GetAllByConsumerID(*c.ID)
		if err != nil {
			return err
		}
		for _, k := range hmacAuth {
			k.ID = nil
			k.CreatedAt = nil
			k.Consumer = nil
			c.HMACAuths = append(c.HMACAuths, &k.HMACAuth)
		}
		jwtSecrets, err := kongState.JWTAuths.GetAllByConsumerID(*c.ID)
		if err != nil {
			return err
		}
		for _, k := range jwtSecrets {
			k.ID = nil
			k.CreatedAt = nil
			k.Consumer = nil
			c.JWTAuths = append(c.JWTAuths, &k.JWTAuth)
		}
		basicAuths, err := kongState.BasicAuths.GetAllByConsumerID(*c.ID)
		if err != nil {
			return err
		}
		for _, k := range basicAuths {
			k.ID = nil
			k.CreatedAt = nil
			k.Consumer = nil
			c.BasicAuths = append(c.BasicAuths, &k.BasicAuth)
		}
		c.ID = nil
		c.CreatedAt = nil
		utils.RemoveTags(&c.Consumer, selectTags)
		file.Consumers = append(file.Consumers, c)
	}
	sort.SliceStable(file.Consumers, func(i, j int) bool {
		return strings.Compare(*file.Consumers[i].Username,
			*file.Consumers[j].Username) < 0
	})
	file.Info.SelectorTags = selectTags
	// hardcoded as only one version exists currently
	file.FormatVersion = "1.1"

	c, err := yaml.Marshal(file)
	if err != nil {
		return err
	}

	if filename == "-" {
		_, err = fmt.Print(string(c))
	} else {
		err = ioutil.WriteFile(filename, c, 0600)
	}
	if err != nil {
		return err
	}
	return nil
}
