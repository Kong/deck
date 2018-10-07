package dump

import (
	"errors"

	"github.com/hbagdi/doko/utils"
	"github.com/hbagdi/go-kong/kong"
)

// Get gets all entities in Kong
func Get(client *kong.Client) (*KongRawState, error) {

	var state KongRawState
	services, err := GetAllServices(client)
	if err != nil {
		return nil, err
	}
	for _, s := range state.Services {
		if utils.Empty(s.Name) {
			return nil, errors.New("service with id '" + *s.ID + "' has no 'name' property." +
				" 'name' property is required if IDs are not being exported.")
		}
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

func GetAllTargets(client *kong.Client, upstreams []*kong.Upstream) ([]*kong.Target, error) {
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
