package file

import (
	"io/ioutil"
	"sort"
	"strings"

	"github.com/hbagdi/deck/state"
	yaml "gopkg.in/yaml.v2"
)

// KongStateToFile writes a state object to file with filename.
// It will omit timestamps and IDs while writing.
func KongStateToFile(kongState *state.KongState, filename string) error {
	var file fileStructure

	services, err := kongState.Services.GetAll()
	if err != nil {
		return err
	}
	for _, s := range services {
		s := service{Service: s.Service}
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
			s.Plugins = append(s.Plugins, &plugin{Plugin: p.Plugin})
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
			route := &route{Route: r.Route}
			for _, p := range plugins {
				p.ID = nil
				p.CreatedAt = nil
				p.Route = nil
				route.Plugins = append(route.Plugins, &plugin{Plugin: p.Plugin})
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
			p := plugin{Plugin: p.Plugin}
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
		u := upstream{Upstream: u.Upstream}
		targets, err := kongState.Targets.GetAllByUpstreamID(*u.ID)
		if err != nil {
			return err
		}
		for _, t := range targets {
			t.Upstream = nil
			t.ID = nil
			t.CreatedAt = nil
			u.Targets = append(u.Targets, &target{Target: t.Target})
		}
		sort.SliceStable(u.Targets, func(i, j int) bool {
			return strings.Compare(*u.Targets[i].Target.Target,
				*u.Targets[j].Target.Target) < 0
		})
		u.ID = nil
		u.CreatedAt = nil
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
		c := certificate{Certificate: c.Certificate}
		sort.SliceStable(c.SNIs, func(i, j int) bool {
			return strings.Compare(*c.SNIs[i], *c.SNIs[j]) < 0
		})
		c.ID = nil
		c.CreatedAt = nil
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
		c := consumer{Consumer: c.Consumer}
		plugins, err := kongState.Plugins.GetAllByConsumerID(*c.ID)
		if err != nil {
			return err
		}
		for _, p := range plugins {
			p.ID = nil
			p.CreatedAt = nil
			p.Consumer = nil
			c.Plugins = append(c.Plugins, &plugin{Plugin: p.Plugin})
		}
		sort.SliceStable(c.Plugins, func(i, j int) bool {
			return strings.Compare(*c.Plugins[i].Name, *c.Plugins[j].Name) < 0
		})
		c.ID = nil
		c.CreatedAt = nil
		file.Consumers = append(file.Consumers, c)
	}
	sort.SliceStable(file.Consumers, func(i, j int) bool {
		return strings.Compare(*file.Consumers[i].Username,
			*file.Consumers[j].Username) < 0
	})

	c, err := yaml.Marshal(file)
	err = ioutil.WriteFile(filename, c, 0600)
	if err != nil {
		return err
	}
	return nil
}
