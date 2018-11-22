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
	Create Op = Op{"Create"}
	Update Op = Op{"Update"}
	Delete Op = Op{"Delete"}
)

type Arg interface{}

// Actions is an interface for CRUD operations on any entity
type Actions interface {
	Create(...Arg) (Arg, error)
	Delete(...Arg) (Arg, error)
	Update(...Arg) (Arg, error)
}
