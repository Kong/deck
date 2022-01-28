package utils

import (
	"context"
	"fmt"
	"reflect"

	"github.com/imdario/mergo"
	"github.com/kong/go-kong/kong"
)

// Defaulter registers types and fills in struct fields with
// default values.
type Defaulter struct {
	r map[string]interface{}

	ctx    context.Context
	client *kong.Client
}

// GetKongDefaulter returns a defaulter which can set default values
// for Kong entities.
func GetKongDefaulter() (*Defaulter, error) {
	var d Defaulter
	err := d.Register(&serviceDefaults)
	if err != nil {
		return nil, fmt.Errorf("registering service with defaulter: %w", err)
	}
	err = d.Register(&routeDefaults)
	if err != nil {
		return nil, fmt.Errorf("registering route with defaulter: %w", err)
	}
	err = d.Register(&upstreamDefaults)
	if err != nil {
		return nil, fmt.Errorf("registering upstream with defaulter: %w", err)
	}
	err = d.Register(&targetDefaults)
	if err != nil {
		return nil, fmt.Errorf("registering target with defaulter: %w", err)
	}
	return &d, nil
}

func (d *Defaulter) once() {
	if d.r == nil {
		d.r = make(map[string]interface{})
	}
}

// Register registers a type and it's default value.
// The default value is passed in and the type is inferred from the
// default value.
func (d *Defaulter) Register(def interface{}) error {
	d.once()
	v := reflect.ValueOf(def)
	if !v.IsValid() {
		return fmt.Errorf("invalid value")
	}
	v = reflect.Indirect(v)
	d.r[v.Type().String()] = def
	return nil
}

type kongTransformer struct{}

func (t kongTransformer) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	var a *int
	var ar []int
	var b *bool
	switch typ {
	case reflect.TypeOf(ar):

		return func(dst, src reflect.Value) error {
			if dst.CanSet() {
				if reflect.DeepEqual(reflect.Zero(dst.Type()).Interface(), dst.Interface()) {
					return nil
				}
			}
			return nil
		}
	case reflect.TypeOf(a):

		return func(dst, src reflect.Value) error {
			if dst.CanSet() {
				if reflect.DeepEqual(reflect.Zero(dst.Type()).Interface(), dst.Interface()) {
					return nil
				}
			}
			return nil
		}
	case reflect.TypeOf(b):

		return func(dst, src reflect.Value) error {
			if dst.CanSet() {
				if reflect.DeepEqual(reflect.Zero(dst.Type()).Interface(), dst.Interface()) {
					return nil
				}
			}
			return nil
		}
	default:

		return nil
	}
}

// Set fills in default values in a struct of a registered type.
func (d *Defaulter) Set(arg interface{}) error {
	d.once()
	v := reflect.ValueOf(arg)
	if !v.IsValid() {
		return fmt.Errorf("invalid value")
	}
	v = reflect.Indirect(v)
	defValue, ok := d.r[v.Type().String()]
	if !ok {
		return fmt.Errorf("type not registered: %v", reflect.TypeOf(arg))
	}
	err := mergo.Merge(arg, defValue, mergo.WithTransformers(kongTransformer{}))
	if err != nil {
		err = fmt.Errorf("merging: %w", err)
	}
	return err
	// return defaulter.Set(arg, defValue)
}

// MustSet is like Set but panics if there is an error.
func (d *Defaulter) MustSet(arg interface{}) {
	err := d.Set(arg)
	if err != nil {
		panic(err)
	}
}

func (d *Defaulter) getEntitySchema(entityType string) (map[string]interface{}, error) {
	var schema map[string]interface{}
	schema, err := d.client.Schemas.Get(d.ctx, entityType)
	if err != nil {
		return schema, err
	}
	return schema, nil
}

func (d *Defaulter) addEntityDefaults(entityType string, entity interface{}) error {
	schema, err := d.getEntitySchema(entityType)
	if err != nil {
		return fmt.Errorf("retrieve schema for %v from Kong: %v", entityType, err)
	}
	return kong.FillEntityDefaults(entity, schema)
}

func GetKongDefaulterWithClient(ctx context.Context, client *kong.Client) (*Defaulter, error) {
	d, err := GetKongDefaulter()
	if err != nil {
		return nil, err
	}
	d.ctx = ctx
	d.client = client

	route := kong.Route{}
	if err := d.addEntityDefaults("routes", &route); err != nil {
		return nil, fmt.Errorf("get defaults for routes: %v", err)
	}
	if err := d.Register(&route); err != nil {
		return nil, fmt.Errorf("registering route with defaulter: %w", err)
	}

	service := kong.Service{}
	if err := d.addEntityDefaults("services", &service); err != nil {
		return nil, fmt.Errorf("get defaults for services: %v", err)
	}
	if err := d.Register(&service); err != nil {
		return nil, fmt.Errorf("registering service with defaulter: %w", err)
	}

	upstream := kong.Upstream{}
	if err := d.addEntityDefaults("upstreams", &upstream); err != nil {
		return nil, fmt.Errorf("get defaults for upstreams: %v", err)
	}
	if err := d.Register(&upstream); err != nil {
		return nil, fmt.Errorf("registering upstream with defaulter: %w", err)
	}

	target := kong.Target{}
	if err := d.addEntityDefaults("targets", &target); err != nil {
		return nil, fmt.Errorf("get defaults for targets: %v", err)
	}
	if err := d.Register(&target); err != nil {
		return nil, fmt.Errorf("registering target with defaulter: %w", err)
	}
	return d, nil
}

func GetDefaulter(ctx context.Context, client *kong.Client) (*Defaulter, error) {
	if client != nil {
		return GetKongDefaulterWithClient(ctx, client)
	}
	return GetKongDefaulter()
}
