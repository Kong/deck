package file

import (
	"io/ioutil"
	"sort"
	"strings"

	"github.com/kong/deck/state"
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
		for _, r := range routes {
			r.Service = nil
			r.ID = nil
			r.CreatedAt = nil
			r.UpdatedAt = nil
			s.Routes = append(s.Routes, &route{Route: r.Route})
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

	c, err := yaml.Marshal(file)
	err = ioutil.WriteFile(filename, c, 0600)
	if err != nil {
		return err
	}
	return nil
}
