package custom

// Type represents type of a custom entity in Kong.
type Type string

// Object is an instance of a custom entity definition
// in Kong.
type Object map[string]interface{}

// Entity represents an instance of a custom entity
// alongwith it's relations to other entities.
type Entity interface {
	// Type returns the type of the entity.
	Type() Type
	// Object returns the object, an instance
	// of a custom entity in Kong.
	Object() Object
	SetObject(Object)

	// AddRelation adds a foreign
	// relation with another entity's ID.
	AddRelation(string, string)
	// GetRelation should return foreign
	// entity's ID that is associated with Entity.
	GetRelation(string) string
	// GetAllRelations should return all
	// relationship of current Entity.
	GetAllRelations() map[string]string
}

// EntityObject is a default implmentation of Entity interface.
type EntityObject struct {
	ref    map[string]string
	object Object
	typ    Type
}

// NewEntityObject creates a new EntityObject
// of type typ with content of object and
// foreign references as defined in ref.
func NewEntityObject(typ Type) *EntityObject {
	return &EntityObject{
		typ: typ,
		ref: make(map[string]string),
	}
}

// Type returns the type of the entity.
// Type() Type
func (E *EntityObject) Type() Type {
	return E.typ
}

// Object returns the object, an instance
// of a custom entity in Kong.
func (E *EntityObject) Object() Object {
	return E.object
}

// SetObject sets the internal object
// to newObject.
func (E *EntityObject) SetObject(newObject Object) {
	E.object = newObject
}

// AddRelation adds a foreign
// relation with another entity's ID.
func (E *EntityObject) AddRelation(k, v string) {
	E.ref[k] = v
}

// GetRelation should return foreign
// entity's ID that is associated with Entity.
func (E *EntityObject) GetRelation(k string) string {
	return E.ref[k]
}

// GetAllRelations should return all
// relationship of current Entity.
func (E *EntityObject) GetAllRelations() map[string]string {
	res := make(map[string]string)
	for k, v := range E.ref {
		res[k] = v
	}
	return res
}
