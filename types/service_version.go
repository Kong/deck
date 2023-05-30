package types

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/konnect"
	"github.com/kong/deck/state"
)

// serviceVersionCRUD implements crud.Actions interface.
type serviceVersionCRUD struct {
	client *konnect.Client
}

func serviceVersionFromStruct(arg crud.Event) *state.ServiceVersion {
	sv, ok := arg.Obj.(*state.ServiceVersion)
	if !ok {
		panic("unexpected type, expected *state.ServiceVersion")
	}
	return sv
}

func oldServiceVersionFromStruct(arg crud.Event) *state.ServiceVersion {
	sv, ok := arg.OldObj.(*state.ServiceVersion)
	if !ok {
		panic("unexpected type, expected *state.ServiceVersion")
	}
	return sv
}

// Create creates a Service version in Konnect.
// The arg should be of type crud.Event, containing the service to be created,
// else the function will panic.
// It returns a the created *state.ServiceVersion.
func (s *serviceVersionCRUD) Create(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	sv := serviceVersionFromStruct(event)
	createdSV, err := s.client.ServiceVersions.Create(ctx, &sv.ServiceVersion)
	if err != nil {
		return nil, err
	}
	if sv.ControlPlaneServiceRelation != nil {
		_, err := s.client.ControlPlaneRelations.Create(ctx, &konnect.ControlPlaneServiceRelationCreateRequest{
			ServiceVersionID:     *createdSV.ID,
			ControlPlaneEntityID: *sv.ControlPlaneServiceRelation.ControlPlaneEntityID,
		})
		if err != nil {
			return nil, err
		}
	}
	return &state.ServiceVersion{ServiceVersion: *createdSV}, nil
}

// Delete deletes a Service version in Konnect.
// The arg should be of type crud.Event, containing the service to be deleted,
// else the function will panic.
// It returns a the deleted *state.ServiceVersion.
func (s *serviceVersionCRUD) Delete(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	event := crud.EventFromArg(arg[0])
	sv := serviceVersionFromStruct(event)
	err := s.client.ServiceVersions.Delete(ctx, sv.ID)
	if err != nil {
		return nil, err
	}
	return sv, nil
}

// Update updates a Service version in Konnect.
// The arg should be of type crud.Event, containing the service to be updated,
// else the function will panic.
// It returns a the updated *state.ServiceVersion.
func (s *serviceVersionCRUD) Update(ctx context.Context, arg ...crud.Arg) (crud.Arg, error) {
	var (
		err       error
		updatedSV *konnect.ServiceVersion
	)
	event := crud.EventFromArg(arg[0])
	version := serviceVersionFromStruct(event)
	oldVersion := oldServiceVersionFromStruct(event)

	// if there is a change in service version entity, make a PATCH
	if !version.EqualWithOpts(oldVersion, false, true, true) {
		versionCopy := &state.ServiceVersion{ServiceVersion: *version.DeepCopy()}
		versionCopy.ControlPlaneServiceRelation = nil
		versionCopy.ServicePackage = nil
		updatedSV, err = s.client.ServiceVersions.Update(ctx, &versionCopy.ServiceVersion)
		if err != nil {
			return nil, err
		}
	} else {
		updatedSV = &version.ServiceVersion
	}

	// When a service versions update is detected, it could be due to changes in
	// control-plane-entity and service-version relations
	// This is possible only during update events
	err = s.relationCRUD(ctx, &version.ServiceVersion, &oldVersion.ServiceVersion)
	if err != nil {
		return nil, err
	}

	return &state.ServiceVersion{ServiceVersion: *updatedSV}, nil
}

func (s *serviceVersionCRUD) relationCRUD(ctx context.Context, version,
	oldVersion *konnect.ServiceVersion,
) error {
	var err error

	if version.ControlPlaneServiceRelation != nil &&
		oldVersion.ControlPlaneServiceRelation == nil {
		// no version existed before, create a new relation
		_, err = s.client.ControlPlaneRelations.Create(ctx, &konnect.ControlPlaneServiceRelationCreateRequest{
			ServiceVersionID:     *version.ID,
			ControlPlaneEntityID: *version.ControlPlaneServiceRelation.ControlPlaneEntityID,
		})
	} else if version.ControlPlaneServiceRelation == nil && oldVersion.
		ControlPlaneServiceRelation != nil {
		// version doesn't need to exist anymore, delete it
		err = s.client.ControlPlaneRelations.Delete(ctx,
			oldVersion.ControlPlaneServiceRelation.ID)
	} else if !reflect.DeepEqual(version.ControlPlaneServiceRelation,
		oldVersion.ControlPlaneServiceRelation) {
		// relations are different, update it
		_, err = s.client.ControlPlaneRelations.Update(ctx,
			&konnect.ControlPlaneServiceRelationUpdateRequest{
				ID: *oldVersion.ControlPlaneServiceRelation.ID,
				ControlPlaneServiceRelationCreateRequest: konnect.ControlPlaneServiceRelationCreateRequest{
					ServiceVersionID:     *version.ID,
					ControlPlaneEntityID: *version.ControlPlaneServiceRelation.ControlPlaneEntityID,
				},
			})
	}
	return err
}

type serviceVersionDiffer struct {
	kind crud.Kind

	currentState, targetState *state.KongState
}

func (d *serviceVersionDiffer) Deletes(handler func(crud.Event) error) error {
	currentServiceVersions, err := d.currentState.ServiceVersions.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching service-versions from state: %w", err)
	}

	for _, sv := range currentServiceVersions {
		n, err := d.deleteServiceVersion(sv)
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

func (d *serviceVersionDiffer) deleteServiceVersion(sv *state.ServiceVersion) (*crud.Event, error) {
	_, err := d.targetState.ServiceVersions.Get(*sv.ServicePackage.ID, *sv.ID)
	if errors.Is(err, state.ErrNotFound) {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: d.kind,
			Obj:  sv,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up service-version %q': %w",
			sv.Identifier(), err)
	}
	return nil, nil
}

func (d *serviceVersionDiffer) CreateAndUpdates(handler func(crud.Event) error) error {
	targetServiceVersions, err := d.targetState.ServiceVersions.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching services from state: %w", err)
	}

	for _, sv := range targetServiceVersions {
		n, err := d.createUpdateServiceVersion(sv)
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

func (d *serviceVersionDiffer) createUpdateServiceVersion(sv *state.ServiceVersion) (*crud.Event, error) {
	svCopy := &state.ServiceVersion{ServiceVersion: *sv.DeepCopy()}
	currentSV, err := d.currentState.ServiceVersions.Get(*sv.ServicePackage.ID, *sv.ID)

	if errors.Is(err, state.ErrNotFound) {
		return &crud.Event{
			Op:   crud.Create,
			Kind: d.kind,
			Obj:  svCopy,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up service-version %q: %w",
			sv.Identifier(), err)
	}

	// found, check if update needed
	if !currentSV.EqualWithOpts(svCopy, false, true, false) {
		return &crud.Event{
			Op:     crud.Update,
			Kind:   d.kind,
			Obj:    svCopy,
			OldObj: currentSV,
		}, nil
	}
	return nil, nil
}
