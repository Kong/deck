package diff

import (
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
)

func (sc *Syncer) deleteServiceVersions() error {
	currentServiceVersions, err := sc.currentState.ServiceVersions.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching service-versions from state: %w", err)
	}

	for _, sv := range currentServiceVersions {
		n, err := sc.deleteServiceVersion(sv)
		if err != nil {
			return err
		}
		if n != nil {
			err = sc.queueEvent(*n)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

func (sc *Syncer) deleteServiceVersion(sv *state.ServiceVersion) (*Event, error) {
	_, err := sc.targetState.ServiceVersions.Get(*sv.ServicePackage.ID, *sv.ID)
	if err == state.ErrNotFound {
		return &Event{
			Op:   crud.Delete,
			Kind: "service-version",
			Obj:  sv,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up service-version %q': %w",
			sv.Identifier(), err)
	}
	return nil, nil
}

func (sc *Syncer) createUpdateServiceVersions() error {
	targetServiceVersions, err := sc.targetState.ServiceVersions.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching services from state: %w", err)
	}

	for _, sv := range targetServiceVersions {
		n, err := sc.createUpdateServiceVersion(sv)
		if err != nil {
			return err
		}
		if n != nil {
			err = sc.queueEvent(*n)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (sc *Syncer) createUpdateServiceVersion(sv *state.ServiceVersion) (*Event, error) {
	svCopy := &state.ServiceVersion{ServiceVersion: *sv.DeepCopy()}
	currentSV, err := sc.currentState.ServiceVersions.Get(*sv.ServicePackage.ID, *sv.ID)

	if err == state.ErrNotFound {
		return &Event{
			Op:   crud.Create,
			Kind: "service-version",
			Obj:  svCopy,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up service-version %q: %w",
			sv.Identifier(), err)
	}

	// found, check if update needed
	if !currentSV.EqualWithOpts(svCopy, false, true, false) {
		return &Event{
			Op:     crud.Update,
			Kind:   "service-version",
			Obj:    svCopy,
			OldObj: currentSV,
		}, nil
	}
	return nil, nil
}
