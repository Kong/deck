package utils

import (
	"fmt"
	"reflect"

	"github.com/imdario/mergo"
)

// Defaulter registers types and fills in struct fields with
// default values.
type Defaulter struct {
	r map[string]interface{}
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
