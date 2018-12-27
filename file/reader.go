package file

import (
	"errors"
	"io/ioutil"
	"strconv"

	"github.com/hbagdi/go-kong/kong"
	"github.com/kong/deck/counter"
	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
	yaml "gopkg.in/yaml.v2"
)

type service struct {
	kong.Service `yaml:",inline"`
	Routes       []*route
}

type route struct {
	kong.Route `yaml:",inline"`
}

type fileStructure struct {
	Services []service
}

var count counter.Counter

// GetStateFromFile reads in a file with filename and constructs
// a state. It will return an error if the file representation is invalid
// or if there is any error during processing.
// All entities without an ID will get a `placeholder-{iota}` ID
// assigned to them.
func GetStateFromFile(filename string) (*state.KongState, error) {

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
		// TODO add override logic
		// TODO add support for file based defaults
		if utils.Empty(s.ID) {
			s.ID = kong.String("placeholder-" + strconv.FormatUint(count.Inc(), 10))
		}
		// TODO add check if service is named or not
		// TODO check for duplicate services (services with same name)
		err := kongState.AddService(state.Service{Service: s.Service})
		if err != nil {
			return nil, err
		}

		for _, r := range s.Routes {
			if utils.Empty(r.ID) {
				r.ID = kong.String("placeholder-" + strconv.FormatUint(count.Inc(), 10))
			}
			// TODO add check if route is named or not
			r.Service = s.Service.DeepCopy()
			err := kongState.AddRoute(state.Route{Route: r.Route})
			if err != nil {
				return nil, err
			}
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
