package diff

import (
	"github.com/hbagdi/go-kong/kong"
	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
	"github.com/pkg/errors"
)

func (sc *Syncer) deleteServices() error {
	currentServices, err := sc.currentState.Services.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching services from state")
	}

	for _, service := range currentServices {
		n, err := sc.deleteService(service)
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

func (sc *Syncer) deleteService(service *state.Service) (*Event, error) {
	// lookup by name
	if utils.Empty(service.Name) {
		return nil, errors.New("'name' attribute for a service cannot be nil")
	}
	_, err := sc.targetState.Services.Get(*service.Name)
	if err == state.ErrNotFound {
		return &Event{
			Op:   crud.Delete,
			Kind: "service",
			Obj:  service,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "looking up service '%v'", *service.Name)
	}
	return nil, nil
}

func (sc *Syncer) createUpdateServices() error {
	targetServices, err := sc.targetState.Services.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching services from state")
	}

	for _, service := range targetServices {
		n, err := sc.createUpdateService(service)
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

func (sc *Syncer) createUpdateService(service *state.Service) (*Event, error) {
	serviceCopy := &state.Service{Service: *service.DeepCopy()}
	currentService, err := sc.currentState.Services.Get(*service.Name)

	if err == state.ErrNotFound {
		// service not present, create it
		serviceCopy.ID = nil
		return &Event{
			Op:   crud.Create,
			Kind: "service",
			Obj:  serviceCopy,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error looking up service %v",
			*service.Name)
	}

	// found, check if update needed
	if !currentService.EqualWithOpts(serviceCopy, true, true) {
		serviceCopy.ID = kong.String(*currentService.ID)
		return &Event{
			Op:     crud.Update,
			Kind:   "service",
			Obj:    serviceCopy,
			OldObj: currentService,
		}, nil
	}
	return nil, nil
}
