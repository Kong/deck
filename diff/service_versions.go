package diff

import (
	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/pkg/errors"
)

func (sc *Syncer) deleteServiceVersions() error {
	currentServiceVersions, err := sc.currentState.ServiceVersions.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching service-versions from state")
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
		return nil, errors.Wrapf(err, "looking up service-version '%v'",
			sv.Identifier())
	}
	return nil, nil
}

func (sc *Syncer) createUpdateServiceVersions() error {
	targetServiceVersions, err := sc.targetState.ServiceVersions.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching services from state")
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
		return nil, errors.Wrapf(err, "error looking up service-version %v",
			sv.Identifier())
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
