package utils

import "reflect"

var zero reflect.Value

func zeroOutField(obj interface{}, field string) {
	ptr := reflect.ValueOf(obj)
	if ptr.Kind() != reflect.Ptr {
		return
	}
	v := reflect.Indirect(ptr)
	ts := v.FieldByName(field)
	if ts == zero {
		return
	}
	ts.Set(reflect.Zero(ts.Type()))
}

func ZeroOutID(obj interface{}, altName *string, withID bool) {
	// withID is set, export the ID
	if withID {
		return
	}
	// altName is not set, export the ID
	if Empty(altName) {
		return
	}
	// zero the ID field
	zeroOutField(obj, "ID")
}

func ZeroOutTimestamps(obj interface{}) {
	zeroOutField(obj, "CreatedAt")
	zeroOutField(obj, "UpdatedAt")
}
