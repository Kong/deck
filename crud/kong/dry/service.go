package dry

import (
	"github.com/kong/deck/crud"
	arg "github.com/kong/deck/crud/kong"
	"github.com/kong/deck/print"
	"github.com/kong/deck/state"
)

// ServiceCRUD implements Actions interface
// from the github.com/kong/crud package for the Service entitiy of Kong.
type ServiceCRUD struct {
	// client    *kong.Client
	// callbacks []Callback // use this to update the current in-memory state
}

func serviceFromStuct(a arg.ArgStruct) *state.Service {
	service, ok := a.Obj.(*state.Service)
	if !ok {
		panic("unexpected type, expected *state.service")
	}

	return service
}

// Create creates a Service in Kong. TODO Doc
func (s *ServiceCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	argStruct := argStructFromArg(arg[0])
	service := serviceFromStuct(argStruct)

	print.CreatePrintln("creating service", service)
	return nil, nil
}

// Delete deletes a service in Kong. TODO Doc
func (s *ServiceCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	argStruct := argStructFromArg(arg[0])
	service := serviceFromStuct(argStruct)

	print.DeletePrintln("deleting service", service)
	return nil, nil
}

// Update udpates a service in Kong. TODO Doc
func (s *ServiceCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	argStruct := argStructFromArg(arg[0])
	service := serviceFromStuct(argStruct)

	print.UpdatePrintln("updating service", service)
	return nil, nil
}
