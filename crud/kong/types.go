package kong

import (
	"github.com/hbagdi/go-kong/kong"
	"github.com/kong/deck/state"
)

// ArgStruct represents the input that Kong's
// CRUD operations take in as a parameter.
type ArgStruct struct {
	Obj, OldObj               interface{}
	CurrentState, TargetState *state.KongState
	Client                    *kong.Client
}
