package sanitize

import (
	"context"
	"fmt"
	"reflect"

	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-database-reconciler/pkg/utils"
	"github.com/kong/go-kong/kong"
)

type Sanitizer struct {
	ctx          context.Context
	client       *kong.Client
	content      *file.Content
	isKonnect    bool
	salt         string
	sanitizedMap map[string]interface{}
}

type SanitizerOptions struct {
	Ctx       context.Context
	Client    *kong.Client
	Content   *file.Content
	Salt      string
	IsKonnect bool
}

func NewSanitizer(opts *SanitizerOptions) *Sanitizer {
	saltToUse := opts.Salt
	if saltToUse == "" {
		saltToUse = utils.UUID()
	}

	return &Sanitizer{
		ctx:          opts.Ctx,
		client:       opts.Client,
		content:      opts.Content,
		isKonnect:    opts.IsKonnect,
		salt:         saltToUse,
		sanitizedMap: make(map[string]interface{}),
	}
}

func (s *Sanitizer) Sanitize() (*file.Content, error) {
	content := reflect.ValueOf(s.content)
	if content.Kind() == reflect.Ptr {
		content = content.Elem()
	}
	contentType := content.Type()

	for i := 0; i < content.NumField(); i++ {
		fieldName := contentType.Field(i).Name
		if _, exempt := topLevelExemptedFields[fieldName]; exempt {
			continue
		}
		fieldValueSet := content.Field(i)

		if fieldValueSet.IsValid() && fieldValueSet.CanInterface() && !fieldValueSet.IsZero() {
			s.sanitizeField(fieldValueSet)
		}
	}

	return s.content, nil
}

func (s *Sanitizer) sanitizeField(field reflect.Value) {
	if !field.IsValid() {
		return
	}

	//nolint:exhaustive
	switch field.Kind() {
	case reflect.Ptr:
		if field.IsNil() {
			return
		}
		s.sanitizeField(field.Elem())
	case reflect.Struct:
		t := field.Type()
		entityName := t.Name()

		entitySkipFields, hasEntitySkips := entityLevelExemptedFields[entityName]
		for i := 0; i < field.NumField(); i++ {
			fieldValue := field.Field(i)
			fieldName := t.Field(i).Name

			if hasEntitySkips && s.shouldSkipSanitization(fieldName, entitySkipFields) {
				continue
			}

			if s.shouldSkipSanitization(fieldName, nil) {
				continue
			}

			// needs special handling for configs as they are not pointers to structs
			// commented for now to satisfy linter
			// if fieldValue.Type() == reflect.TypeOf(kong.Configuration{}) {
			// 	// handle configs - tba
			// }

			s.sanitizeField(fieldValue)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < field.Len(); i++ {
			s.sanitizeField(field.Index(i))
		}
	case reflect.Map:
		iter := field.MapRange()
		for iter.Next() {
			mapKey := iter.Key().String()
			if s.shouldSkipSanitization(mapKey, nil) {
				continue
			}
			mapValue := iter.Value()
			if mapValue.Kind() == reflect.Ptr {
				mapValue = mapValue.Elem()
			}
			s.sanitizeField(mapValue)
		}
	case reflect.Interface:
		if field.IsNil() {
			return
		}
		s.sanitizeField(field.Elem())
	case reflect.String:
		originalValue := field.String()
		sanitizedValue, exists := s.sanitizedMap[originalValue]
		if !exists {
			sanitizedValue = s.sanitizeValue(originalValue)
			s.sanitizedMap[originalValue] = sanitizedValue
		}

		if field.CanSet() {
			field.SetString(sanitizedValue.(string))
		} else {
			fmt.Println("Cannot set sanitized value for field:", field.Type(), "with value:", originalValue)
		}
	default:
		// No operation needed for other kinds
	}
}
