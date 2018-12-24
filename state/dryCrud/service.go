package drycrud

import (
	"fmt"

	"github.com/hbagdi/go-kong/kong"
	"github.com/kong/deck/crud"
	"github.com/kong/deck/state"
)

// ServiceCRUD implements Actions interface
// from the github.com/kong/crud package for the Service entitiy of Kong.
type ServiceCRUD struct {
	// client    *kong.Client
	// callbacks []Callback // use this to update the current in-memory state
}

func argsForService(arg ...crud.Arg) (*state.Service, *state.KongState, *state.KongState, *kong.Client) {
	service, ok := arg[0].(*state.Service)
	if !ok {
		panic("unexpected type, expected *state.Service")
	}

	return service, nil, nil, nil
}

// Create creates a Service in Kong. TODO Doc
func (s *ServiceCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	service, _, _, _ := argsForService(arg...)
	fmt.Println("creating service", service)
	return nil, nil
}

// Delete deletes a service in Kong. TODO Doc
func (s *ServiceCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	service, _, _, _ := argsForService(arg...)
	fmt.Println("deleting service", service)
	return nil, nil
}

// Update udpates a service in Kong. TODO Doc
func (s *ServiceCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	service, _, _, _ := argsForService(arg...)
	fmt.Println("updating service", service)
	return nil, nil
}
