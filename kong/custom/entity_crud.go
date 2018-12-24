package custom

import (
	"errors"
	"regexp"
	"strings"
)

// EntityCRUD defines endpoints on Kong's
// Admin API to interact with the Custom entity in Kong.
// Various *Endpoint methods can render
// the endpoint to interact with a custom
// entity in Kong. The RESTful endpoints are dynamically
// generated since based on foreign relations,
// the URLs change.
type EntityCRUD interface {
	// Type return the type of custom entitiy
	// (like key-auth, basic-auth, acl).
	Type() Type
	// GetEndpoint returns the URL to get
	// an existing entity e.
	// This is useful when one has the foreign relations
	// and the primary key for that entity.
	GetEndpoint(e Entity) (string, error)
	// PostEndpoint returns the URL to create
	// an entity e of a Type.
	PostEndpoint(e Entity) (string, error)
	// PatchEndpoint returns the URL to use for updating
	// custom entity e.
	PatchEndpoint(e Entity) (string, error)
	// DeleteEndpoint returns the URL to use
	// to delete entity e.
	DeleteEndpoint(e Entity) (string, error)
	// ListEndpoint returns the list URL.
	// This can be used to list all
	// instances of a type.
	ListEndpoint(e Entity) (string, error)
}

// EntityCRUDDefinition implements the EntityCRUD interface.
type EntityCRUDDefinition struct {
	Name       Type   `yaml:"name" json:"name"`
	CRUDPath   string `yaml:"crud" json:"curd"`
	PrimaryKey string `yaml:"primary_key" json:"primary_key"`
}

var r = regexp.MustCompile("(?:\\$\\{)(\\w+)(?:\\})")

func render(template string, entity Entity) (string, error) {
	result := template
	matches := r.FindAllStringSubmatch(template, -1)
	for _, m := range matches {
		if v := entity.GetRelation(m[1]); v != "" {
			result = strings.Replace(result, m[0], v, 1)
		} else {
			return "", errors.New("cannot substitute '" + m[1] +
				"' in URL: " + template)
		}
	}
	return result, nil
}

func (e EntityCRUDDefinition) renderWithPK(entity Entity) (string, error) {
	endpoint, err := render(e.CRUDPath, entity)
	if err != nil {
		return "", err
	}
	p, ok := entity.Object()[e.PrimaryKey]
	if !ok {
		return "", errors.New("primary key not found in entity")
	}
	key, ok := p.(string)
	if !ok {
		return "", errors.New("primary key can't be converted to string")
	}
	return endpoint + "/" + key, nil
}

// Type return the type of custom entitiy in Kong.
func (e *EntityCRUDDefinition) Type() Type {
	return e.Name
}

// GetEndpoint returns the URL to get
// an existing entity e.
// This is useful when one has the foreign relations
// and the primary key for that entity.
func (e *EntityCRUDDefinition) GetEndpoint(entity Entity) (string, error) {
	return e.renderWithPK(entity)
}

// PostEndpoint returns the URL to create
// an entity e of a Type.
func (e *EntityCRUDDefinition) PostEndpoint(entity Entity) (string, error) {
	return render(e.CRUDPath, entity)
}

// PatchEndpoint returns the URL to use for updating
// custom entity e.
func (e *EntityCRUDDefinition) PatchEndpoint(entity Entity) (string, error) {
	return e.renderWithPK(entity)
}

// DeleteEndpoint returns the URL to use
// to delete entity e.
func (e *EntityCRUDDefinition) DeleteEndpoint(entity Entity) (string, error) {
	return e.renderWithPK(entity)
}

// ListEndpoint returns the list URL.
// This can be used to list all
// instances of a type.
func (e EntityCRUDDefinition) ListEndpoint(entity Entity) (string, error) {
	return render(e.CRUDPath, entity)
}
