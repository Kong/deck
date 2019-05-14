package utils

import (
	"reflect"

	"github.com/pkg/errors"
)

// MergeTags merges Tags in the object with tags.
func MergeTags(obj interface{}, tags []string) error {
	if len(tags) == 0 {
		return nil
	}
	ptr := reflect.ValueOf(obj)
	if ptr.Kind() != reflect.Ptr {
		return errors.New("obj is not a pointer")
	}
	v := reflect.Indirect(ptr)
	structTags := v.FieldByName("Tags")
	var zero reflect.Value
	if structTags == zero {
		return nil
	}
	m := make(map[string]bool)
	for i := 0; i < structTags.Len(); i++ {
		tag := reflect.Indirect(structTags.Index(i)).String()
		m[tag] = true
	}
	for _, tag := range tags {
		if _, ok := m[tag]; !ok {
			t := tag
			structTags.Set(reflect.Append(structTags, reflect.ValueOf(&t)))
		}
	}
	return nil
}

// RemoveTags removes tags from the Tags in obj.
func RemoveTags(obj interface{}, tags []string) error {
	if len(tags) == 0 {
		return nil
	}

	m := make(map[string]bool)
	for _, tag := range tags {
		m[tag] = true
	}

	ptr := reflect.ValueOf(obj)
	if ptr.Kind() != reflect.Ptr {
		return errors.New("obj is not a pointer")
	}
	v := reflect.Indirect(ptr)
	structTags := v.FieldByName("Tags")
	var zero reflect.Value
	if structTags == zero {
		return nil
	}

	res := reflect.MakeSlice(reflect.SliceOf(reflect.PtrTo(reflect.TypeOf(""))), 0, 0)
	for i := 0; i < structTags.Len(); i++ {
		tag := reflect.Indirect(structTags.Index(i)).String()
		if !m[tag] {
			res = reflect.Append(res, structTags.Index(i))
		}
	}
	structTags.Set(res)
	return nil
}
