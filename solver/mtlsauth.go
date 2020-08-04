package solver

import (
	"github.com/kong/deck/crud"
	"github.com/kong/deck/diff"
	"github.com/kong/deck/state"
	"github.com/kong/deck/utils"
	"github.com/kong/go-kong/kong"
)

// mtlsAuthCRUD implements crud.Actions interface.
type mtlsAuthCRUD struct {
	client *kong.Client
}

func mtlsAuthFromStuct(arg diff.Event) *state.MTLSAuth {
	mtlsAuth, ok := arg.Obj.(*state.MTLSAuth)
	if !ok {
		panic("unexpected type, expected *state.Route")
	}

	return mtlsAuth
}

// Create creates an mtls-auth credential in Kong.
// The arg should be of type diff.Event, containing the mtlsAuth to be created,
// else the function will panic.
// It returns a the created *state.Route.
func (s *mtlsAuthCRUD) Create(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	mtlsAuth := mtlsAuthFromStuct(event)
	createdMTLSAuth, err := s.client.MTLSAuths.Create(nil, mtlsAuth.Consumer.ID,
		&mtlsAuth.MTLSAuth)
	if err != nil {
		return nil, err
	}
	return &state.MTLSAuth{MTLSAuth: *createdMTLSAuth}, nil
}

// Delete deletes an mtls-auth credential in Kong.
// The arg should be of type diff.Event, containing the mtlsAuth to be deleted,
// else the function will panic.
// It returns a the deleted *state.Route.
func (s *mtlsAuthCRUD) Delete(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	mtlsAuth := mtlsAuthFromStuct(event)
	cid := ""
	if !utils.Empty(mtlsAuth.Consumer.Username) {
		cid = *mtlsAuth.Consumer.Username
	}
	if !utils.Empty(mtlsAuth.Consumer.ID) {
		cid = *mtlsAuth.Consumer.ID
	}
	err := s.client.MTLSAuths.Delete(nil, &cid, mtlsAuth.ID)
	if err != nil {
		return nil, err
	}
	return mtlsAuth, nil
}

// Update updates an mtls-auth credential in Kong.
// The arg should be of type diff.Event, containing the mtlsAuth to be updated,
// else the function will panic.
// It returns a the updated *state.Route.
func (s *mtlsAuthCRUD) Update(arg ...crud.Arg) (crud.Arg, error) {
	event := eventFromArg(arg[0])
	mtlsAuth := mtlsAuthFromStuct(event)

	updatedMTLSAuth, err := s.client.MTLSAuths.Create(nil, mtlsAuth.Consumer.ID,
		&mtlsAuth.MTLSAuth)
	if err != nil {
		return nil, err
	}
	return &state.MTLSAuth{MTLSAuth: *updatedMTLSAuth}, nil
}
