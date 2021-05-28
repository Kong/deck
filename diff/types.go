package diff

import (
	"github.com/kong/deck/crud"
)

// Event represents an event to perform
// an imperative operation
// that gets Kong closer to the target state.
// TODO drop this type once this type is no longer in use
// This is kept around for gradual refactoring, crud.Event deprecates this type
type Event struct {
	Op     crud.Op
	Kind   crud.Kind
	Obj    interface{}
	OldObj interface{}
}
