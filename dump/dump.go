package dump

import (
	"context"
	"fmt"

	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// Config can be used to skip exporting certain entities
type Config struct {
	// If true, only RBAC resources are exported.
	// SkipConsumers and SelectorTags should be falsy when this is set.
	RBACResourcesOnly bool

	// If true, consumers and any plugins associated with it
	// are not exported.
	SkipConsumers bool

	// SelectorTags can be used to export entities tagged with only specific
	// tags.
	SelectorTags []string
}

func deduplicate(stringSlice []string) []string {
	existing := map[string]struct{}{}
	result := []string{}

	for _, s := range stringSlice {
		if _, exists := existing[s]; !exists {
			existing[s] = struct{}{}
			result = append(result, s)
		}
	}

	return result
}

func newOpt(tags []string) *kong.ListOpt {
	opt := new(kong.ListOpt)
	opt.Size = 1000
	opt.Tags = kong.StringSlice(deduplicate(tags)...)
	opt.MatchAllTags = true
	return opt
}

func validateConfig(config Config) error {
	if config.RBACResourcesOnly {
		if config.SkipConsumers {
			return fmt.Errorf("dump: config: SkipConsumer cannot be set when RBACResourcesOnly is set")
		}
		if len(config.SelectorTags) != 0 {
			return fmt.Errorf("dump: config: SelectorTags cannot be set when RBACResourcesOnly is set")
		}
	}
	return nil
}

func getConsumerConfiguration(ctx context.Context, group *errgroup.Group,
	client *kong.Client, config Config, state *utils.KongRawState) {
	group.Go(func() error {
		consumers, err := GetAllConsumers(ctx, client, config.SelectorTags)
		if err != nil {
			return errors.Wrap(err, "consumers")
		}
		state.Consumers = consumers
		return nil
	})

	group.Go(func() error {
		keyAuths, err := GetAllKeyAuths(ctx, client, config.SelectorTags)
		if err != nil {
			return errors.Wrap(err, "key-auths")
		}
		state.KeyAuths = keyAuths
		return nil
	})

	group.Go(func() error {
		hmacAuths, err := GetAllHMACAuths(ctx, client, config.SelectorTags)
		if err != nil {
			return errors.Wrap(err, "hmac-auths")
		}
		state.HMACAuths = hmacAuths
		return nil
	})

	group.Go(func() error {
		jwtAuths, err := GetAllJWTAuths(ctx, client, config.SelectorTags)
		if err != nil {
			return errors.Wrap(err, "jwts")
		}
		state.JWTAuths = jwtAuths
		return nil
	})

	group.Go(func() error {
		basicAuths, err := GetAllBasicAuths(ctx, client, config.SelectorTags)
		if err != nil {
			return errors.Wrap(err, "basic-auths")
		}
		state.BasicAuths = basicAuths
		return nil
	})

	group.Go(func() error {
		oauth2Creds, err := GetAllOauth2Creds(ctx, client, config.SelectorTags)
		if err != nil {
			return errors.Wrap(err, "oauth2")
		}
		state.Oauth2Creds = oauth2Creds
		return nil
	})

	group.Go(func() error {
		aclGroups, err := GetAllACLGroups(ctx, client, config.SelectorTags)
		if err != nil {
			return errors.Wrap(err, "acls")
		}
		state.ACLGroups = aclGroups
		return nil
	})

	group.Go(func() error {
		mtlsAuths, err := GetAllMTLSAuths(ctx, client, config.SelectorTags)
		if err != nil {
			return errors.Wrap(err, "mtls-auths")
		}
		state.MTLSAuths = mtlsAuths
		return nil
	})
}

func getProxyConfiguration(ctx context.Context, group *errgroup.Group,
	client *kong.Client, config Config, state *utils.KongRawState) {
	group.Go(func() error {
		services, err := GetAllServices(ctx, client, config.SelectorTags)
		if err != nil {
			return errors.Wrap(err, "services")
		}
		state.Services = services
		return nil
	})

	group.Go(func() error {
		routes, err := GetAllRoutes(ctx, client, config.SelectorTags)
		if err != nil {
			return errors.Wrap(err, "routes")
		}
		state.Routes = routes
		return nil
	})

	group.Go(func() error {
		plugins, err := GetAllPlugins(ctx, client, config.SelectorTags)
		if err != nil {
			return errors.Wrap(err, "plugins")
		}
		if config.SkipConsumers {
			state.Plugins = excludeConsumersPlugins(plugins)
		} else {
			state.Plugins = plugins
		}
		return nil
	})

	group.Go(func() error {
		certificates, err := GetAllCertificates(ctx, client, config.SelectorTags)
		if err != nil {
			return errors.Wrap(err, "certificates")
		}
		state.Certificates = certificates
		return nil
	})

	group.Go(func() error {
		caCerts, err := GetAllCACertificates(ctx, client, config.SelectorTags)
		if err != nil {
			return errors.Wrap(err, "ca-certificates")
		}
		state.CACertificates = caCerts
		return nil
	})

	group.Go(func() error {
		snis, err := GetAllSNIs(ctx, client, config.SelectorTags)
		if err != nil {
			return errors.Wrap(err, "snis")
		}
		state.SNIs = snis
		return nil
	})

	group.Go(func() error {
		upstreams, err := GetAllUpstreams(ctx, client, config.SelectorTags)
		if err != nil {
			return errors.Wrap(err, "upstreams")
		}
		state.Upstreams = upstreams
		targets, err := GetAllTargets(ctx, client, upstreams, config.SelectorTags)
		if err != nil {
			return errors.Wrap(err, "targets")
		}
		state.Targets = targets
		return nil
	})
}

func getEnterpriseRBACConfiguration(ctx context.Context, group *errgroup.Group,
	client *kong.Client, state *utils.KongRawState) {
	group.Go(func() error {
		roles, err := GetAllRBACRoles(ctx, client)
		if err != nil {
			return errors.Wrap(err, "roles")
		}
		state.RBACRoles = roles
		return nil
	})

	group.Go(func() error {
		eps, err := GetAllRBACREndpointPermissions(ctx, client)
		if err != nil {
			return errors.Wrap(err, "eps")
		}
		state.RBACEndpointPermissions = eps
		return nil
	})
}

// Get queries all the entities using client and returns
// all the entities in KongRawState.
func Get(client *kong.Client, config Config) (*utils.KongRawState, error) {

	var state utils.KongRawState

	if err := validateConfig(config); err != nil {
		return nil, err
	}

	group, ctx := errgroup.WithContext(context.Background())

	// dump only rbac resources
	if config.RBACResourcesOnly {
		getEnterpriseRBACConfiguration(ctx, group, client, &state)
	} else {
		// regular case
		getProxyConfiguration(ctx, group, client, config, &state)
		if !config.SkipConsumers {
			getConsumerConfiguration(ctx, group, client, config, &state)
		}
	}

	err := group.Wait()
	if err != nil {
		return nil, err
	}

	return &state, nil
}

// GetAllServices queries Kong for all the services using client.
func GetAllServices(ctx context.Context, client *kong.Client,
	tags []string) ([]*kong.Service, error) {
	var services []*kong.Service
	opt := newOpt(tags)

	for {
		s, nextopt, err := client.Services.List(ctx, opt)
		if err != nil {
			return nil, err
		}
		if err := ctx.Err(); err != nil {
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
func GetAllRoutes(ctx context.Context, client *kong.Client,
	tags []string) ([]*kong.Route, error) {
	var routes []*kong.Route
	opt := newOpt(tags)

	for {
		s, nextopt, err := client.Routes.List(ctx, opt)
		if err != nil {
			return nil, err
		}
		if err := ctx.Err(); err != nil {
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
func GetAllPlugins(ctx context.Context,
	client *kong.Client, tags []string) ([]*kong.Plugin, error) {
	var plugins []*kong.Plugin
	opt := newOpt(tags)

	for {
		s, nextopt, err := client.Plugins.List(ctx, opt)
		if err != nil {
			return nil, err
		}
		if err := ctx.Err(); err != nil {
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
func GetAllCertificates(ctx context.Context, client *kong.Client,
	tags []string) ([]*kong.Certificate, error) {
	var certificates []*kong.Certificate
	opt := newOpt(tags)

	for {
		s, nextopt, err := client.Certificates.List(ctx, opt)
		if err != nil {
			return nil, err
		}
		if err := ctx.Err(); err != nil {
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
func GetAllCACertificates(ctx context.Context,
	client *kong.Client,
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
		if err := ctx.Err(); err != nil {
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
func GetAllSNIs(ctx context.Context,
	client *kong.Client, tags []string) ([]*kong.SNI, error) {
	var snis []*kong.SNI
	opt := newOpt(tags)

	for {
		s, nextopt, err := client.SNIs.List(ctx, opt)
		if err != nil {
			return nil, err
		}
		if err := ctx.Err(); err != nil {
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
func GetAllConsumers(ctx context.Context,
	client *kong.Client, tags []string) ([]*kong.Consumer, error) {
	var consumers []*kong.Consumer
	opt := newOpt(tags)

	for {
		s, nextopt, err := client.Consumers.List(ctx, opt)
		if err != nil {
			return nil, err
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if err := ctx.Err(); err != nil {
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
func GetAllUpstreams(ctx context.Context,
	client *kong.Client, tags []string) ([]*kong.Upstream, error) {
	var upstreams []*kong.Upstream
	opt := newOpt(tags)

	for {
		s, nextopt, err := client.Upstreams.List(ctx, opt)
		if err != nil {
			return nil, err
		}
		if err := ctx.Err(); err != nil {
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
func GetAllTargets(ctx context.Context, client *kong.Client,
	upstreams []*kong.Upstream, tags []string) ([]*kong.Target, error) {
	var targets []*kong.Target
	opt := newOpt(tags)

	for _, upstream := range upstreams {
		for {
			t, nextopt, err := client.Targets.List(ctx, upstream.ID, opt)
			if err != nil {
				return nil, err
			}
			if err := ctx.Err(); err != nil {
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
func GetAllKeyAuths(ctx context.Context,
	client *kong.Client, tags []string) ([]*kong.KeyAuth, error) {
	var keyAuths []*kong.KeyAuth
	// tags are not supported on credentials
	// opt := newOpt(tags)
	opt := newOpt(nil)

	for {
		s, nextopt, err := client.KeyAuths.List(ctx, opt)
		if kong.IsNotFoundErr(err) {
			return keyAuths, nil
		}
		if err != nil {
			return nil, err
		}
		if err := ctx.Err(); err != nil {
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
func GetAllHMACAuths(ctx context.Context,
	client *kong.Client, tags []string) ([]*kong.HMACAuth, error) {
	var hmacAuths []*kong.HMACAuth
	// tags are not supported on credentials
	// opt := newOpt(tags)
	opt := newOpt(nil)

	for {
		s, nextopt, err := client.HMACAuths.List(ctx, opt)
		if kong.IsNotFoundErr(err) {
			return hmacAuths, nil
		}
		if err != nil {
			return nil, err
		}
		if err := ctx.Err(); err != nil {
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
func GetAllJWTAuths(ctx context.Context,
	client *kong.Client, tags []string) ([]*kong.JWTAuth, error) {
	var jwtAuths []*kong.JWTAuth
	// tags are not supported on credentials
	// opt := newOpt(tags)
	opt := newOpt(nil)

	for {
		s, nextopt, err := client.JWTAuths.List(ctx, opt)
		if kong.IsNotFoundErr(err) {
			return jwtAuths, nil
		}
		if err != nil {
			return nil, err
		}
		if err := ctx.Err(); err != nil {
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
func GetAllBasicAuths(ctx context.Context,
	client *kong.Client, tags []string) ([]*kong.BasicAuth, error) {
	var basicAuths []*kong.BasicAuth
	// tags are not supported on credentials
	// opt := newOpt(tags)
	opt := newOpt(nil)

	for {
		s, nextopt, err := client.BasicAuths.List(ctx, opt)
		if kong.IsNotFoundErr(err) {
			return basicAuths, nil
		}
		if err != nil {
			return nil, err
		}
		if err := ctx.Err(); err != nil {
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
func GetAllOauth2Creds(ctx context.Context, client *kong.Client,
	tags []string) ([]*kong.Oauth2Credential, error) {
	var oauth2Creds []*kong.Oauth2Credential
	// tags are not supported on credentials
	// opt := newOpt(tags)
	opt := newOpt(nil)

	for {
		s, nextopt, err := client.Oauth2Credentials.List(ctx, opt)
		if kong.IsNotFoundErr(err) {
			return oauth2Creds, nil
		}
		if err != nil {
			return nil, err
		}
		if err := ctx.Err(); err != nil {
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
func GetAllACLGroups(ctx context.Context,
	client *kong.Client, tags []string) ([]*kong.ACLGroup, error) {
	var aclGroups []*kong.ACLGroup
	// tags are not supported on credentials
	// opt := newOpt(tags)
	opt := newOpt(nil)

	for {
		s, nextopt, err := client.ACLs.List(ctx, opt)
		if kong.IsNotFoundErr(err) {
			return aclGroups, nil
		}
		if err != nil {
			return nil, err
		}
		if err := ctx.Err(); err != nil {
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

// GetAllMTLSAuths queries Kong for all basic-auth credentials using client.
func GetAllMTLSAuths(ctx context.Context,
	client *kong.Client, tags []string) ([]*kong.MTLSAuth, error) {
	var mtlsAuths []*kong.MTLSAuth
	// tags are not supported on credentials
	// opt := newOpt(tags)
	opt := newOpt(nil)

	for {
		s, nextopt, err := client.MTLSAuths.List(ctx, opt)
		if kong.IsNotFoundErr(err) {
			return mtlsAuths, nil
		}
		// TODO figure out a better way to handle unauthorized endpoints
		// per https://github.com/Kong/deck/issues/274 we can't dump these resources
		// from an Enterprise instance running in free mode, and the 403 results in a
		// fatal error when running "deck dump". We don't want to just treat 403s the
		// same as 404s because Kong also uses them to indicate missing RBAC permissions,
		// but this is currently necessary for compatibility. We need a better approach
		// before adding other Enterprise resources that decK handles by default (versus,
		// for example, RBAC roles, which require the --rbac-resources-only flag).
		if err.(*kong.APIError).Code() == 403 {
			return mtlsAuths, nil
		}
		if err != nil {
			return nil, err
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		mtlsAuths = append(mtlsAuths, s...)
		if nextopt == nil {
			break
		}
		opt = nextopt
	}
	return mtlsAuths, nil
}

// GetAllRBACRoles queries Kong for all the RBACRoles using client.
func GetAllRBACRoles(ctx context.Context,
	client *kong.Client) ([]*kong.RBACRole, error) {

	roles, err := client.RBACRoles.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}

func GetAllRBACREndpointPermissions(ctx context.Context,
	client *kong.Client) ([]*kong.RBACEndpointPermission, error) {

	var eps = []*kong.RBACEndpointPermission{}
	roles, err := client.RBACRoles.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// retrieve all permissions for the role
	for _, r := range roles {
		reps, err := client.RBACEndpointPermissions.ListAllForRole(ctx, r.ID)
		if err != nil {
			return nil, err
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		eps = append(eps, reps...)
	}

	return eps, nil
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
