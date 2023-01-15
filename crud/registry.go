package crud

import (
	"context"
	"fmt"
)

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
		return fmt.Errorf("kind cannot be empty")
	}
	m := r.typesMap()
	if _, ok := m[kind]; ok {
		return fmt.Errorf("kind %q already registered", kind)
	}
	m[kind] = a
	return nil
}

// MustRegister is same as Register but panics on error.
func (r *Registry) MustRegister(kind Kind, a Actions) {
	err := r.Register(kind, a)
	if err != nil {
		panic(err)
	}
}

// Get returns actions associated with kind.
// An error will be returned if kind was never registered.
func (r *Registry) Get(kind Kind) (Actions, error) {
	if kind == "" {
		return nil, fmt.Errorf("kind cannot be empty")
	}
	m := r.typesMap()
	a, ok := m[kind]
	if !ok {
		return nil, fmt.Errorf("kind %q is not registered", kind)
	}
	return a, nil
}

// Create calls the registered create action of kind with args
// and returns the result and error (if any).
func (r *Registry) Create(ctx context.Context, kind Kind, args ...Arg) (Arg, error) {
	a, err := r.Get(kind)
	if err != nil {
		return nil, fmt.Errorf("create failed: %w", err)
	}

	res, err := a.Create(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("create failed: %w", err)
	}
	return res, nil
}

// Update calls the registered update action of kind with args
// and returns the result and error (if any).
func (r *Registry) Update(ctx context.Context, kind Kind, args ...Arg) (Arg, error) {
	a, err := r.Get(kind)
	if err != nil {
		return nil, fmt.Errorf("update failed: %w", err)
	}

	res, err := a.Update(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("update failed: %w", err)
	}
	return res, nil
}

// Delete calls the registered delete action of kind with args
// and returns the result and error (if any).
func (r *Registry) Delete(ctx context.Context, kind Kind, args ...Arg) (Arg, error) {
	a, err := r.Get(kind)
	if err != nil {
		return nil, fmt.Errorf("delete failed: %w", err)
	}

	res, err := a.Delete(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("delete failed: %w", err)
	}
	return res, nil
}

// Do calls an action based on op with args and returns the result and error.
func (r *Registry) Do(ctx context.Context, kind Kind, op Op, args ...Arg) (Arg, error) {
	a, err := r.Get(kind)
	if err != nil {
		return nil, fmt.Errorf("%v failed: %w", op, err)
	}

	var res Arg

	switch op.name {
	case Create.name:
		res, err = a.Create(ctx, args...)
	case Update.name:
		res, err = a.Update(ctx, args...)
	case Delete.name:
		res, err = a.Delete(ctx, args...)
	default:
		return nil, fmt.Errorf("unknown operation: %s", op.name)
	}

	if err != nil {
		return nil, err
	}
	return res, nil
}
