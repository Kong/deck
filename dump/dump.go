package dump

import (
	"github.com/hbagdi/go-kong/kong"
	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
	"github.com/pkg/errors"
)

// GetState queries Kong for all entities using client and
// constructs a structered state.
func GetState(client *kong.Client) (*state.KongState, error) {
	raw, err := Get(client)
	if err != nil {
		return nil, err
	}
	kongState, err := state.NewKongState()
	for _, s := range raw.Services {
		if utils.Empty(s.Name) {
			return nil, errors.New("service '" + *s.ID + "' does not" +
				" have a name. decK needs services to be named.")
		}
		err := kongState.AddService(state.Service{Service: *s})
		if err != nil {
			return nil, errors.Wrap(err, "inserting service into state")
		}
	}
	for _, r := range raw.Routes {
		if utils.Empty(r.Name) {
			return nil, errors.New("route '" + *r.ID + "' does not" +
				" have a name. decK needs routes to be named.")
		}
		err := kongState.AddRoute(state.Route{Route: *r})
		if err != nil {
			return nil, errors.Wrap(err, "inserting route into state")
		}
	}

	return kongState, nil
}

// Get queries all the entities using client and returns
// all the entities in KongRawState.
func Get(client *kong.Client) (*utils.KongRawState, error) {

	var state utils.KongRawState
	services, err := GetAllServices(client)
	if err != nil {
		return nil, err
	}

	state.Services = services

	routes, err := GetAllRoutes(client)
	if err != nil {
		return nil, err
	}
	state.Routes = routes

	plugins, err := GetAllPlugins(client)
	if err != nil {
		return nil, err
	}
	state.Plugins = plugins

	certificates, err := GetAllCertificates(client)
	if err != nil {
		return nil, err
	}
	state.Certificates = certificates

	snis, err := GetAllSNIs(client)
	if err != nil {
		return nil, err
	}
	state.SNIs = snis

	consumers, err := GetAllConsumers(client)
	if err != nil {
		return nil, err
	}
	state.Consumers = consumers

	upstreams, err := GetAllUpstreams(client)
	if err != nil {
		return nil, err
	}
	state.Upstreams = upstreams

	targets, err := GetAllTargets(client, upstreams)
	if err != nil {
		return nil, err
	}
	state.Targets = targets

	return &state, nil
}

// GetAllServices queries Kong for all the services using client.
func GetAllServices(client *kong.Client) ([]*kong.Service, error) {
	var services []*kong.Service
	opt := new(kong.ListOpt)
	opt.Size = 1000
	for {
		s, opt, err := client.Services.List(nil, opt)
		if err != nil {
			return nil, err
		}
		services = append(services, s...)
		if opt == nil {
			break
		}
	}
	return services, nil
}

// GetAllRoutes queries Kong for all the routes using client.
func GetAllRoutes(client *kong.Client) ([]*kong.Route, error) {
	var routes []*kong.Route
	opt := new(kong.ListOpt)
	opt.Size = 1000
	for {
		s, opt, err := client.Routes.List(nil, opt)
		if err != nil {
			return nil, err
		}
		routes = append(routes, s...)
		if opt == nil {
			break
		}
	}
	return routes, nil
}

// GetAllPlugins queries Kong for all the plugins using client.
func GetAllPlugins(client *kong.Client) ([]*kong.Plugin, error) {
	var plugins []*kong.Plugin
	opt := new(kong.ListOpt)
	opt.Size = 1000
	for {
		s, opt, err := client.Plugins.List(nil, opt)
		if err != nil {
			return nil, err
		}
		plugins = append(plugins, s...)
		if opt == nil {
			break
		}
	}
	return plugins, nil
}

// GetAllCertificates queries Kong for all the certificates using client.
func GetAllCertificates(client *kong.Client) ([]*kong.Certificate, error) {
	var certificates []*kong.Certificate
	opt := new(kong.ListOpt)
	opt.Size = 1000
	for {
		s, opt, err := client.Certificates.List(nil, opt)
		if err != nil {
			return nil, err
		}
		certificates = append(certificates, s...)
		if opt == nil {
			break
		}
	}
	return certificates, nil
}

// GetAllSNIs queries Kong for all the SNIs using client.
func GetAllSNIs(client *kong.Client) ([]*kong.SNI, error) {
	var snis []*kong.SNI
	opt := new(kong.ListOpt)
	opt.Size = 1000
	for {
		s, opt, err := client.SNIs.List(nil, opt)
		if err != nil {
			return nil, err
		}
		snis = append(snis, s...)
		if opt == nil {
			break
		}
	}
	return snis, nil
}

// GetAllConsumers queries Kong for all the consumers using client.
// Please use this method with caution if you have a lot of consumers.
func GetAllConsumers(client *kong.Client) ([]*kong.Consumer, error) {
	var consumers []*kong.Consumer
	opt := new(kong.ListOpt)
	opt.Size = 1000
	for {
		s, opt, err := client.Consumers.List(nil, opt)
		if err != nil {
			return nil, err
		}
		consumers = append(consumers, s...)
		if opt == nil {
			break
		}
	}
	return consumers, nil
}

// GetAllUpstreams queries Kong for all the Upstreams using client.
func GetAllUpstreams(client *kong.Client) ([]*kong.Upstream, error) {
	var upstreams []*kong.Upstream
	opt := new(kong.ListOpt)
	opt.Size = 1000
	for {
		s, opt, err := client.Upstreams.List(nil, opt)
		if err != nil {
			return nil, err
		}
		upstreams = append(upstreams, s...)
		if opt == nil {
			break
		}
	}
	return upstreams, nil
}

// GetAllTargets queries Kong for all the Targets of upstreams using client.
// Targets are queries per upstream as there exists no endpoint in Kong
// to list all targets of all upstreams.
func GetAllTargets(client *kong.Client,
	upstreams []*kong.Upstream) ([]*kong.Target, error) {
	var targets []*kong.Target
	opt := new(kong.ListOpt)
	opt.Size = 1000
	for _, upstream := range upstreams {
		for {
			t, opt, err := client.Targets.List(nil, upstream.ID, opt)
			if err != nil {
				return nil, err
			}
			targets = append(targets, t...)
			if opt == nil {
				break
			}
		}
	}

	return targets, nil
}
