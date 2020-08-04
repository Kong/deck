package solver

import (
	"github.com/kong/deck/crud"
	"github.com/kong/deck/diff"
	"github.com/kong/deck/state"
	"github.com/hbagdi/go-kong/kong"
)

// sniCRUD implements crud.Actions interface.
type sniCRUD struct {
	client *kong.Client
}

func sniFromStuct(arg diff.Event) *state.SNI {
	sni, ok := arg.Obj.(*state.SNI)
	if !ok {
		panic("unexpected type, expected *state.SNI")
	}

	return sni
}

// Create creates a SNI in Kong.
// The arg should be of type diff.Event, containing the sni to be created,
// else the function will panic.
// It returns a the created *state.SNI.
func (s *sniCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	sni := sniFromStuct(event)
	createdSNI, err := s.client.SNIs.Create(nil, &sni.SNI)
	if err != nil {
		return nil, err
	}
	return &state.SNI{SNI: *createdSNI}, nil
}

// Delete deletes a SNI in Kong.
// The arg should be of type diff.Event, containing the sni to be deleted,
// else the function will panic.
// It returns a the deleted *state.SNI.
func (s *sniCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	sni := sniFromStuct(event)
	err := s.client.SNIs.Delete(nil, sni.ID)
	if err != nil {
		return nil, err
	}
	return sni, nil
}

// Update updates a SNI in Kong.
// The arg should be of type diff.Event, containing the sni to be updated,
// else the function will panic.
// It returns a the updated *state.SNI.
func (s *sniCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	sni := sniFromStuct(event)

	updatedSNI, err := s.client.SNIs.Create(nil, &sni.SNI)
	if err != nil {
		return nil, err
	}
	return &state.SNI{SNI: *updatedSNI}, nil
}
