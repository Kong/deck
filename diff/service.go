package diff

import (
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
)

func (sc *Syncer) deleteServices() error {
	currentServices, err := sc.currentState.Services.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching services from state: %w", err)
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

func (sc *Syncer) deleteService(service *state.Service) (*crud.Event, error) {
	_, err := sc.targetState.Services.Get(*service.ID)
	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Delete,
			Kind: "service",
			Obj:  service,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up service %q: %w",
			service.Identifier(), err)
	}
	return nil, nil
}

func (sc *Syncer) createUpdateServices() error {
	targetServices, err := sc.targetState.Services.GetAll()
	if err != nil {
		return fmt.Errorf("error fetching services from state: %w", err)
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

func (sc *Syncer) createUpdateService(service *state.Service) (*crud.Event, error) {
	serviceCopy := &state.Service{Service: *service.DeepCopy()}
	currentService, err := sc.currentState.Services.Get(*service.ID)

	if err == state.ErrNotFound {
		return &crud.Event{
			Op:   crud.Create,
			Kind: "service",
			Obj:  serviceCopy,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error looking up service %q: %w",
			*service.Name, err)
	}

	// found, check if update needed
	if !currentService.EqualWithOpts(serviceCopy, false, true) {
		return &crud.Event{
			Op:     crud.Update,
			Kind:   "service",
			Obj:    serviceCopy,
			OldObj: currentService,
		}, nil
	}
	return nil, nil
}
