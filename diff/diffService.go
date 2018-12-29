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
		ok, err := sc.deleteService(service)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		n := Node{
			Op:   crud.Delete,
			Kind: "service",
			Obj:  service,
		}
		sc.sendEvent(n)
	}
	return nil
}

func (sc *Syncer) deleteService(service *state.Service) (bool, error) {
	// lookup by name
	if utils.Empty(service.Name) {
		return false, errors.New("'name' attribute for a service cannot be nil")
	}
	_, err := sc.targetState.Services.Get(*service.Name)
	if err == state.ErrNotFound {
		return true, nil
	}
	if err != nil {
		return false, errors.Wrapf(err, "looking up service '%v'", *service.Name)
	}
	return false, nil
}

func (sc *Syncer) createUpdateServices() error {

	targetServices, err := sc.targetState.Services.GetAll()
	if err != nil {
		return errors.Wrap(err, "error fetching services from state")
	}

	for _, service := range targetServices {
		err := sc.createUpdateService(service)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sc *Syncer) createUpdateService(service *state.Service) error {
	serviceCopy := &state.Service{Service: *service.DeepCopy()}
	currentService, err := sc.currentState.Services.Get(*service.Name)

	if err == state.ErrNotFound {
		// service not present, create it
		serviceCopy.ID = nil
		n := Node{
			Op:   crud.Create,
			Kind: "service",
			Obj:  serviceCopy,
		}
		sc.sendEvent(n)
		return nil
	}
	if err != nil {
		return errors.Wrapf(err, "error looking up service %v", *service.Name)
	}

	// found, check if update needed
	if !currentService.EqualWithOpts(serviceCopy, true, true) {
		serviceCopy.ID = kong.String(*currentService.ID)
		n := Node{
			Op:     crud.Update,
			Kind:   "service",
			Obj:    serviceCopy,
			OldObj: currentService,
		}
		sc.sendEvent(n)
	}
	return nil
}
