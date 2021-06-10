package types

import (
	"context"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/konnect"
	"github.com/kong/deck/state"
)

// servicePackageCRUD implements crud.Actions interface.
type servicePackageCRUD struct {
	client *konnect.Client
}

func servicePackageFromStruct(arg crud.Event) *state.ServicePackage {
	sp, ok := arg.Obj.(*state.ServicePackage)
	if !ok {
		panic("unexpected type, expected *state.ServicePackage")
	}
	return sp
}

// Create creates a Service package in Konnect.
// The arg should be of type crud.Event, containing the service to be created,
// else the function will panic.
// It returns a the created *state.ServicePackage.
func (s *servicePackageCRUD) Create(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	sp := servicePackageFromStruct(event)
	createdSP, err := s.client.ServicePackages.Create(ctx, &sp.ServicePackage)
	if err != nil {
		return nil, err
	}
	return &state.ServicePackage{ServicePackage: *createdSP}, nil
}

// Delete deletes a Service package in Konnect.
// The arg should be of type crud.Event, containing the service to be deleted,
// else the function will panic.
// It returns a the deleted *state.ServicePackage.
func (s *servicePackageCRUD) Delete(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	sp := servicePackageFromStruct(event)
	err := s.client.ServicePackages.Delete(ctx, sp.ID)
	if err != nil {
		return nil, err
	}
	return sp, nil
}

// Update updates a Service package in Konnect.
// The arg should be of type crud.Event, containing the service to be updated,
// else the function will panic.
// It returns a the updated *state.ServicePackage.
func (s *servicePackageCRUD) Update(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	sp := servicePackageFromStruct(event)

	updatedSP, err := s.client.ServicePackages.Update(ctx, &sp.ServicePackage)
	if err != nil {
		return nil, err
	}
	return &state.ServicePackage{ServicePackage: *updatedSP}, nil
}
