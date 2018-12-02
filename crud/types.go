package crud

type t string

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
	Create(...Arg) (Arg, error)
	Delete(...Arg) (Arg, error)
	Update(...Arg) (Arg, error)
}
