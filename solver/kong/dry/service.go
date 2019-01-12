package dry

import (
	"github.com/hbagdi/deck/crud"
	"github.com/hbagdi/deck/diff"
	"github.com/hbagdi/deck/print"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/deck/utils"
	"github.com/hbagdi/go-kong/kong"
)

// ServiceCRUD implements Actions interface
// from the github.com/kong/crud package for the Service entitiy of Kong.
type ServiceCRUD struct {
	// client    *kong.Client
	// callbacks []Callback // use this to update the current in-memory state
}

func serviceFromStuct(a diff.Event) *state.Service {
	service, ok := a.Obj.(*state.Service)
	if !ok {
		panic("unexpected type, expected *state.service")
	}

	return service
}

// Create creates a fake service.
// The arg should be of type diff.Event, containing the service to be created,
// else the function will panic.
// It returns a the created *state.Service.
func (s *ServiceCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	service := serviceFromStuct(event)

	print.CreatePrintln("creating service", *service.Name)
	service.ID = kong.String(utils.UUID())
	return service, nil
}

// Delete deletes a fake service.
// The arg should be of type diff.Event, containing the service to be deleted,
// else the function will panic.
// It returns a the deleted *state.Service.
func (s *ServiceCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	service := serviceFromStuct(event)

	print.DeletePrintln("deleting service", *service.Name)
	return service, nil
}

// Update updates a fake service.
// The arg should be of type diff.Event, containing the service to be updated,
// else the function will panic.
// It returns a the updated *state.Service.
func (s *ServiceCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	service := serviceFromStuct(event)
	oldServiceObj, ok := event.OldObj.(*state.Service)
	if !ok {
		panic("unexpected type, expected *state.service")
	}
	oldService := oldServiceObj.DeepCopy()
	// TODO remove this hack
	oldService.CreatedAt = nil
	oldService.UpdatedAt = nil
	diff := getDiff(oldService, &service.Service)
	print.UpdatePrintf("updating service %s\n%s", *service.Name, diff)
	return service, nil
}
