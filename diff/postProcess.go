package diff

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/state"
)

type servicePostAction struct{}

// Create creates the service from state.
// The first arg should be of type *state.KongState, the state in
// which the second arg, of type *state.Service, will be added.
// If the args are of incorrect types, Create will panic.
func (crud *servicePostAction) Create(arg ...crud.Arg) (crud.Arg, error) {
	s, ok := arg[0].(*state.KongState)
	if !ok {
		panic("whoops")
	}
	svc, ok := arg[1].(*state.Service)
	if !ok {
		panic("whoops")
	}
	s.Services.Add(*svc)
	return nil, nil
}

// Delete deletes the service from state.
// The first arg should be of type *state.KongState, the state from
// which the second arg, of type *state.Service, will be deleted.
// If the args are of incorrect types, Delete will panic.
func (crud *servicePostAction) Delete(arg ...crud.Arg) (crud.Arg, error) {
	s, ok := arg[0].(*state.KongState)
	if !ok {
		panic("whoops")
	}
	svc, ok := arg[1].(*state.Service)
	if !ok {
		panic("whoops")
	}
	s.Services.Delete(*svc.ID)
	return nil, nil
}

// Update updates the service from state.
// The first arg should be of type *state.KongState, the state in
// which the second arg, of type *state.Service, will be updated.
// If the args are of incorrect types, Update will panic.
func (crud *servicePostAction) Update(arg ...crud.Arg) (crud.Arg, error) {
	s, ok := arg[0].(*state.KongState)
	if !ok {
		panic("whoops")
	}
	svc, ok := arg[1].(*state.Service)
	if !ok {
		panic("whoops")
	}
	s.Services.Update(*svc)
	return nil, nil
}

type routePostAction struct{}

// Create creates the route from state.
// The first arg should be of type *state.KongState, the state in
// which the second arg, of type *state.Route, will be added.
// If the args are of incorrect types, Create will panic.
func (crud *routePostAction) Create(arg ...crud.Arg) (crud.Arg, error) {
	s, ok := arg[0].(*state.KongState)
	if !ok {
		panic("whoops")
	}
	svc, ok := arg[1].(*state.Route)
	if !ok {
		panic("whoops")
	}
	s.Routes.Add(*svc)
	return nil, nil
}

// Delete deletes the route from state.
// The first arg should be of type *state.KongState, the state from
// which the second arg, of type *state.Route, will be deleted.
// If the args are of incorrect types, Delete will panic.
func (crud *routePostAction) Delete(arg ...crud.Arg) (crud.Arg, error) {
	s, ok := arg[0].(*state.KongState)
	if !ok {
		panic("whoops")
	}
	svc, ok := arg[1].(*state.Route)
	if !ok {
		panic("whoops")
	}
	s.Routes.Delete(*svc.ID)
	return nil, nil
}

// Update updates the route from state.
// The first arg should be of type *state.KongState, the state in
// which the second arg, of type *state.Route, will be updated.
// If the args are of incorrect types, Update will panic.
func (crud *routePostAction) Update(arg ...crud.Arg) (crud.Arg, error) {
	s, ok := arg[0].(*state.KongState)
	if !ok {
		panic("whoops")
	}
	svc, ok := arg[1].(*state.Route)
	if !ok {
		panic("whoops")
	}
	s.Routes.Update(*svc)
	return nil, nil
}

type upstreamPostAction struct{}

// Create creates the Upstream in state.
// The first arg should be of type *state.KongState, the state in
// which the second arg, of type *state.Upstream, will be added.
// If the args are of incorrect types, Create will panic.
func (crud *upstreamPostAction) Create(arg ...crud.Arg) (crud.Arg, error) {
	s, ok := arg[0].(*state.KongState)
	if !ok {
		panic("whoops")
	}
	svc, ok := arg[1].(*state.Upstream)
	if !ok {
		panic("whoops")
	}
	s.Upstreams.Add(*svc)
	return nil, nil
}

// Delete deletes the Upstream from state.
// The first arg should be of type *state.KongState, the state from
// which the second arg, of type *state.Upstream, will be deleted.
// If the args are of incorrect types, Delete will panic.
func (crud *upstreamPostAction) Delete(arg ...crud.Arg) (crud.Arg, error) {
	s, ok := arg[0].(*state.KongState)
	if !ok {
		panic("whoops")
	}
	svc, ok := arg[1].(*state.Upstream)
	if !ok {
		panic("whoops")
	}
	s.Upstreams.Delete(*svc.ID)
	return nil, nil
}

// Update updates the upstream in state.
// The first arg should be of type *state.KongState, the state in
// which the second arg, of type *state.Upstream, will be updated.
// If the args are of incorrect types, Update will panic.
func (crud *upstreamPostAction) Update(arg ...crud.Arg) (crud.Arg, error) {
	s, ok := arg[0].(*state.KongState)
	if !ok {
		panic("whoops")
	}
	svc, ok := arg[1].(*state.Upstream)
	if !ok {
		panic("whoops")
	}
	s.Upstreams.Update(*svc)
	return nil, nil
}

type targetPostAction struct{}

// Create creates the target in state.
// The first arg should be of type *state.KongState, the state in
// which the second arg, of type *state.Target, will be added.
// If the args are of incorrect types, Create will panic.
func (crud *targetPostAction) Create(arg ...crud.Arg) (crud.Arg, error) {
	s, ok := arg[0].(*state.KongState)
	if !ok {
		panic("whoops")
	}
	svc, ok := arg[1].(*state.Target)
	if !ok {
		panic("whoops")
	}
	s.Targets.Add(*svc)
	return nil, nil
}

// Delete deletes the target from state.
// The first arg should be of type *state.KongState, the state from
// which the second arg, of type *state.Target, will be deleted.
// If the args are of incorrect types, Delete will panic.
func (crud *targetPostAction) Delete(arg ...crud.Arg) (crud.Arg, error) {
	s, ok := arg[0].(*state.KongState)
	if !ok {
		panic("whoops")
	}
	svc, ok := arg[1].(*state.Target)
	if !ok {
		panic("whoops")
	}
	s.Targets.Delete(*svc.ID)
	return nil, nil
}

// Update updates the target in state.
// The first arg should be of type *state.KongState, the state in
// which the second arg, of type *state.Target, will be updated.
// If the args are of incorrect types, Update will panic.
func (crud *targetPostAction) Update(arg ...crud.Arg) (crud.Arg, error) {
	s, ok := arg[0].(*state.KongState)
	if !ok {
		panic("whoops")
	}
	svc, ok := arg[1].(*state.Target)
	if !ok {
		panic("whoops")
	}
	s.Targets.Update(*svc)
	return nil, nil
}

type certificatePostAction struct{}

// Create creates the Certificate in state.
// The first arg should be of type *state.KongState, the state in
// which the second arg, of type *state.Certificate, will be added.
// If the args are of incorrect types, Create will panic.
func (crud *certificatePostAction) Create(arg ...crud.Arg) (crud.Arg, error) {
	s, ok := arg[0].(*state.KongState)
	if !ok {
		panic("whoops")
	}
	svc, ok := arg[1].(*state.Certificate)
	if !ok {
		panic("whoops")
	}
	s.Certificates.Add(*svc)
	return nil, nil
}

// Delete deletes the Certificate from state.
// The first arg should be of type *state.KongState, the state from
// which the second arg, of type *state.Certificate, will be deleted.
// If the args are of incorrect types, Delete will panic.
func (crud *certificatePostAction) Delete(arg ...crud.Arg) (crud.Arg, error) {
	s, ok := arg[0].(*state.KongState)
	if !ok {
		panic("whoops")
	}
	svc, ok := arg[1].(*state.Certificate)
	if !ok {
		panic("whoops")
	}
	s.Certificates.Delete(*svc.ID)
	return nil, nil
}

// Update updates the certificate in state.
// The first arg should be of type *state.KongState, the state in
// which the second arg, of type *state.Certificate, will be updated.
// If the args are of incorrect types, Update will panic.
func (crud *certificatePostAction) Update(arg ...crud.Arg) (crud.Arg, error) {
	s, ok := arg[0].(*state.KongState)
	if !ok {
		panic("whoops")
	}
	svc, ok := arg[1].(*state.Certificate)
	if !ok {
		panic("whoops")
	}
	s.Certificates.Update(*svc)
	return nil, nil
}

type pluginPostAction struct{}

// Create creates the Plugin in state.
// The first arg should be of type *state.KongState, the state in
// which the second arg, of type *state.Plugin, will be added.
// If the args are of incorrect types, Create will panic.
func (crud *pluginPostAction) Create(arg ...crud.Arg) (crud.Arg, error) {
	s, ok := arg[0].(*state.KongState)
	if !ok {
		panic("whoops")
	}
	svc, ok := arg[1].(*state.Plugin)
	if !ok {
		panic("whoops")
	}
	s.Plugins.Add(*svc)
	return nil, nil
}

// Delete deletes the Plugin from state.
// The first arg should be of type *state.KongState, the state from
// which the second arg, of type *state.Plugin, will be deleted.
// If the args are of incorrect types, Delete will panic.
func (crud *pluginPostAction) Delete(arg ...crud.Arg) (crud.Arg, error) {
	s, ok := arg[0].(*state.KongState)
	if !ok {
		panic("whoops")
	}
	svc, ok := arg[1].(*state.Plugin)
	if !ok {
		panic("whoops")
	}
	s.Plugins.Delete(*svc.ID)
	return nil, nil
}

// Update updates the plugin in state.
// The first arg should be of type *state.KongState, the state in
// which the second arg, of type *state.Plugin, will be updated.
// If the args are of incorrect types, Update will panic.
func (crud *pluginPostAction) Update(arg ...crud.Arg) (crud.Arg, error) {
	s, ok := arg[0].(*state.KongState)
	if !ok {
		panic("whoops")
	}
	svc, ok := arg[1].(*state.Plugin)
	if !ok {
		panic("whoops")
	}
	s.Plugins.Update(*svc)
	return nil, nil
}

type consumerPostAction struct{}

// Create creates the consumer from state.
// The first arg should be of type *state.KongState, the state in
// which the second arg, of type *state.Consumer, will be added.
// If the args are of incorrect types, Create will panic.
func (crud *consumerPostAction) Create(arg ...crud.Arg) (crud.Arg, error) {
	s, ok := arg[0].(*state.KongState)
	if !ok {
		panic("whoops")
	}
	svc, ok := arg[1].(*state.Consumer)
	if !ok {
		panic("whoops")
	}
	s.Consumers.Add(*svc)
	return nil, nil
}

// Delete deletes the consumer from state.
// The first arg should be of type *state.KongState, the state from
// which the second arg, of type *state.Consumer, will be deleted.
// If the args are of incorrect types, Delete will panic.
func (crud *consumerPostAction) Delete(arg ...crud.Arg) (crud.Arg, error) {
	s, ok := arg[0].(*state.KongState)
	if !ok {
		panic("whoops")
	}
	svc, ok := arg[1].(*state.Consumer)
	if !ok {
		panic("whoops")
	}
	s.Consumers.Delete(*svc.ID)
	return nil, nil
}

// Update updates the consumer from state.
// The first arg should be of type *state.KongState, the state in
// which the second arg, of type *state.Consumer, will be updated.
// If the args are of incorrect types, Update will panic.
func (crud *consumerPostAction) Update(arg ...crud.Arg) (crud.Arg, error) {
	s, ok := arg[0].(*state.KongState)
	if !ok {
		panic("whoops")
	}
	svc, ok := arg[1].(*state.Consumer)
	if !ok {
		panic("whoops")
	}
	s.Consumers.Update(*svc)
	return nil, nil
}
