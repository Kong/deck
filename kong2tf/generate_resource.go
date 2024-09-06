package kong2tf

import (
	"fmt"
	"sort"
	"strings"
)

type importConfig struct {
	controlPlaneID *string
	importValues   map[string]*string
}

func generateResource(
	entityType,
	name string,
	entity map[string]any,
	parents map[string]string,
	imports importConfig,
	lifecycle []string,
) string {
	return generateResourceWithCustomizations(entityType, name, entity, parents, map[string]string{}, imports, lifecycle)
}

func generateResourceWithCustomizations(
	entityType,
	name string,
	entity map[string]any,
	parents map[string]string,
	customizations map[string]string,
	imports importConfig,
	lifecycle []string,
) string {
	// Cache ID in case we need to use it for imports
	entityID := ""
	if entity["id"] != nil {
		entityID = entity["id"].(string)
	}

	// Populate parents with foreign keys as needed
	parentKeys := []string{"service", "route", "consumer", "upstream", "certificate", "consumer_group"}

	// Populate parents with foreign keys as needed
	for _, key := range parentKeys {
		if entity[key] != nil {
			// Switch on type of parent
			switch entity[key].(type) {
			case string:
				parents[key] = entity[key].(string)
			case map[string]interface{}:
				parents[key] = entity[key].(map[string]interface{})["name"].(string)
			default:
				panic(fmt.Sprintf("Unknown type for parent %s", key))
			}
		}
	}

	// List of keys to remove
	removeKeys := []string{
		"id",
	}

	// Build a map of entity types to keys
	entityTypeToKeys := map[string][]string{
		"gateway_service": {"routes", "plugins"},
		"gateway_route":   {"plugins", "service"},
		"gateway_plugin":  {"service", "route", "consumer"},
		"gateway_consumer": {
			"groups", "acls", "basicauth_credentials", "keyauth_credentials",
			"jwt_secrets", "hmacauth_credentials", "basicauth_credentials", "plugins",
		},
		"gateway_upstream":       {"targets"},
		"gateway_consumer_group": {"consumers", "plugins"},
		"gateway_certificate":    {"snis"},
	}

	if additionalKeys := entityTypeToKeys[entityType]; additionalKeys != nil {
		removeKeys = append(removeKeys, additionalKeys...)
	}

	// Remove keys that are not needed
	for _, k := range removeKeys {
		delete(entity, k)
	}

	if entityType == "gateway_plugin" {
		entityType = fmt.Sprintf("%s_%s", entityType, name)
		delete(entity, "name")
	}

	// We don't need to prefix SNIs with the Cert name
	// Or routes with the service name
	if entityType != "gateway_sni" && entityType != "gateway_route" {
		for k := range parents {
			name = fmt.Sprintf("%s_%s", strings.ReplaceAll(parents[k], "-", "_"), name)
		}
	}

	s := fmt.Sprintf(`
resource "konnect_%s" "%s" {
%s

%s  control_plane_id = var.control_plane_id%s
}
`,
		entityType, name,
		strings.TrimRight(output(entityType, entity, 1, true, "\n", customizations), "\n"),
		generateParents(parents),
		generateLifecycle(lifecycle))

	// Generate imports
	if imports.controlPlaneID != nil && entityID != "" {
		entity["id"] = entityID
		s += generateImports(entityType, name, imports.importValues, imports.controlPlaneID)
	}

	return strings.TrimSpace(s) + "\n\n"
}

func generateRelationship(
	entityType string,
	name string,
	relations map[string]string,
	_ map[string]any, // 'entity' when TODO is resolved
	_ importConfig, // 'imports' when TODO is resolved
) string {
	// TODO: We don't support relationship importing in the provider yet
	// entityID := entity["id"].(string)

	s := fmt.Sprintf(`resource "konnect_%s" "%s" {`, entityType, name)

	// Extract keys to iterate in a deterministic order
	keys := make([]string, 0)
	for k := range relations {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	// Output each item in the relationship
	for _, k := range keys {
		s += fmt.Sprintf("\n"+`  %s_id = konnect_gateway_%s.%s.id`, k, k, relations[k])
	}
	s += "\n  control_plane_id = var.control_plane_id"
	s += "\n}\n\n"

	// TODO: We don't support relationship importing in the provider yet
	/*
		│ Error: Not Implemented
		│
		│ No available import state operation is available for resource gateway_consumer_group_member.
	*/
	//if imports.controlPlaneID != nil {
	//	entity["id"] = entityID
	//	s += generateImports(entityType, name, entity, imports.importValues, imports.controlPlaneID) + "\n\n"
	//}

	return s
}

func generateImports(
	entityType string,
	name string,
	keysFromEntity map[string]*string,
	cpID *string,
) string {
	if len(keysFromEntity) == 0 {
		return ""
	}

	return fmt.Sprintf("\n"+`import {
  to = konnect_%s.%s
  id = "%s"
}`, entityType, name, generateImportKeys(keysFromEntity, cpID))
}

func generateImportKeys(keys map[string]*string, cpID *string) string {
	if len(keys) == 0 {
		return ""
	}

	s := "{"
	for k, val := range keys {
		s += fmt.Sprintf(`\"%s\": \"%s\", `, k, *val)
	}

	s += fmt.Sprintf(`\"control_plane_id\": \"%s\", `, *cpID)

	s = strings.TrimRight(s, ", ")

	s += "}"

	return s
}

func generateLifecycle(lifecycle []string) string {
	if len(lifecycle) == 0 {
		return ""
	}

	s := `
  lifecycle {
    ignore_changes = [`
	for _, l := range lifecycle {
		s += "\n      " + l + ","
	}
	s = strings.TrimRight(s, ",")

	s += `
    ]
  }
`

	return s
}

func generateParents(parents map[string]string) string {
	if len(parents) == 0 {
		return ""
	}

	var result []string
	for k, v := range parents {
		v = strings.ReplaceAll(v, "-", "_")
		// if parent ends with _id, use it as-is
		if strings.HasSuffix(k, "_id") {
			result = append(result, fmt.Sprintf(`  %s = konnect_gateway_%s.%s.id`, k, strings.TrimSuffix(k, "_id"), v)+"\n")
			continue
		}
		result = append(result, fmt.Sprintf(`  %s = {
    id = konnect_gateway_%s.%s.id
  }`+"\n", k, k, v))
	}

	return strings.Join(result, "\n") + "\n"
}

// Output function that handles the dynamic data
func output(
	entityType string,
	object map[string]interface{},
	depth int,
	isRoot bool,
	eol string,
	customizations map[string]string,
) string {
	var result []string

	// Loop through object in order of keys
	keys := make([]string, 0)
	for k := range object {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	// Move the most common keys to the front
	var prioritizedKeys []string
	for _, k := range []string{"enabled", "name", "username"} {
		if _, exists := object[k]; exists {
			prioritizedKeys = append(prioritizedKeys, k)
		}
	}

	// Append the rest of the keys
	for _, k := range keys {
		if contains(prioritizedKeys, k) {
			continue
		}
		if k != "name" && k != "enabled" {
			prioritizedKeys = append(prioritizedKeys, k)
		}
	}
	keys = prioritizedKeys

	for _, k := range keys {
		v := object[k]

		// TODO: Remove this once deck dump doesn't export nil values
		if v == nil {
			continue
		}

		switch v := v.(type) {
		case map[string]interface{}:
			result = append(result, outputHash(entityType, k, v, depth, isRoot, eol, customizations))
		case []interface{}:
			result = append(result, outputList(entityType, k, v, depth))
		default:
			result = append(result, line(fmt.Sprintf("%s = %s", k, quote(v)), depth, eol))
		}
	}
	return strings.Join(result, "")
}

// Handles rendering a map (hash) in Go
func outputHash(
	entityType string,
	key string,
	input map[string]interface{},
	depth int,
	isRoot bool,
	eol string,
	customizations map[string]string,
) string {
	s := ""
	if !isRoot {
		s += "\n"
	}

	custom := customizations[key]

	if custom != "" {
		s += line(fmt.Sprintf("%s = %s({", key, custom), depth, eol)
	} else {
		s += line(fmt.Sprintf("%s = {", key), depth, eol)
	}

	s += output(entityType, input, depth+1, true, eol, customizations)

	if custom != "" {
		s += line("})", depth, eol)
	} else {
		s += line("}", depth, eol)
	}
	return s
}

// Handles rendering a map within a list in Go
func outputHashInList(entityType string, input map[string]interface{}, depth int) string {
	s := "\n"
	s += line("{", depth+1, "\n")
	s += output(entityType, input, depth+2, false, "\n", map[string]string{})
	s += line("},", depth+1, "\n")
	return s
}

// Handles rendering a list (array) in Go
func outputList(entityType string, key string, input []interface{}, depth int) string {
	s := line(fmt.Sprintf("%s = [", key), depth, "")
	for _, v := range input {
		switch v := v.(type) {
		case map[string]interface{}:
			s += outputHashInList(entityType, v, depth)
		default:
			s += fmt.Sprintf("%s, ", quote(v))
		}
	}
	s = strings.TrimRight(s, ", ")
	s += endList(input, depth)
	return s
}

// Ends a list rendering in Go
func endList(input []interface{}, depth int) string {
	lastLine := line("]", depth, "\n")
	if _, ok := input[len(input)-1].(map[string]interface{}); ok {
		return lastLine
	}
	return strings.TrimLeft(lastLine, " ")
}

// Formats a line with proper indentation and end-of-line characters
func line(input string, depth int, eol string) string {
	return strings.Repeat("  ", depth) + input + eol
}

// Properly quotes a value based on its type
func quote(input interface{}) string {
	switch v := input.(type) {
	case nil:
		return ""
	case bool, int, float64:
		return fmt.Sprintf("%v", v)
	case string:
		if strings.Contains(v, "\n") {
			return fmt.Sprintf("<<EOF\n%s\nEOF", strings.TrimRight(v, "\n"))
		}
		return fmt.Sprintf("\"%s\"", strings.ReplaceAll(v, "\"", "\\\""))
	default:
		return fmt.Sprintf("\"%v\"", v)
	}
}

// contains checks if a slice contains a specific element
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
