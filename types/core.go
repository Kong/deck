package types

import (
	"fmt"

	"github.com/kong/deck/crud"
	"github.com/kong/deck/konnect"
	"github.com/kong/deck/state"
	"github.com/kong/go-kong/kong"
)

func eventFromArg(arg crud.Arg) crud.Event {
	event, ok := arg.(crud.Event)
	if !ok {
		panic("unexpected type, expected diff.Event")
	}
	return event
}

type Entity interface {
	Type() string
	CRUDActions() crud.Actions
	PostProcessActions() crud.Actions
}

type entityImpl struct {
	typ                string
	cRUDActions        crud.Actions // needs to set client
	postProcessActions crud.Actions // needs currentstate Set
}

func (e entityImpl) Type() string {
	return e.typ
}

func (e entityImpl) CRUDActions() crud.Actions {
	return e.cRUDActions
}

func (e entityImpl) PostProcessActions() crud.Actions {
	return e.postProcessActions
}

type EntityOpts struct {
	CurrentState  *state.KongState
	TargetState   *state.KongState
	KongClient    *kong.Client
	KonnectClient *konnect.Client
}

const (
	Service = "service"
	Route   = "route"
)

func NewEntity(t string, opts EntityOpts) (Entity, error) {
	switch t {
	case Service:
		return entityImpl{
			typ: Service,
			cRUDActions: &serviceCRUD{
				client: opts.KongClient,
			},
			postProcessActions: &servicePostAction{
				CurrentState: opts.CurrentState,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown type: %q", t)
	}
}
