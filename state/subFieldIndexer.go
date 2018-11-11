package state

import (
	"fmt"
	"reflect"
)

// SubFieldIndexer is used to extract a field from an object
// using reflection and builds an index on that field.
type SubFieldIndexer struct {
	StructField string
	SubField    string
}

func (s *SubFieldIndexer) FromObject(obj interface{}) (bool, []byte, error) {
	v := reflect.ValueOf(obj)
	v = reflect.Indirect(v) // Dereference the pointer if any

	structV := v.FieldByName(s.StructField)
	structV = reflect.Indirect(structV)
	if !structV.IsValid() {
		return false, nil,
			fmt.Errorf("field '%s' for %#v is invalid", s.StructField, obj)
	}
	subField := structV.FieldByName(s.SubField)
	subField = reflect.Indirect(subField)

	val := subField.String()
	if val == "" {
		return false, nil, nil
	}

	// Add the null character as a terminator
	val += "\x00"
	return true, []byte(val), nil
}

func (s *SubFieldIndexer) FromArgs(args ...interface{}) ([]byte, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("must provide only a single argument")
	}
	arg, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("argument must be a string: %#v", args[0])
	}
	// Add the null character as a terminator
	arg += "\x00"
	return []byte(arg), nil
}
