package sanitize

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"

	"github.com/kong/go-database-reconciler/pkg/cprint"
	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-database-reconciler/pkg/types"
	"github.com/kong/go-database-reconciler/pkg/utils"
	"github.com/kong/go-kong/kong"
)

type Sanitizer struct {
	ctx                 context.Context
	client              *kong.Client
	content             *file.Content
	isKonnect           bool
	salt                string
	sanitizedMap        map[string]interface{}
	pluginSchemasCache  *types.SchemaCache
	partialSchemasCache *types.SchemaCache
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
		pluginSchemasCache: types.NewSchemaCache(func(ctx context.Context,
			pluginName string,
		) (map[string]interface{}, error) {
			return opts.Client.Plugins.GetFullSchema(ctx, &pluginName)
		}),
		partialSchemasCache: types.NewSchemaCache(func(ctx context.Context,
			partialType string,
		) (map[string]interface{}, error) {
			return opts.Client.Partials.GetFullSchema(ctx, &partialType)
		}),
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

		// specificEntityName is used to identify plugin or partial types
		// it is useful for building exempted fields
		specificEntityName, entitySchema, err := s.fetchEntitySchema(entityName, field)
		if err != nil {
			warningMessage := fmt.Sprintf("Error fetching schema for entity: %s %v\n"+
				"Some sanitization features may not work as expected.\n", entityName, err)
			cprint.UpdatePrintlnStdErr(warningMessage)
		}

		if entitySchema != nil {
			buildExemptedFieldsFromSchema(specificEntityName, entitySchema)
		}

		entitySkipFields, hasEntitySkips := entityLevelExemptedFields[entityName]
		for i := 0; i < field.NumField(); i++ {
			fieldValue := field.Field(i)
			fieldName := t.Field(i).Name

			if hasEntitySkips && shouldSkipSanitization(specificEntityName, fieldName, entitySkipFields) {
				continue
			}

			if shouldSkipSanitization(specificEntityName, fieldName, nil) {
				continue
			}

			// needs special handling for configs as they are not pointers to structs
			if fieldValue.Type() == reflect.TypeOf(kong.Configuration{}) {
				sanitizedConfig := s.sanitizeConfig(fieldValue)
				if fieldValue.CanSet() {
					fieldValue.Set(reflect.ValueOf(sanitizedConfig))
				} else {
					fmt.Println("Cannot set sanitized config for field:", field.Type())
				}
				continue
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
			if shouldSkipSanitization("", mapKey, nil) {
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

func (s *Sanitizer) sanitizeConfig(config reflect.Value) interface{} {
	sanitizedConfig := make(map[string]interface{})

	//nolint:exhaustive
	switch config.Kind() {
	case reflect.Invalid:
		return nil
	case reflect.String:
		configValue := config.String()
		if configValue == "" {
			return configValue
		}
		return s.sanitizeValue(configValue)
	case reflect.Map:
		for _, key := range config.MapKeys() {
			val := config.MapIndex(key)
			if !val.IsValid() {
				continue
			}
			k, ok := key.Interface().(string)
			if !ok {
				continue
			}

			if shouldSkipSanitization("", k, nil) {
				sanitizedConfig[k] = val.Interface()
				continue
			}

			switch v := val.Interface().(type) {
			case string:
				sanitizedVal := s.sanitizeValue(v)
				sanitizedConfig[k] = sanitizedVal
			case map[string]interface{}:
				sanitizedConfig[k] = s.sanitizeConfig(reflect.ValueOf(v))
			case []interface{}:
				newSlice := make([]interface{}, len(v))
				for i, elem := range v {
					newSlice[i] = s.sanitizeConfig(reflect.ValueOf(elem))
				}
				sanitizedConfig[k] = newSlice
			default:
				sanitizedConfig[k] = v
			}
		}
	default:
		if config.CanInterface() {
			return config.Interface()
		}
	}

	return sanitizedConfig
}
