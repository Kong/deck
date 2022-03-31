package utils

import (
	"context"
	"fmt"
	"reflect"
	"strings"

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

type DefaulterOpts struct {
	KongDefaults           interface{}
	DisableDynamicDefaults bool
	Client                 *kong.Client
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

func getKongDefaulter(opts DefaulterOpts) (*Defaulter, error) {
	d := NewDefaulter()
	if err := d.populateDefaultsFromInput(opts.KongDefaults); err != nil {
		return nil, err
	}

	if opts.DisableDynamicDefaults {
		if err := d.populateStaticDefaultsForKonnect(); err != nil {
			return nil, err
		}
	}

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

// Check if `entity` has restricted fields set
func checkEntityDefaults(entity interface{}, restrictedFields []string) error {
	var invalidFields []string
	r := reflect.ValueOf(entity)
	for _, fieldName := range restrictedFields {
		field := reflect.Indirect(r).FieldByName(fieldName)
		if field.IsValid() && !field.IsNil() {
			invalidFields = append(invalidFields, strings.ToLower(fieldName))
		}
	}
	if len(invalidFields) > 0 {
		return fmt.Errorf("cannot have these restricted fields set: %s",
			strings.Join(invalidFields, ", "))
	}
	return nil
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

func getKongDefaulterWithClient(ctx context.Context, opts DefaulterOpts) (*Defaulter, error) {
	// fills defaults from input
	d, err := getKongDefaulter(opts)
	if err != nil {
		return nil, err
	}
	d.ctx = ctx
	d.client = opts.Client

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

// GetDefaulter returns a Defaulter object to be used to set defaults
// on Kong entities. The order of precedence is as follow, from higher to lower:
//
// 1. values set in the state file
// 2. values set in the {_info: defaults:} object in the state file
// 3. schema defaults coming from Admin API (excluded Konnect)
// 4. hardcoded defaults under utils/constants.go (Konnect-only)
func GetDefaulter(ctx context.Context, opts DefaulterOpts) (*Defaulter, error) {
	exists, err := WorkspaceExists(ctx, opts.Client)
	if err != nil {
		return nil, fmt.Errorf("ensure workspace exists: %w", err)
	}
	if opts.Client != nil && !opts.DisableDynamicDefaults && exists {
		return getKongDefaulterWithClient(ctx, opts)
	}
	opts.DisableDynamicDefaults = true
	return getKongDefaulter(opts)
}

func (d *Defaulter) populateDefaultsFromInput(defaults interface{}) error {
	err := validateKongDefaults(defaults)
	if err != nil {
		return fmt.Errorf("validating defaults: %w", err)
	}

	r := reflect.ValueOf(defaults)

	service := reflect.Indirect(r).FieldByName("Service")
	serviceObj := service.Interface().(*kong.Service)
	if serviceObj != nil {
		err := mergo.Merge(d.service, serviceObj, mergo.WithTransformers(kongTransformer{}))
		if err != nil {
			return fmt.Errorf("merging: %w", err)
		}
	}

	route := reflect.Indirect(r).FieldByName("Route")
	routeObj := route.Interface().(*kong.Route)
	if routeObj != nil {
		err := mergo.Merge(d.route, routeObj, mergo.WithTransformers(kongTransformer{}))
		if err != nil {
			return fmt.Errorf("merging: %w", err)
		}
	}

	upstream := reflect.Indirect(r).FieldByName("Upstream")
	upstreamObj := upstream.Interface().(*kong.Upstream)
	if upstreamObj != nil {
		err := mergo.Merge(d.upstream, upstreamObj, mergo.WithTransformers(kongTransformer{}))
		if err != nil {
			return fmt.Errorf("merging: %w", err)
		}
	}

	target := reflect.Indirect(r).FieldByName("Target")
	targetObj := target.Interface().(*kong.Target)
	if targetObj != nil {
		err := mergo.Merge(d.target, targetObj, mergo.WithTransformers(kongTransformer{}))
		if err != nil {
			return fmt.Errorf("merging: %w", err)
		}
	}
	return nil
}

func validateKongDefaults(defaults interface{}) error {
	var errs ErrArray
	r := reflect.ValueOf(defaults)
	for objectName, restrictedFields := range defaultsRestrictedFields {
		objectValue := reflect.Indirect(r).FieldByName(objectName)
		if objectValue.IsNil() || !objectValue.IsValid() {
			continue
		}
		object := objectValue.Interface()
		err := checkEntityDefaults(object, restrictedFields)
		if err != nil {
			entityErr := fmt.Errorf(
				"%s defaults %s", strings.ToLower(objectName), err)
			errs.Errors = append(errs.Errors, entityErr)
		}
	}
	if errs.Errors != nil {
		return errs
	}
	return nil
}

func (d *Defaulter) populateStaticDefaultsForKonnect() error {
	if err := mergo.Merge(
		d.service, &serviceDefaults, mergo.WithTransformers(kongTransformer{}),
	); err != nil {
		return fmt.Errorf("merging service static defaults: %w", err)
	}
	if err := mergo.Merge(
		d.route, &routeDefaults, mergo.WithTransformers(kongTransformer{}),
	); err != nil {
		return fmt.Errorf("merging route static defaults: %w", err)
	}
	if err := mergo.Merge(
		d.upstream, &upstreamDefaults, mergo.WithTransformers(kongTransformer{}),
	); err != nil {
		return fmt.Errorf("merging upstream static defaults: %w", err)
	}
	if err := mergo.Merge(
		d.target, &targetDefaults, mergo.WithTransformers(kongTransformer{}),
	); err != nil {
		return fmt.Errorf("merging target static defaults: %w", err)
	}

	return nil
}
