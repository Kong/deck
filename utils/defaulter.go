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

	service  *kong.Service
	route    *kong.Route
	upstream *kong.Upstream
	target   *kong.Target
}

// NewDefaulter initializes a Defaulter with empty entities.
func NewDefaulter() *Defaulter {
	return &Defaulter{
		service:  &kong.Service{},
		route:    &kong.Route{},
		upstream: &kong.Upstream{},
		target:   &kong.Target{},
	}
}

// GetKongDefaulter returns a defaulter which can set default values
// for Kong entities.
func GetKongDefaulter(kongDefaults interface{}) (*Defaulter, error) {
	d := NewDefaulter()
	d.populateDefaultsFromInput(kongDefaults)

	err := d.Register(d.service)
	if err != nil {
		return nil, fmt.Errorf("registering service with defaulter: %w", err)
	}
	err = d.Register(d.route)
	if err != nil {
		return nil, fmt.Errorf("registering route with defaulter: %w", err)
	}
	err = d.Register(d.upstream)
	if err != nil {
		return nil, fmt.Errorf("registering upstream with defaulter: %w", err)
	}
	err = d.Register(d.target)
	if err != nil {
		return nil, fmt.Errorf("registering target with defaulter: %w", err)
	}
	return d, nil
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

func getKongDefaulterWithClient(ctx context.Context, client *kong.Client,
	kongDefaults interface{}) (*Defaulter, error) {
	// fills defaults from input
	d, err := GetKongDefaulter(kongDefaults)
	if err != nil {
		return nil, err
	}
	d.ctx = ctx
	d.client = client

	// fills defaults from Kong API
	if err := d.addEntityDefaults("services", d.service); err != nil {
		return nil, fmt.Errorf("get defaults for services: %v", err)
	}
	if err := d.Register(d.service); err != nil {
		return nil, fmt.Errorf("registering service with defaulter: %w", err)
	}

	if err := d.addEntityDefaults("routes", d.route); err != nil {
		return nil, fmt.Errorf("get defaults for routes: %v", err)
	}
	if err := d.Register(d.route); err != nil {
		return nil, fmt.Errorf("registering route with defaulter: %w", err)
	}

	if err := d.addEntityDefaults("upstreams", d.upstream); err != nil {
		return nil, fmt.Errorf("get defaults for upstreams: %v", err)
	}
	if err := d.Register(d.upstream); err != nil {
		return nil, fmt.Errorf("registering upstream with defaulter: %w", err)
	}

	if err := d.addEntityDefaults("targets", d.target); err != nil {
		return nil, fmt.Errorf("get defaults for targets: %v", err)
	}
	if err := d.Register(d.target); err != nil {
		return nil, fmt.Errorf("registering target with defaulter: %w", err)
	}
	return d, nil
}

func GetDefaulter(ctx context.Context, client *kong.Client, kongDefaults interface{}) (*Defaulter, error) {
	if client != nil {
		return getKongDefaulterWithClient(ctx, client, kongDefaults)
	}
	return GetKongDefaulter(kongDefaults)
}

func (d *Defaulter) populateDefaultsFromInput(defaults interface{}) {
	r := reflect.ValueOf(defaults)

	service := reflect.Indirect(r).FieldByName("Service")
	serviceObj := service.Interface().(*kong.Service)
	if serviceObj != nil {
		d.service = serviceObj
	}

	route := reflect.Indirect(r).FieldByName("Route")
	routeObj := route.Interface().(*kong.Route)
	if routeObj != nil {
		d.route = routeObj
	}

	upstream := reflect.Indirect(r).FieldByName("Upstream")
	upstreamObj := upstream.Interface().(*kong.Upstream)
	if upstreamObj != nil {
		d.upstream = upstreamObj
	}

	target := reflect.Indirect(r).FieldByName("Target")
	targetObj := target.Interface().(*kong.Target)
	if targetObj != nil {
		d.target = target.Interface().(*kong.Target)
	}
}
