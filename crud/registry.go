package crud

import "github.com/pkg/errors"

// Kind represents Kind of an entity or object.
type Kind string

// Registry can hold Kinds and their respective CRUD operations.
type Registry struct {
	types map[Kind]Actions
}

func (r *Registry) typesMap() map[Kind]Actions {
	if r.types == nil {
		r.types = make(map[Kind]Actions)
	}
	return r.types
}

// Register a kind with actions.
// An error will be returned if kind was previously registered.
func (r *Registry) Register(kind Kind, a Actions) error {
	if kind == "" {
		return errors.New("kind cannot be empty")
	}
	m := r.typesMap()
	if _, ok := m[kind]; ok {
		return errors.New("kind '" + string(kind) + "' already registered")
	}
	m[kind] = a
	return nil
}

// Get returns actions associated with kind.
// An error will be returned if kind was never registered.
func (r *Registry) Get(kind Kind) (Actions, error) {
	if kind == "" {
		return nil, errors.New("kind cannot be empty")
	}
	m := r.typesMap()
	a, ok := m[kind]
	if !ok {
		return nil, errors.New("kind '" + string(kind) + "' is not registered")
	}
	return a, nil
}

// Create calls the registered create action of kind with args
// and returns the result and error (if any).
func (r *Registry) Create(kind Kind, args ...Arg) (Arg, error) {
	a, err := r.Get(kind)
	if err != nil {
		return nil, errors.Wrap(err, "create failed")
	}

	res, err := a.Create(args...)
	if err != nil {
		return nil, errors.Wrap(err, "create failed")
	}
	return res, nil
}

// Update calls the registered update action of kind with args
// and returns the result and error (if any).
func (r *Registry) Update(kind Kind, args ...Arg) (Arg, error) {
	a, err := r.Get(kind)
	if err != nil {
		return nil, errors.Wrap(err, "update failed")
	}

	res, err := a.Update(args...)
	if err != nil {
		return nil, errors.Wrap(err, "update failed")
	}
	return res, nil
}

// Delete calls the registered delete action of kind with args
// and returns the result and error (if any).
func (r *Registry) Delete(kind Kind, args ...Arg) (Arg, error) {
	a, err := r.Get(kind)
	if err != nil {
		return nil, errors.Wrap(err, "delete failed")
	}

	res, err := a.Delete(args...)
	if err != nil {
		return nil, errors.Wrap(err, "delete failed")
	}
	return res, nil
}

// Do calls an aciton based on op with args and returns the result and error.
func (r *Registry) Do(kind Kind, op Op, args ...Arg) (Arg, error) {
	a, err := r.Get(kind)
	if err != nil {
		return nil, errors.Wrapf(err, "%v failed", op)
	}

	var res Arg

	switch op.name {
	case Create.name:
		res, err = a.Create(args...)
	case Update.name:
		res, err = a.Update(args...)
	case Delete.name:
		res, err = a.Delete(args...)
	default:
		return nil, errors.New("unknown operation: " + op.name)
	}

	if err != nil {
		return nil, errors.Wrapf(err, "%v failed", op)
	}
	return res, nil
}
