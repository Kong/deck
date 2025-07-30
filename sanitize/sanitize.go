package sanitize

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"

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

			if hasEntitySkips && shouldSkipSanitization(fieldName, entitySkipFields) {
				continue
			}

			if shouldSkipSanitization(fieldName, nil) {
				continue
			}

			// needs special handling for configs as they are not pointers to structs
			// commented for now to satisfy linter
			// if fieldValue.Type() == reflect.TypeOf(kong.Configuration{}) {
			// 	// handle configs - tba
			// }
			if fieldValue.Type() == reflect.TypeOf(kong.Configuration{}) {
				// handle configs - tba
			}

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
			if shouldSkipSanitization(mapKey, nil) {
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
		if field.CanSet() {
			field.SetString(s.sanitizeValue(originalValue))
		} else {
			fmt.Println("Cannot set sanitized value for field:", field.Type(), "with value:", originalValue)
		}
	default:
		// No operation needed for other kinds
	}
}

func (s *Sanitizer) sanitizeValue(value string) string {
	sanitizedValue, exists := s.sanitizedMap[value]
	if exists {
		if str, ok := sanitizedValue.(string); ok {
			return str
		}
	}

	hashedValue := s.hashValue(value)
	if !strings.Contains(value, "/") {
		s.sanitizedMap[value] = hashedValue
		return hashedValue
	}

	var redactedPath string

	// this means it is a path field
	if strings.HasPrefix(value, "/") || strings.HasPrefix(value, "~/") {
		redactedPath = "/redacted/path/"
	} else {
		// if it is not a path, but a string with a /, it could be a content-type
		redactedPath = "redacted/"
	}

	hashedPath := redactedPath + hashedValue

	s.sanitizedMap[value] = hashedPath
	return hashedPath
}

func (s *Sanitizer) hashValue(value string) string {
	hasher := sha256.New()
	hasher.Write([]byte(s.salt + value))
	return hex.EncodeToString(hasher.Sum(nil))
}
