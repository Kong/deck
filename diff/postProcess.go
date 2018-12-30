package diff

import (
	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
)

type servicePostAction struct{}

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

// Delete deletes a service in Kong. TODO Doc
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

// Update udpates a service in Kong. TODO Doc
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

// Delete deletes a route in Kong. TODO Doc
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

// Update udpates a route in Kong. TODO Doc
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
