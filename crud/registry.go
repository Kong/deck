package crud

import "github.com/pkg/errors"

type Kind string

type Registry struct {
	types map[Kind]Actions
}

func (r *Registry) typesMap() map[Kind]Actions {
	if r.types == nil {
		r.types = make(map[Kind]Actions)
	}
	return r.types
}

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

func (r *Registry) Create(kind Kind, arg Arg) (Arg, error) {
	a, err := r.Get(kind)
	if err != nil {
		return nil, errors.Wrap(err, "create failed")
	}

	res, err := a.Create(arg)
	if err != nil {
		return nil, errors.Wrap(err, "create failed")
	}
	return res, nil
}

func (r *Registry) Update(kind Kind, arg Arg) (Arg, error) {
	a, err := r.Get(kind)
	if err != nil {
		return nil, errors.Wrap(err, "update failed")
	}

	res, err := a.Update(arg)
	if err != nil {
		return nil, errors.Wrap(err, "update failed")
	}
	return res, nil
}

func (r *Registry) Delete(kind Kind, arg Arg) (Arg, error) {
	a, err := r.Get(kind)
	if err != nil {
		return nil, errors.Wrap(err, "delete failed")
	}

	res, err := a.Delete(arg)
	if err != nil {
		return nil, errors.Wrap(err, "delete failed")
	}
	return res, nil
}

func (r *Registry) Do(kind Kind, op Op, arg Arg) (Arg, error) {
	a, err := r.Get(kind)
	if err != nil {
		return nil, errors.Wrapf(err, "%v failed", op)
	}

	var res Arg

	switch op.name {
	case Create.name:
		res, err = a.Create(arg)
	case Update.name:
		res, err = a.Update(arg)
	case Delete.name:
		res, err = a.Delete(arg)
	default:
		return nil, errors.New("unknown operation: " + op.name)
	}

	if err != nil {
		return nil, errors.Wrapf(err, "%v failed", op)
	}
	return res, nil
}
