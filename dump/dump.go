package dump

import (
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	"github.com/hbagdi/go-kong/kong"
	"github.com/pkg/errors"
)

// Config can be used to skip exporting certain entities
type Config struct {
	// If true, consumers and any plugins associated with it
	// are not exported.
	SkipConsumers bool
}

// GetState queries Kong for all entities using client and
// constructs a structered state.
func GetState(client *kong.Client, config Config) (*state.KongState, error) {
	raw, err := Get(client, config)
	if err != nil {
		return nil, err
	}
	kongState, err := state.NewKongState()
	for _, s := range raw.Services {
		if utils.Empty(s.Name) {
			return nil, errors.New("service '" + *s.ID + "' does not" +
				" have a name. decK needs services to be named.")
		}
		err := kongState.Services.Add(state.Service{Service: *s})
		if err != nil {
			return nil, errors.Wrap(err, "inserting service into state")
		}
	}
	for _, r := range raw.Routes {
		if utils.Empty(r.Name) {
			return nil, errors.New("route '" + *r.ID + "' does not" +
				" have a name. decK needs routes to be named.")
		}
		s, err := kongState.Services.Get(*r.Service.ID)
		if err != nil {
			return nil, errors.Wrapf(err,
				"looking up service '%v' for route '%v'",
				*r.Service.ID, *r.Name)
		}
		r.Service = s.DeepCopy()
		err = kongState.Routes.Add(state.Route{Route: *r})
		if err != nil {
			return nil, errors.Wrap(err, "inserting route into state")
		}
	}
	if !config.SkipConsumers {
		for _, c := range raw.Consumers {
			if utils.Empty(c.Username) {
				return nil, errors.New("consumer '" + *c.ID + "' does not" +
					" have a username. decK needs consumers to have " +
					"If you're using custom_id, " +
					"please set the username of the consumer " +
					"same as the custom_id.")
			}
			err := kongState.Consumers.Add(state.Consumer{Consumer: *c})
			if err != nil {
				return nil, errors.Wrap(err, "inserting consumer into state")
			}
		}
	}
	for _, u := range raw.Upstreams {
		if utils.Empty(u.Name) {
			return nil, errors.New("upstream '" + *u.ID + "' does not" +
				" have a name. decK needs upstreams to be named.")
		}
		err := kongState.Upstreams.Add(state.Upstream{Upstream: *u})
		if err != nil {
			return nil, errors.Wrap(err, "inserting upstream into state")
		}
	}
	for _, t := range raw.Targets {
		u, err := kongState.Upstreams.Get(*t.Upstream.ID)
		if err != nil {
			return nil, errors.Wrapf(err,
				"looking up upstream '%v' for target '%v'",
				*t.Upstream.ID, *t.Target)
		}
		t.Upstream = u.DeepCopy()
		err = kongState.Targets.Add(state.Target{Target: *t})
		if err != nil {
			return nil, errors.Wrap(err, "inserting target into state")
		}
	}

	for _, c := range raw.Certificates {
		err := kongState.Certificates.Add(state.Certificate{Certificate: *c})
		if err != nil {
			return nil, errors.Wrap(err, "inserting certificate into state")
		}
	}

	for _, p := range raw.Plugins {
		relations := 0
		if p.Service != nil {
			relations++
		}
		if p.Route != nil {
			relations++
		}
		if p.Consumer != nil {
			relations++
		}
		if relations > 1 {
			panic("plugins on a combination of route/service/consumer " +
				"are not yet supported by decK")
		}
		if p.Service != nil {
			s, err := kongState.Services.Get(*p.Service.ID)
			if err != nil {
				return nil, errors.Wrapf(err,
					"looking up service '%v' for plugin '%v'",
					*p.Service.ID, *p.Name)
			}
			p.Service = s.DeepCopy()
		}
		if p.Route != nil {
			r, err := kongState.Routes.Get(*p.Route.ID)
			if err != nil {
				return nil, errors.Wrapf(err,
					"looking up route '%v' for plugin '%v'",
					*p.Route.ID, *p.Name)
			}
			p.Route = r.DeepCopy()
		}
		if p.Consumer != nil {
			// if consumer export is disabled, do not export
			// plugins associated with consumers as well
			if config.SkipConsumers {
				continue
			}
			c, err := kongState.Consumers.Get(*p.Consumer.ID)
			if err != nil {
				return nil, errors.Wrapf(err,
					"looking up consumer '%v' for plugin '%v'",
					*p.Consumer.ID, *p.Name)
			}
			p.Consumer = c.DeepCopy()
		}
		err := kongState.Plugins.Add(state.Plugin{Plugin: *p})
		if err != nil {
			return nil, errors.Wrap(err, "inserting plugins into state")
		}
	}
	return kongState, nil
}

// Get queries all the entities using client and returns
// all the entities in KongRawState.
func Get(client *kong.Client, config Config) (*utils.KongRawState, error) {

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

	if !config.SkipConsumers {
		consumers, err := GetAllConsumers(client)
		if err != nil {
			return nil, err
		}
		state.Consumers = consumers
	}

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
		s, nextopt, err := client.Services.List(nil, opt)
		if err != nil {
			return nil, err
		}
		services = append(services, s...)
		if nextopt == nil {
			break
		}
		opt = nextopt
	}
	return services, nil
}

// GetAllRoutes queries Kong for all the routes using client.
func GetAllRoutes(client *kong.Client) ([]*kong.Route, error) {
	var routes []*kong.Route
	opt := new(kong.ListOpt)
	opt.Size = 1000
	for {
		s, nextopt, err := client.Routes.List(nil, opt)
		if err != nil {
			return nil, err
		}
		routes = append(routes, s...)
		if nextopt == nil {
			break
		}
		opt = nextopt
	}
	return routes, nil
}

// GetAllPlugins queries Kong for all the plugins using client.
func GetAllPlugins(client *kong.Client) ([]*kong.Plugin, error) {
	var plugins []*kong.Plugin
	opt := new(kong.ListOpt)
	opt.Size = 1000
	for {
		s, nextopt, err := client.Plugins.List(nil, opt)
		if err != nil {
			return nil, err
		}
		plugins = append(plugins, s...)
		if nextopt == nil {
			break
		}
		opt = nextopt
	}
	return plugins, nil
}

// GetAllCertificates queries Kong for all the certificates using client.
func GetAllCertificates(client *kong.Client) ([]*kong.Certificate, error) {
	var certificates []*kong.Certificate
	opt := new(kong.ListOpt)
	opt.Size = 1000
	for {
		s, nextopt, err := client.Certificates.List(nil, opt)
		if err != nil {
			return nil, err
		}
		certificates = append(certificates, s...)
		if nextopt == nil {
			break
		}
		opt = nextopt
	}
	return certificates, nil
}

// GetAllSNIs queries Kong for all the SNIs using client.
func GetAllSNIs(client *kong.Client) ([]*kong.SNI, error) {
	var snis []*kong.SNI
	opt := new(kong.ListOpt)
	opt.Size = 1000
	for {
		s, nextopt, err := client.SNIs.List(nil, opt)
		if err != nil {
			return nil, err
		}
		snis = append(snis, s...)
		if nextopt == nil {
			break
		}
		opt = nextopt
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
		s, nextopt, err := client.Consumers.List(nil, opt)
		if err != nil {
			return nil, err
		}
		consumers = append(consumers, s...)
		if nextopt == nil {
			break
		}
		opt = nextopt
	}
	return consumers, nil
}

// GetAllUpstreams queries Kong for all the Upstreams using client.
func GetAllUpstreams(client *kong.Client) ([]*kong.Upstream, error) {
	var upstreams []*kong.Upstream
	opt := new(kong.ListOpt)
	opt.Size = 1000
	for {
		s, nextopt, err := client.Upstreams.List(nil, opt)
		if err != nil {
			return nil, err
		}
		upstreams = append(upstreams, s...)
		if nextopt == nil {
			break
		}
		opt = nextopt
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
			t, nextopt, err := client.Targets.List(nil, upstream.ID, opt)
			if err != nil {
				return nil, err
			}
			targets = append(targets, t...)
			if nextopt == nil {
				break
			}
			opt = nextopt
		}
	}

	return targets, nil
}
