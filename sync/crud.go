package sync

import (
	"fmt"

	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/go-kong/kong"
)

type Callback func(crud.Arg, error) error

type ServiceCRUD struct {
	// client    *kong.Client
	// callbacks []Callback // use this to update the current in-memory state
}

func args(arg ...crud.Arg) (*state.Service, *state.KongState, *state.KongState, *kong.Client) {
	r, ok := arg[0].(interface{})
	fmt.Println("r , ok Is ", r, ok)
	s, ok := r.(state.Service)
	fmt.Println("s, ok is ", s, ok)
	service, ok := arg[0].(*state.Service)
	fmt.Printf("%T\n", service)
	if !ok {
		panic("unexpected type, expected *state.Service")
	}
	currentState, ok := arg[1].(*state.KongState)
	if !ok {
		panic("unexpected type, expected *state.KongState")
	}
	targetState, ok := arg[2].(*state.KongState)
	if !ok {
		panic("unexpected type, expected *state.KongState")
	}
	client, ok := arg[3].(*kong.Client)
	if !ok {
		panic("unexpected type, expected *kong.Client")
	}

	return service, currentState, targetState, client
}

func (s *ServiceCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	service, _, _, client := args(arg...)
	fmt.Println("Create called with ", service)
	createdService, err := client.Services.Create(nil, &service.Service)
	if err != nil {
		return nil, err
	}
	fmt.Println("created: ", createdService)
	return createdService, nil
}
func (s *ServiceCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	service, _, _, client := args(arg...)
	fmt.Println("Delete called with ", service)
	err := client.Services.Delete(nil, service.ID)
	return nil, err
}
func (s *ServiceCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	service, _, _, client := args(arg...)
	fmt.Println("Update called with ", service)
	updatedService, err := client.Services.Update(nil, &service.Service)
	if err != nil {
		return nil, err
	}
	return updatedService, nil
}
