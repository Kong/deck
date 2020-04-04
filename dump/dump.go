package dump

import (
	"github.com/hbagdi/deck/utils"
	"github.com/hbagdi/go-kong/kong"
	"github.com/pkg/errors"
)

// Config can be used to skip exporting certain entities
type Config struct {
	// If true, consumers and any plugins associated with it
	// are not exported.
	SkipConsumers bool

	// SelectorTags can be used to export entities tagged with only specific
	// tags.
	SelectorTags []string
}

func newOpt(tags []string) *kong.ListOpt {
	opt := new(kong.ListOpt)
	opt.Size = 1000
	opt.Tags = kong.StringSlice(tags...)
	opt.MatchAllTags = true
	return opt
}

// Get queries all the entities using client and returns
// all the entities in KongRawState.
func Get(client *kong.Client, config Config) (*utils.KongRawState, error) {

	// TODO make these requests concurrent

	var state utils.KongRawState
	services, err := GetAllServices(client, config.SelectorTags)
	if err != nil {
		return nil, errors.Wrap(err, "services")
	}

	state.Services = services

	routes, err := GetAllRoutes(client, config.SelectorTags)
	if err != nil {
		return nil, errors.Wrap(err, "routes")
	}
	state.Routes = routes

	plugins, err := GetAllPlugins(client, config.SelectorTags)
	if err != nil {
		return nil, errors.Wrap(err, "plugins")
	}
	if config.SkipConsumers {
		state.Plugins = excludeConsumersPlugins(plugins)
	} else {
		state.Plugins = plugins
	}

	certificates, err := GetAllCertificates(client, config.SelectorTags)
	if err != nil {
		return nil, errors.Wrap(err, "certificates")
	}
	state.Certificates = certificates

	caCerts, err := GetAllCACertificates(client, config.SelectorTags)
	if err != nil {
		return nil, errors.Wrap(err, "ca-certificates")
	}
	state.CACertificates = caCerts

	snis, err := GetAllSNIs(client, config.SelectorTags)
	if err != nil {
		return nil, errors.Wrap(err, "snis")
	}
	state.SNIs = snis

	upstreams, err := GetAllUpstreams(client, config.SelectorTags)
	if err != nil {
		return nil, errors.Wrap(err, "upstreams")
	}
	state.Upstreams = upstreams

	targets, err := GetAllTargets(client, upstreams, config.SelectorTags)
	if err != nil {
		return nil, errors.Wrap(err, "targets")
	}
	state.Targets = targets

	if !config.SkipConsumers {
		consumers, err := GetAllConsumers(client, config.SelectorTags)
		if err != nil {
			return nil, err
		}
		state.Consumers = consumers

		keyAuths, err := GetAllKeyAuths(client, config.SelectorTags)
		if err != nil {
			return nil, err
		}
		state.KeyAuths = keyAuths

		hmacAuths, err := GetAllHMACAuths(client, config.SelectorTags)
		if err != nil {
			return nil, err
		}
		state.HMACAuths = hmacAuths

		jwtAuths, err := GetAllJWTAuths(client, config.SelectorTags)
		if err != nil {
			return nil, err
		}
		state.JWTAuths = jwtAuths

		basicAuths, err := GetAllBasicAuths(client, config.SelectorTags)
		if err != nil {
			return nil, err
		}
		state.BasicAuths = basicAuths

		oauth2Creds, err := GetAllOauth2Creds(client, config.SelectorTags)
		if err != nil {
			return nil, err
		}
		state.Oauth2Creds = oauth2Creds

		aclGroups, err := GetAllACLGroups(client, config.SelectorTags)
		if err != nil {
			return nil, err
		}
		state.ACLGroups = aclGroups
	}

	return &state, nil
}

// GetAllServices queries Kong for all the services using client.
func GetAllServices(client *kong.Client, tags []string) ([]*kong.Service, error) {
	var services []*kong.Service
	opt := newOpt(tags)

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
func GetAllRoutes(client *kong.Client, tags []string) ([]*kong.Route, error) {
	var routes []*kong.Route
	opt := newOpt(tags)

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
func GetAllPlugins(client *kong.Client, tags []string) ([]*kong.Plugin, error) {
	var plugins []*kong.Plugin
	opt := newOpt(tags)

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
func GetAllCertificates(client *kong.Client, tags []string) ([]*kong.Certificate, error) {
	var certificates []*kong.Certificate
	opt := newOpt(tags)

	for {
		s, nextopt, err := client.Certificates.List(nil, opt)
		if err != nil {
			return nil, err
		}
		for _, cert := range s {
			c := cert
			c.SNIs = nil
			certificates = append(certificates, cert)
		}
		if nextopt == nil {
			break
		}
		opt = nextopt
	}
	return certificates, nil
}

// GetAllCACertificates queries Kong for all the CACertificates using client.
func GetAllCACertificates(client *kong.Client,
	tags []string) ([]*kong.CACertificate, error) {
	var caCertificates []*kong.CACertificate
	opt := newOpt(tags)

	for {
		s, nextopt, err := client.CACertificates.List(nil, opt)
		// Compatibility for Kong < 1.3
		// This core entitiy was not present in the past
		// and the Admin API request will error with 404 Not Found
		// If we do get the error, we return back an empty array of
		// CACertificates, effectively disabling the entity for versions
		// which don't have it.
		// A better solution would be to have a version check, and based
		// on the version, the entities are loaded and synced.
		if err != nil {
			if kong.IsNotFoundErr(err) {
				return caCertificates, nil
			}
			return nil, err
		}
		caCertificates = append(caCertificates, s...)
		if nextopt == nil {
			break
		}
		opt = nextopt
	}
	return caCertificates, nil
}

// GetAllSNIs queries Kong for all the SNIs using client.
func GetAllSNIs(client *kong.Client, tags []string) ([]*kong.SNI, error) {
	var snis []*kong.SNI
	opt := newOpt(tags)

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
func GetAllConsumers(client *kong.Client, tags []string) ([]*kong.Consumer, error) {
	var consumers []*kong.Consumer
	opt := newOpt(tags)

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
func GetAllUpstreams(client *kong.Client, tags []string) ([]*kong.Upstream, error) {
	var upstreams []*kong.Upstream
	opt := newOpt(tags)

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
	upstreams []*kong.Upstream, tags []string) ([]*kong.Target, error) {
	var targets []*kong.Target
	opt := newOpt(tags)

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

// GetAllKeyAuths queries Kong for all key-auth credentials using client.
func GetAllKeyAuths(client *kong.Client, tags []string) ([]*kong.KeyAuth, error) {
	var keyAuths []*kong.KeyAuth
	// tags are not supported on credentials
	// opt := newOpt(tags)
	opt := newOpt(nil)

	for {
		s, nextopt, err := client.KeyAuths.List(nil, opt)
		if kong.IsNotFoundErr(err) {
			return keyAuths, nil
		}
		if err != nil {
			return nil, err
		}
		keyAuths = append(keyAuths, s...)
		if nextopt == nil {
			break
		}
		opt = nextopt
	}
	return keyAuths, nil
}

// GetAllHMACAuths queries Kong for all hmac-auth credentials using client.
func GetAllHMACAuths(client *kong.Client, tags []string) ([]*kong.HMACAuth, error) {
	var hmacAuths []*kong.HMACAuth
	// tags are not supported on credentials
	// opt := newOpt(tags)
	opt := newOpt(nil)

	for {
		s, nextopt, err := client.HMACAuths.List(nil, opt)
		if kong.IsNotFoundErr(err) {
			return hmacAuths, nil
		}
		if err != nil {
			return nil, err
		}
		hmacAuths = append(hmacAuths, s...)
		if nextopt == nil {
			break
		}
		opt = nextopt
	}
	return hmacAuths, nil
}

// GetAllJWTAuths queries Kong for all jwt credentials using client.
func GetAllJWTAuths(client *kong.Client, tags []string) ([]*kong.JWTAuth, error) {
	var jwtAuths []*kong.JWTAuth
	// tags are not supported on credentials
	// opt := newOpt(tags)
	opt := newOpt(nil)

	for {
		s, nextopt, err := client.JWTAuths.List(nil, opt)
		if kong.IsNotFoundErr(err) {
			return jwtAuths, nil
		}
		if err != nil {
			return nil, err
		}
		jwtAuths = append(jwtAuths, s...)
		if nextopt == nil {
			break
		}
		opt = nextopt
	}
	return jwtAuths, nil
}

// GetAllBasicAuths queries Kong for all basic-auth credentials using client.
func GetAllBasicAuths(client *kong.Client, tags []string) ([]*kong.BasicAuth, error) {
	var basicAuths []*kong.BasicAuth
	// tags are not supported on credentials
	// opt := newOpt(tags)
	opt := newOpt(nil)

	for {
		s, nextopt, err := client.BasicAuths.List(nil, opt)
		if kong.IsNotFoundErr(err) {
			return basicAuths, nil
		}
		if err != nil {
			return nil, err
		}
		basicAuths = append(basicAuths, s...)
		if nextopt == nil {
			break
		}
		opt = nextopt
	}
	return basicAuths, nil
}

// GetAllOauth2Creds queries Kong for all oauth2 credentials using client.
func GetAllOauth2Creds(client *kong.Client,
	tags []string) ([]*kong.Oauth2Credential, error) {
	var oauth2Creds []*kong.Oauth2Credential
	// tags are not supported on credentials
	// opt := newOpt(tags)
	opt := newOpt(nil)

	for {
		s, nextopt, err := client.Oauth2Credentials.List(nil, opt)
		if kong.IsNotFoundErr(err) {
			return oauth2Creds, nil
		}
		if err != nil {
			return nil, err
		}
		oauth2Creds = append(oauth2Creds, s...)
		if nextopt == nil {
			break
		}
		opt = nextopt
	}
	return oauth2Creds, nil
}

// GetAllACLGroups queries Kong for all ACL groups using client.
func GetAllACLGroups(client *kong.Client, tags []string) ([]*kong.ACLGroup, error) {
	var aclGroups []*kong.ACLGroup
	// tags are not supported on credentials
	// opt := newOpt(tags)
	opt := newOpt(nil)

	for {
		s, nextopt, err := client.ACLs.List(nil, opt)
		if kong.IsNotFoundErr(err) {
			return aclGroups, nil
		}
		if err != nil {
			return nil, err
		}
		aclGroups = append(aclGroups, s...)
		if nextopt == nil {
			break
		}
		opt = nextopt
	}
	return aclGroups, nil
}

// excludeConsumersPlugins filter out consumer plugins
func excludeConsumersPlugins(plugins []*kong.Plugin) []*kong.Plugin {
	var filtered []*kong.Plugin
	for _, p := range plugins {
		if p.Consumer != nil && !utils.Empty(p.Consumer.ID) {
			continue
		}
		filtered = append(filtered, p)
	}
	return filtered
}
