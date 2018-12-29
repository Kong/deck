package dry

import (
	"github.com/hbagdi/go-kong/kong"
	"github.com/kong/deck/crud"
	arg "github.com/kong/deck/diff"
	"github.com/kong/deck/print"
	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
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

	print.CreatePrintln("creating service", *service.Name)
	service.ID = kong.String(utils.UUID())
	return service, nil
}

// Delete deletes a service in Kong. TODO Doc
func (s *ServiceCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	argStruct := argStructFromArg(arg[0])
	service := serviceFromStuct(argStruct)

	print.DeletePrintln("deleting service", *service.Name)
	return service, nil
}

// Update udpates a service in Kong. TODO Doc
func (s *ServiceCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	argStruct := argStructFromArg(arg[0])
	service := serviceFromStuct(argStruct)
	oldServiceObj, ok := argStruct.OldObj.(*state.Service)
	if !ok {
		panic("unexpected type, expected *state.service")
	}
	oldService := oldServiceObj.DeepCopy()
	// TODO remove this hack
	oldService.CreatedAt = nil
	oldService.UpdatedAt = nil
	diff := getDiff(oldService, &service.Service)
	print.UpdatePrintln("updating service", *service.Name)
	print.UpdatePrintf("%s", diff)
	return service, nil
}
