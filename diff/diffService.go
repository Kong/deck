package diff

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	"github.com/hbagdi/go-kong/kong"
	"github.com/pkg/errors"
)

func (sc *Syncer) deleteServices() error {
	currentServices, err := sc.currentState.GetAllServices()
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
		n := &Node{
			Op:   crud.Delete,
			Kind: "service",
			Obj:  service,
		}
		sc.deleteGraph.Add(n)
		service.AddMeta(nodeKey, n)
		sc.currentState.UpdateService(*service)
	}
	return nil
}

func (sc *Syncer) deleteService(service *state.Service) (bool, error) {
	// lookup by name
	if utils.Empty(service.Name) {
		return false, errors.New("'name' attribute for a service cannot be nil")
	}
	_, err := sc.targetState.GetService(*service.Name)
	if err == state.ErrNotFound {
		return true, nil
	}
	// any other type of error
	if err != nil {
		return false, err
	}
	return false, nil
}

func (sc *Syncer) createUpdateServices() error {

	targetServices, err := sc.targetState.GetAllServices()
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
	service = &state.Service{Service: *service.DeepCopy()}
	s, err := sc.currentState.GetService(*service.Name)
	if err == state.ErrNotFound {
		service.ID = nil
		n := &Node{
			Op:   crud.Create,
			Kind: "service",
			Obj:  service,
		}
		sc.createUpdateGraph.Add(n)
		service.AddMeta(nodeKey, n)
		sc.currentState.UpdateService(*service)
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "error looking up service")
	}
	// if found, check if update needed
	if !s.EqualWithOpts(service, true, true) {
		service.ID = kong.String(*s.ID)
		n := &Node{
			Op:   crud.Update,
			Kind: "service",
			Obj:  service,
		}
		sc.createUpdateGraph.Add(n)
		service.AddMeta(nodeKey, n)
		sc.currentState.UpdateService(*service)
	}
	return nil
}
