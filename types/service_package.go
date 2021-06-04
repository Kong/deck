package types

import (
	"context"
	"fmt"

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

type servicePackageDiffer struct {
	kind crud.Kind

	currentState, targetState *state.KongState
}

func (d *servicePackageDiffer) Deletes(handler func(crud.Event) error) error {
	currentServicePackages, err := d.currentState.ServicePackages.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching services-packages from state: %w", err)
	}

	for _, sp := range currentServicePackages {
		n, err := d.deleteServicePackage(sp)
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

func (d *servicePackageDiffer) deleteServicePackage(sp *state.ServicePackage) (*crud.Event, error) {
	_, err := d.targetState.ServicePackages.Get(*sp.ID)
	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: "service-package",
			Obj:  sp,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up service-package %q: %w",
			sp.Identifier(), err)
	}
	return nil, nil
}

func (d *servicePackageDiffer) CreateAndUpdates(handler func(crud.Event) error) error {
	targetServicePackages, err := d.targetState.ServicePackages.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching services-packages from state: %w", err)
	}

	for _, sp := range targetServicePackages {
		n, err := d.createUpdateServicePackage(sp)
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

func (d *servicePackageDiffer) createUpdateServicePackage(sp *state.ServicePackage) (*crud.Event, error) {
	spCopy := &state.ServicePackage{ServicePackage: *sp.DeepCopy()}
	currentSP, err := d.currentState.ServicePackages.Get(*sp.ID)

	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Create,
			Kind: "service-package",
			Obj:  spCopy,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up service-package %q: %w",
			sp.Identifier(), err)
	}

	// found, check if update needed
	if !currentSP.EqualWithOpts(spCopy, false, true) {
		return &crud.Event{
			Op:     crud.Update,
			Kind:   "service-package",
			Obj:    spCopy,
			OldObj: currentSP,
		}, nil
	}
	return nil, nil
}
