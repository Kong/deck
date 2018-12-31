package file

import (
	"io/ioutil"
	"strconv"

	"github.com/hbagdi/go-kong/kong"
	"github.com/kong/deck/counter"
	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

var count counter.Counter

// GetStateFromFile reads in a file with filename and constructs
// a state. It will return an error if the file representation is invalid
// or if there is any error during processing.
// All entities without an ID will get a `placeholder-{iota}` ID
// assigned to them.
func GetStateFromFile(filename string) (*state.KongState, error) {
	// TODO add override logic
	// TODO add support for file based defaults
	if filename == "" {
		return nil, errors.New("filename cannot be empty")
	}
	fileContent, err := readFile(filename)
	if err != nil {
		return nil, err
	}
	kongState, err := state.NewKongState()
	if err != nil {
		return nil, err
	}
	for _, s := range fileContent.Services {
		if utils.Empty(s.ID) {
			s.ID = kong.String("placeholder-" +
				strconv.FormatUint(count.Inc(), 10))
		}
		if utils.Empty(s.Service.Name) {
			return nil, errors.New("all services in the file must be named")
		}
		_, err := kongState.Services.Get(*s.Service.Name)
		if err != state.ErrNotFound {
			return nil, errors.Errorf("duplicate service definitions"+
				" found for: '%s'", *s.Service.Name)
		}
		err = kongState.Services.Add(state.Service{Service: s.Service})
		if err != nil {
			return nil, err
		}

		for _, r := range s.Routes {
			if utils.Empty(r.ID) {
				r.ID = kong.String("placeholder-" +
					strconv.FormatUint(count.Inc(), 10))
			}
			if utils.Empty(r.Name) {
				return nil, errors.New("all routes in the file must be named")
			}
			_, err := kongState.Routes.Get(*r.Name)
			if err != state.ErrNotFound {
				return nil, errors.Errorf("duplicate route definitions"+
					" found for: '%s'", *r.Name)
			}
			// TODO add check if route is named or not
			r.Service = s.Service.DeepCopy()
			err = kongState.Routes.Add(state.Route{Route: r.Route})
			if err != nil {
				return nil, err
			}
		}
	}

	for _, u := range fileContent.Upstreams {
		if utils.Empty(u.ID) {
			u.ID = kong.String("placeholder-" +
				strconv.FormatUint(count.Inc(), 10))
		}
		if utils.Empty(u.Name) {
			return nil, errors.New("all upstreams in the file must be named")
		}
		_, err := kongState.Upstreams.Get(*u.Name)
		if err != state.ErrNotFound {
			return nil, errors.Errorf("duplicate upstream definitions"+
				" found for: '%s'", *u.Name)
		}
		err = kongState.Upstreams.Add(state.Upstream{Upstream: u.Upstream})
		if err != nil {
			return nil, err
		}

		for _, t := range u.Targets {
			if utils.Empty(t.ID) {
				t.ID = kong.String("placeholder-" +
					strconv.FormatUint(count.Inc(), 10))
			}
			_, err := kongState.Targets.Get(*t.Target.Target)
			if err != state.ErrNotFound {
				return nil, errors.Errorf("duplicate target definitions"+
					" found for: '%s'", *t.Target.Target)
			}
			t.Upstream = u.Upstream.DeepCopy()
			err = kongState.Targets.Add(state.Target{Target: t.Target})
			if err != nil {
				return nil, err
			}
		}
	}

	allSNIs := make(map[string]bool)
	for _, c := range fileContent.Certificates {
		if utils.Empty(c.ID) {
			c.ID = kong.String("placeholder-" +
				strconv.FormatUint(count.Inc(), 10))
		}
		if utils.Empty(c.Cert) || utils.Empty(c.Key) {
			return nil, errors.Errorf("all certificates must have a cert" +
				" and a key")
		}
		// check if an SNI is present in multiple certificates
		for _, s := range c.SNIs {
			if allSNIs[*s] {
				return nil, errors.Errorf("duplicate sni found: '%s'", *s)
			}
			allSNIs[*s] = true
		}

		_, err := kongState.Certificates.GetByCertKey(*c.Cert, *c.Key)
		if err != state.ErrNotFound {
			return nil, errors.Errorf("duplicate certificate definitions"+
				" found for the following certificate:\n'%s'", *c.Cert)
		}
		err = kongState.Certificates.Add(state.Certificate{
			Certificate: c.Certificate,
		})
		if err != nil {
			return nil, err
		}
	}

	return kongState, nil
}

func readFile(kongStateFile string) (*fileStructure, error) {

	var s fileStructure
	b, err := ioutil.ReadFile(kongStateFile)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(b, &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
