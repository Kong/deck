package sanitize

import (
	"encoding/json"
	"sync"

	"github.com/ettle/strcase"
	"github.com/kong/go-kong/kong"
	"github.com/tidwall/gjson"
)

const (
	typeString = "string"
	typeSet    = "set"
	typeArray  = "array"
)

var exemptionsMu sync.Mutex

func shouldSkipSanitization(entityName, fieldName string) bool {
	if entityName != "" {
		exemptionMap, hasEntitySkips := entityLevelExemptedFields[entityName]
		if hasEntitySkips {
			if _, exempt := exemptionMap[fieldName]; exempt {
				return true
			}
		}
	}

	// checking for config-level exemptions
	if _, exempt := configLevelExemptedFields[fieldName]; exempt {
		return true
	}

	// checking for schema-generated exemptions per entity-type
	if entityName != "" && entityLevelExemptedFieldsFromSchema != nil {
		// schemas are stored in snake_case, so we convert the field name
		// to snake_case before checking against the schema exemptions
		schemaField := strcase.ToSnake(fieldName)
		if exemptedFields, exists := entityLevelExemptedFieldsFromSchema[entityName]; exists {
			if _, exempt := exemptedFields[schemaField]; exempt {
				return true
			}
		}
	}

	return false
}

func buildExemptedFieldsFromSchema(entityName string, entitySchema kong.Schema) {
	exemptionsMu.Lock()
	defer exemptionsMu.Unlock()

	if entityLevelExemptedFieldsFromSchema == nil {
		entityLevelExemptedFieldsFromSchema = make(map[string]map[string]bool)
	}

	if _, exists := entityLevelExemptedFieldsFromSchema[entityName]; exists {
		// already built for this entity type
		return
	}

	exemptedFieldsFromSchema := make(map[string]bool)

	jsonb, err := json.Marshal(&entitySchema)
	if err != nil {
		return
	}
	gjsonSchema := gjson.ParseBytes(jsonb)

	schemaFields := gjsonSchema.Get("fields")
	if schemaFields.Type == gjson.Null {
		schemaFields = gjsonSchema.Get("properties")
	}
	findRelevantFieldNamesWithKey(exemptedFieldsFromSchema, schemaFields, "one_of", "elements")
	findRelevantFieldNamesWithKey(exemptedFieldsFromSchema, schemaFields, "enum", "elements")
	findRelevantFieldNamesWithKey(exemptedFieldsFromSchema, schemaFields, "one_of", "items")
	findRelevantFieldNamesWithKey(exemptedFieldsFromSchema, schemaFields, "enum", "items")

	entityLevelExemptedFieldsFromSchema[entityName] = exemptedFieldsFromSchema
}

// findRelevantFieldNamesWithKey finds all field names (at any depth) where:
// - type is "string" and targetKey is present
// - type is "set" or "array" and the elements contains targetKey
// The identified field names are added to the exemptedFieldsFromSchema map.
// This is useful for identifying fields that should be exempted from sanitization based on the schema
// The arrayKeyName is used to identify the key name for array elements in the schema.
// If targetKey is found in the array, we don't wish to include the arrayKeyName in the exempted fields.
func findRelevantFieldNamesWithKey(exemptedFieldsFromSchema map[string]bool,
	schemaFields gjson.Result, targetKey string, arrayKeyName string,
) {
	var walkSchema func(gjson.Result, bool)
	walkSchema = func(res gjson.Result, skipElements bool) {
		if res.IsObject() {
			for key, value := range res.Map() {
				fieldType := value.Get("type").String()
				if fieldType == typeSet || fieldType == typeArray {
					elements := value.Get(arrayKeyName)
					if elements.Exists() && elements.Get(targetKey).Exists() {
						exemptedFieldsFromSchema[key] = true
						// Skip walking into 'array elements' subtree
						continue
					}
				}
				if !skipElements && fieldType == typeString && value.Get(targetKey).Exists() {
					exemptedFieldsFromSchema[key] = true
				}
				// Only skip walking into 'array elements' if parent is set and we've already added it
				if !(skipElements && key == arrayKeyName) {
					walkSchema(value, skipElements)
				}
			}
		} else if res.IsArray() {
			for _, value := range res.Array() {
				walkSchema(value, skipElements)
			}
		}
	}
	walkSchema(schemaFields, false)
}
