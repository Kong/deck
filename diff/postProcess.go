package diff

import (
	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
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
