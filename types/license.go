package types

import (
	"context"
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/kong/go-kong/kong"
)

// licenseCRUD implements crud.Actions interface.
type licenseCRUD struct {
	client *kong.Client
}

func licenseFromStruct(arg crud.Event) *state.License {
	license, ok := arg.Obj.(*state.License)
	if !ok {
		panic("unexpected type, expected *state.License")
	}
	return license
}

// Create creates a License in Kong.
// The arg should be of type crud.Event, containing the license to be created,
// else the function will panic.
// It returns a the created *state.License.
func (s *licenseCRUD) Create(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	license := licenseFromStruct(event)
	createdLicense, err := s.client.Licenses.Create(ctx, &license.License)
	if err != nil {
		return nil, err
	}
	return &state.License{License: *createdLicense}, nil
}

// Delete deletes a License in Kong.
// The arg should be of type crud.Event, containing the license to be deleted,
// else the function will panic.
// It returns a the deleted *state.License.
func (s *licenseCRUD) Delete(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	license := licenseFromStruct(event)
	err := s.client.Licenses.Delete(ctx, license.ID)
	if err != nil {
		return nil, err
	}
	return license, nil
}

// Update updates a License in Kong.
// The arg should be of type crud.Event, containing the license to be updated,
// else the function will panic.
// It returns a the updated *state.License.
func (s *licenseCRUD) Update(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	license := licenseFromStruct(event)

	updatedLicense, err := s.client.Licenses.Create(ctx, &license.License)
	if err != nil {
		return nil, err
	}
	return &state.License{License: *updatedLicense}, nil
}

type licenseDiffer struct {
	kind crud.Kind

	currentState, targetState *state.KongState
}

// Deletes generates a memdb CRUD DELETE event for Licenses
// which is then consumed by the differ and used to gate Kong client calls.
func (d *licenseDiffer) Deletes(handler func(crud.Event) error) error {
	currentLicenses, err := d.currentState.Licenses.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching licenses from state: %w", err)
	}

	for _, license := range currentLicenses {
		n, err := d.deleteLicense(license)
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

func (d *licenseDiffer) deleteLicense(license *state.License) (*crud.Event, error) {
	_, err := d.targetState.Licenses.Get(*license.ID)
	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: "license",
			Obj:  license,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up license %q: %w",
			*license.ID, err)
	}
	return nil, nil
}

// CreateAndUpdates generates a memdb CRUD CREATE/UPDATE event for Licenses
// which is then consumed by the differ and used to gate Kong client calls.
func (d *licenseDiffer) CreateAndUpdates(handler func(crud.Event) error) error {
	targetLicenses, err := d.targetState.Licenses.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching licenses from state: %w", err)
	}

	for _, license := range targetLicenses {
		n, err := d.createUpdateLicense(license)
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

func (d *licenseDiffer) createUpdateLicense(license *state.License) (*crud.Event,
	error,
) {
	licenseCopy := &state.License{License: *license.DeepCopy()}
	currentLicense, err := d.currentState.Licenses.Get(*license.ID)

	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Create,
			Kind: "license",
			Obj:  licenseCopy,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up license %v: %w",
			*license.ID, err)
	}

	// found, check if update needed
	if !currentLicense.EqualWithOpts(licenseCopy, false, true) {
		return &crud.Event{
			Op:     crud.Update,
			Kind:   "license",
			Obj:    licenseCopy,
			OldObj: currentLicense,
		}, nil
	}
	return nil, nil
}
