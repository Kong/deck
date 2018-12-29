package kong

import (
	"github.com/kong/deck/crud"
	"github.com/kong/deck/diff"
	"github.com/kong/deck/state"
)

// ServiceCRUD implements Actions interface
// from the github.com/kong/crud package for the Service entitiy of Kong.
type ServiceCRUD struct {
	// client    *kong.Client
	// callbacks []Callback // use this to update the current in-memory state
}

func serviceFromStuct(arg diff.ArgStruct) *state.Service {
	service, ok := arg.Obj.(*state.Service)
	if !ok {
		panic("unexpected type, expected *state.service")
	}
	return service
}

// Create creates a Service in Kong. TODO Doc
func (s *ServiceCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	argStruct := argStructFromArg(arg[0])
	service := serviceFromStuct(argStruct)
	createdService, err := argStruct.Client.Services.Create(nil, &service.Service)
	if err != nil {
		return nil, err
	}
	return &state.Service{Service: *createdService}, nil
}

// Delete deletes a service in Kong. TODO Doc
func (s *ServiceCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	argStruct := argStructFromArg(arg[0])
	service := serviceFromStuct(argStruct)
	err := argStruct.Client.Services.Delete(nil, service.ID)
	if err != nil {
		return nil, err
	}
	return service, nil
}

// Update udpates a service in Kong. TODO Doc
func (s *ServiceCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	argStruct := argStructFromArg(arg[0])
	service := serviceFromStuct(argStruct)

	updatedService, err := argStruct.Client.Services.Update(nil, &service.Service)
	if err != nil {
		return nil, err
	}
	return &state.Service{Service: *updatedService}, nil
}
