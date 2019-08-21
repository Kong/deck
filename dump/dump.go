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

	// SelectorTags can be used to export entities tagged with only specific
	// tags.
	SelectorTags []string
}

// GetState queries Kong for all entities using client and
// constructs a structured state.
func GetState(client *kong.Client, config Config) (*state.KongState, error) {
	raw, err := Get(client, config)
	if err != nil {
		return nil, err
	}
	kongState, err := state.NewKongState()
	if err != nil {
		return nil, errors.Wrap(err, "creating new in-memory state of Kong")
	}
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
		for _, cred := range raw.KeyAuths {
			consumer, err := kongState.Consumers.Get(*cred.Consumer.ID)
			if err != nil {
				// key could belong to a consumer which is not part
				// of this sub-set of the entire data-base
				if err == state.ErrNotFound && len(config.SelectorTags) > 0 {
					continue
				}
				return nil, errors.Wrapf(err,
					"looking up consumer '%v' for key-auth '%v'",
					*cred.Consumer.ID, *cred.ID)
			}
			cred.Consumer = consumer.DeepCopy()
			err = kongState.KeyAuths.Add(state.KeyAuth{KeyAuth: *cred})
			if err != nil {
				return nil, errors.Wrap(err, "inserting key-auth into state")
			}
		}
		for _, cred := range raw.HMACAuths {
			consumer, err := kongState.Consumers.Get(*cred.Consumer.ID)
			if err != nil {
				// key could belong to a consumer which is not part
				// of this sub-set of the entire data-base
				if err == state.ErrNotFound && len(config.SelectorTags) > 0 {
					continue
				}
				return nil, errors.Wrapf(err,
					"looking up consumer '%v' for hmac-auth '%v'",
					*cred.Consumer.ID, *cred.ID)
			}
			cred.Consumer = consumer.DeepCopy()
			err = kongState.HMACAuths.Add(state.HMACAuth{HMACAuth: *cred})
			if err != nil {
				return nil, errors.Wrap(err, "inserting hmac-auth into state")
			}
		}
		for _, cred := range raw.JWTAuths {
			consumer, err := kongState.Consumers.Get(*cred.Consumer.ID)
			if err != nil {
				// key could belong to a consumer which is not part
				// of this sub-set of the entire data-base
				if err == state.ErrNotFound && len(config.SelectorTags) > 0 {
					continue
				}
				return nil, errors.Wrapf(err,
					"looking up consumer '%v' for jwt '%v'",
					*cred.Consumer.ID, *cred.ID)
			}
			cred.Consumer = consumer.DeepCopy()
			err = kongState.JWTAuths.Add(state.JWTAuth{JWTAuth: *cred})
			if err != nil {
				return nil, errors.Wrap(err, "inserting jwt into state")
			}
		}
		for _, cred := range raw.BasicAuths {
			consumer, err := kongState.Consumers.Get(*cred.Consumer.ID)
			if err != nil {
				// key could belong to a consumer which is not part
				// of this sub-set of the entire data-base
				if err == state.ErrNotFound && len(config.SelectorTags) > 0 {
					continue
				}
				return nil, errors.Wrapf(err,
					"looking up consumer '%v' for basic-auth '%v'",
					*cred.Consumer.ID, *cred.ID)
			}
			cred.Consumer = consumer.DeepCopy()
			err = kongState.BasicAuths.Add(state.BasicAuth{BasicAuth: *cred})
			if err != nil {
				return nil, errors.Wrap(err, "inserting basic-auth into state")
			}
		}
		for _, cred := range raw.Oauth2Creds {
			consumer, err := kongState.Consumers.Get(*cred.Consumer.ID)
			if err != nil {
				// key could belong to a consumer which is not part
				// of this sub-set of the entire data-base
				if err == state.ErrNotFound && len(config.SelectorTags) > 0 {
					continue
				}
				return nil, errors.Wrapf(err,
					"looking up consumer '%v' for oauth2 '%v'",
					*cred.Consumer.ID, *cred.ID)
			}
			cred.Consumer = consumer.DeepCopy()
			err = kongState.Oauth2Creds.Add(state.Oauth2Credential{Oauth2Credential: *cred})
			if err != nil {
				return nil, errors.Wrap(err, "inserting oauth2-cred into state")
			}
		}
		for _, cred := range raw.ACLGroups {
			consumer, err := kongState.Consumers.Get(*cred.Consumer.ID)
			if err != nil {
				// key could belong to a consumer which is not part
				// of this sub-set of the entire data-base
				if err == state.ErrNotFound && len(config.SelectorTags) > 0 {
					continue
				}
				return nil, errors.Wrapf(err,
					"looking up consumer '%v' for acl '%v'",
					*cred.Consumer.ID, *cred.Group)
			}
			cred.Consumer = consumer.DeepCopy()
			err = kongState.ACLGroups.Add(state.ACLGroup{ACLGroup: *cred})
			if err != nil {
				return nil, errors.Wrap(err, "inserting basic-auth into state")
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

	for _, c := range raw.CACertificates {
		err := kongState.CACertificates.Add(state.CACertificate{
			CACertificate: *c,
		})
		if err != nil {
			return nil, errors.Wrap(err, "inserting ca_certificate into state")
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

	var state utils.KongRawState
	services, err := GetAllServices(client, config.SelectorTags)
	if err != nil {
		return nil, err
	}

	state.Services = services

	routes, err := GetAllRoutes(client, config.SelectorTags)
	if err != nil {
		return nil, err
	}
	state.Routes = routes

	plugins, err := GetAllPlugins(client, config.SelectorTags)
	if err != nil {
		return nil, err
	}
	state.Plugins = plugins

	certificates, err := GetAllCertificates(client, config.SelectorTags)
	if err != nil {
		return nil, err
	}
	state.Certificates = certificates

	caCerts, err := GetAllCACertificates(client, config.SelectorTags)
	if err != nil {
		return nil, err
	}
	state.CACertificates = caCerts

	snis, err := GetAllSNIs(client, config.SelectorTags)
	if err != nil {
		return nil, err
	}
	state.SNIs = snis

	if !config.SkipConsumers {
		consumers, err := GetAllConsumers(client, config.SelectorTags)
		if err != nil {
			return nil, err
		}
		state.Consumers = consumers
	}

	upstreams, err := GetAllUpstreams(client, config.SelectorTags)
	if err != nil {
		return nil, err
	}
	state.Upstreams = upstreams

	targets, err := GetAllTargets(client, upstreams, config.SelectorTags)
	if err != nil {
		return nil, err
	}
	state.Targets = targets

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
		certificates = append(certificates, s...)
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
	var jwtAuths []*kong.BasicAuth
	// tags are not supported on credentials
	// opt := newOpt(tags)
	opt := newOpt(nil)

	for {
		s, nextopt, err := client.BasicAuths.List(nil, opt)
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

// GetAllOauth2Creds queries Kong for all oauth2 credentials using client.
func GetAllOauth2Creds(client *kong.Client,
	tags []string) ([]*kong.Oauth2Credential, error) {
	var oauth2Creds []*kong.Oauth2Credential
	// tags are not supported on credentials
	// opt := newOpt(tags)
	opt := newOpt(nil)

	for {
		s, nextopt, err := client.Oauth2Credentials.List(nil, opt)
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
