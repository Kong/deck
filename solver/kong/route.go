package kong

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/diff"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/go-kong/kong"
	"github.com/pkg/errors"
)

// RouteCRUD implements Actions interface
// from the github.com/kong/crud package for the Route entitiy of Kong.
type RouteCRUD struct {
	client *kong.Client
}

// NewRouteCRUD creates a new RouteCRUD. Client is required.
func NewRouteCRUD(client *kong.Client) (*RouteCRUD, error) {
	if client == nil {
		return nil, errors.New("client is required")
	}
	return &RouteCRUD{
		client: client,
	}, nil
}

func routeFromStuct(arg diff.Event) *state.Route {
	route, ok := arg.Obj.(*state.Route)
	if !ok {
		panic("unexpected type, expected *state.Route")
	}

	return route
}

// Create creates a Route in Kong.
// The arg should be of type diff.Event, containing the route to be created,
// else the function will panic.
// It returns a the created *state.Route.
func (s *RouteCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	route := routeFromStuct(event)
	createdRoute, err := s.client.Routes.Create(nil, &route.Route)
	if err != nil {
		return nil, err
	}
	return &state.Route{Route: *createdRoute}, nil
}

// Delete deletes a Route in Kong.
// The arg should be of type diff.Event, containing the route to be deleted,
// else the function will panic.
// It returns a the deleted *state.Route.
func (s *RouteCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	route := routeFromStuct(event)
	err := s.client.Routes.Delete(nil, route.ID)
	if err != nil {
		return nil, err
	}
	return route, nil
}

// Update updates a Route in Kong.
// The arg should be of type diff.Event, containing the route to be updated,
// else the function will panic.
// It returns a the updated *state.Route.
func (s *RouteCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	route := routeFromStuct(event)

	updatedRoute, err := s.client.Routes.Create(nil, &route.Route)
	if err != nil {
		return nil, err
	}
	return &state.Route{Route: *updatedRoute}, nil
}
