package crud

import "context"

// Op represents
type Op struct {
	name string
}

func (op *Op) String() string {
	return op.name
}

var (
	// Create is a constant representing create operations.
	Create = Op{"Create"}
	// Update is a constant representing update operations.
	Update = Op{"Update"}
	// Delete is a constant representing delete operations.
	Delete = Op{"Delete"}
)

// Arg is an argument to a callback function.
type Arg interface{}

// Actions is an interface for CRUD operations on any entity
type Actions interface {
	Create(context.Context, ...Arg) (Arg, error)
	Delete(context.Context, ...Arg) (Arg, error)
	Update(context.Context, ...Arg) (Arg, error)
}

// Event represents an event to perform
// an imperative operation
// that gets Kong closer to the target state.
type Event struct {
	Op     Op
	Kind   Kind
	Obj    interface{}
	OldObj interface{}
}
