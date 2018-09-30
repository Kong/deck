package custom

import "errors"

// Registry is a store of EntityCRUD objects
type Registry interface {
	// Register puts EntityCRUD in the internal
	// store and returns an error if the entity
	// of Type is already registered
	Register(Type, EntityCRUD) error
	// Lookup returns the EntityCRUD object associated
	// with typ, or nil if one is not Registered yet.
	Lookup(Type) EntityCRUD
	// Unregister unregisters the entity
	// registered with Type from the store.
	// It returns an
	// error if the Type was not registered
	// before this call.
	Unregister(Type) error
}

// defaultRegistry is an out of the box implementation
// of Registry object.
type defaultRegistry struct {
	store map[Type]EntityCRUD
}

// NewDefaultRegistry returns a default registry
func NewDefaultRegistry() Registry {
	return &defaultRegistry{
		store: make(map[Type]EntityCRUD),
	}
}

// Register puts EntityCRUD in the internal
// store and returns an error if the entity
// of Type is already registered.
func (r *defaultRegistry) Register(typ Type, def EntityCRUD) error {
	if _, ok := r.store[typ]; ok {
		return errors.New("type already registered")
	}
	r.store[typ] = def
	return nil
}

// Lookup returns the EntityCRUD object associated
// with typ, or nil if one is not Registered yet.
func (r *defaultRegistry) Lookup(typ Type) EntityCRUD {
	return r.store[typ]
}

// Unregister unregisters the entity
// registered with Type from the store.
// It returns an
// error if the Type was not registered
// before this call.
func (r *defaultRegistry) Unregister(typ Type) error {
	if _, ok := r.store[typ]; !ok {
		return errors.New("type not registered")
	}
	delete(r.store, typ)
	return nil
}
