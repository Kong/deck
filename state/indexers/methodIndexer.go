package indexers

import (
	"fmt"
	"reflect"
)

// MethodIndexer is used to create an index based on a string returned
// as a result of calling a method on the object.
// It is assumed that the method has no arguments.
type MethodIndexer struct {
	// Method name to call to get the string to bulid the index on.
	Method string
}

// FromObject take Obj and returns index key formed using
// the fields.
func (s *MethodIndexer) FromObject(obj interface{}) (bool, []byte, error) {
	v := reflect.ValueOf(obj)

	method := v.MethodByName(s.Method)
	resp := method.Call(nil)
	if len(resp) != 1 {
		return false, nil, fmt.Errorf("function call returned unexpected result")
	}
	key := resp[0].String()

	if key == "" {
		return false, nil, nil
	}
	return true, []byte(key), nil
}

// FromArgs takes in a string and returns its byte form.
func (s *MethodIndexer) FromArgs(args ...interface{}) ([]byte, error) {
	blob := ""
	for _, arg := range args {
		s, ok := arg.(string)
		if !ok {
			return nil, fmt.Errorf("argument must be a string: %#v", arg)
		}
		blob += s
	}
	if blob == "" {
		return nil, fmt.Errorf("empty args is not a valid value")
	}
	return []byte(blob), nil
}
