package types

import (
	"context"
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/kong/go-kong/kong"
)

// routeCRUD implements crud.Actions interface.
type routeCRUD struct {
	client *kong.Client
}

func routeFromStruct(arg crud.Event) *state.Route {
	route, ok := arg.Obj.(*state.Route)
	if !ok {
		panic("unexpected type, expected *state.Route")
	}

	return route
}

// Create creates a Route in Kong.
// The arg should be of type crud.Event, containing the route to be created,
// else the function will panic.
// It returns a the created *state.Route.
func (s *routeCRUD) Create(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	route := routeFromStruct(event)
	createdRoute, err := s.client.Routes.Create(ctx, &route.Route)
	if err != nil {
		return nil, err
	}
	return &state.Route{Route: *createdRoute}, nil
}

// Delete deletes a Route in Kong.
// The arg should be of type crud.Event, containing the route to be deleted,
// else the function will panic.
// It returns a the deleted *state.Route.
func (s *routeCRUD) Delete(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	route := routeFromStruct(event)
	err := s.client.Routes.Delete(ctx, route.ID)
	if err != nil {
		return nil, err
	}
	return route, nil
}

// Update updates a Route in Kong.
// The arg should be of type crud.Event, containing the route to be updated,
// else the function will panic.
// It returns a the updated *state.Route.
func (s *routeCRUD) Update(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	route := routeFromStruct(event)

	updatedRoute, err := s.client.Routes.Create(ctx, &route.Route)
	if err != nil {
		return nil, err
	}
	return &state.Route{Route: *updatedRoute}, nil
}

type routeDiffer struct {
	kind crud.Kind

	currentState, targetState *state.KongState
}

func (d *routeDiffer) Deletes(handler func(crud.Event) error) error {
	currentRoutes, err := d.currentState.Routes.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching routes from state: %w", err)
	}

	for _, route := range currentRoutes {
		n, err := d.deleteRoute(route)
		if err != nil {
			return err
		}
		if n != nil {
			err = handler(*n)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *routeDiffer) deleteRoute(route *state.Route) (*crud.Event, error) {
	_, err := d.targetState.Routes.Get(*route.ID)
	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: d.kind,
			Obj:  route,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up route %q: %w",
			route.Identifier(), err)
	}
	return nil, nil
}

func (d *routeDiffer) CreateAndUpdates(handler func(crud.Event) error) error {
	targetRoutes, err := d.targetState.Routes.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching routes from state: %w", err)
	}

	for _, route := range targetRoutes {
		n, err := d.createUpdateRoute(route)
		if err != nil {
			return err
		}
		if n != nil {
			err = handler(*n)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *routeDiffer) createUpdateRoute(route *state.Route) (*crud.Event, error) {
	route = &state.Route{Route: *route.DeepCopy()}
	currentRoute, err := d.currentState.Routes.Get(*route.ID)
	if err == state.ErrNotFound {
		// route not present, create it

		return &crud.Event{
			Op:   crud.Create,
			Kind: d.kind,
			Obj:  route,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up route %q: %w",
			route.Identifier(), err)
	}
	// found, check if update needed

	if !currentRoute.EqualWithOpts(route, false, true, false) {
		return &crud.Event{
			Op:     crud.Update,
			Kind:   d.kind,
			Obj:    route,
			OldObj: currentRoute,
		}, nil
	}
	return nil, nil
}
