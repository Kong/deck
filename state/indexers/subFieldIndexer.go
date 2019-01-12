package indexers

import (
	"fmt"
	"reflect"
)

// Field represents a field that needs to be used for
// subfield indexing.
type Field struct {
	// Struct is the name of the field of the struct
	// being indexed.
	Struct string
	// Sub is the name of the field inside the struct Struct,
	// which is being indexed.
	Sub string
}

// SubFieldIndexer is used to extract a field from an object
// using reflection and builds an index on that field.
type SubFieldIndexer struct {
	Fields []Field
}

// FromObject take Obj and returns index key formed using
// the field SubField.
func (s *SubFieldIndexer) FromObject(obj interface{}) (bool, []byte, error) {
	v := reflect.ValueOf(obj)
	v = reflect.Indirect(v) // Dereference the pointer if any

	val := ""
	for _, f := range s.Fields {
		structV := v.FieldByName(f.Struct)
		structV = reflect.Indirect(structV)
		if !structV.IsValid() {
			continue
		}
		subField := structV.FieldByName(f.Sub)
		subField = reflect.Indirect(subField)

		val += subField.String()
	}

	if val == "" {
		return false, nil, nil
	}

	// Add the null character as a terminator
	val += "\x00"
	return true, []byte(val), nil
}

// FromArgs takes in a string and returns its byte form.
func (s *SubFieldIndexer) FromArgs(args ...interface{}) ([]byte, error) {
	val := ""
	for _, arg := range args {
		s, ok := arg.(string)
		if !ok {
			return nil, fmt.Errorf("argument must be a string: %#v", args[0])
		}
		val += s
	}
	// Add the null character as a terminator
	val += "\x00"
	return []byte(val), nil
}
