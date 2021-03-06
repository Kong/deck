package diff

import (
	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/pkg/errors"
)

func (sc *Syncer) deleteServicePackages() error {
	currentServicePackages, err := sc.currentState.ServicePackages.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching services-packages from state")
	}

	for _, sp := range currentServicePackages {
		n, err := sc.deleteServicePackage(sp)
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

func (sc *Syncer) deleteServicePackage(sp *state.ServicePackage) (*Event, error) {
	_, err := sc.targetState.ServicePackages.Get(*sp.ID)
	if err == state.ErrNotFound {
		return &Event{
			Op:   crud.Delete,
			Kind: "service-package",
			Obj:  sp,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "looking up service-package '%v'",
			sp.Identifier())
	}
	return nil, nil
}

func (sc *Syncer) createUpdateServicePackages() error {
	targetServicePackages, err := sc.targetState.ServicePackages.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching services-packages from state")
	}

	for _, sp := range targetServicePackages {
		n, err := sc.createUpdateServicePackage(sp)
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

func (sc *Syncer) createUpdateServicePackage(sp *state.ServicePackage) (*Event, error) {
	spCopy := &state.ServicePackage{ServicePackage: *sp.DeepCopy()}
	currentSP, err := sc.currentState.ServicePackages.Get(*sp.ID)

	if err == state.ErrNotFound {
		return &Event{
			Op:   crud.Create,
			Kind: "service-package",
			Obj:  spCopy,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error looking up service-package %v",
			sp.Identifier())
	}

	// found, check if update needed
	if !currentSP.EqualWithOpts(spCopy, false, true) {
		return &Event{
			Op:     crud.Update,
			Kind:   "service-package",
			Obj:    spCopy,
			OldObj: currentSP,
		}, nil
	}
	return nil, nil
}
