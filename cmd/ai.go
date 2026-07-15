package cmd

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/spf13/cobra"
)

// aiGatewayDocsURL documents the supported way to manage AI Gateway
// configuration on Konnect.
const aiGatewayDocsURL = "https://developer.konghq.com/ai-gateway/kongctl"

func newAiSubCmd() *cobra.Command {
	aiSubCmd := &cobra.Command{
		Use:   "ai [sub-command]...",
		Short: "Subcommand to host decK AI Gateway operations",
		Long:  `Subcommand to host decK AI Gateway operations.`,
	}

	return aiSubCmd
}

// errAIManagedEntitiesOnKonnect is returned when a user attempts to sync/apply
// AI Gateway entities (tagged managed_by:deck-ai) to Konnect using decK. Those
// entities must be managed with kongctl instead.
func errAIManagedEntitiesOnKonnect() error {
	return fmt.Errorf(
		"AI Gateway entities (tagged %q) cannot be managed on Konnect using decK.\n"+
			"Use kongctl to manage AI Gateway configuration on Konnect instead: %s",
		managedByAIDeckTag, aiGatewayDocsURL)
}

// contentHasmanagedByAIDeckTag reports whether the configuration is marked as managed
// by the AI Gateway tooling, either via a select tag in _info or via the tags
// of any individual entity.
func contentHasmanagedByAIDeckTag(content *file.Content) bool {
	if content == nil {
		return false
	}
	if content.Info != nil && slices.Contains(content.Info.SelectorTags, managedByAIDeckTag) {
		return true
	}
	return valueHasTag(reflect.ValueOf(content))
}

// valueHasTag recursively walks v looking for any exported field named "Tags"
// (a []*string, as used by every Kong entity) that contains managedByAIDeckTag.
// Non-Tags fields are traversed structurally but their string values are never
// matched, so plugin config values or names cannot trigger a false positive.
func valueHasTag(v reflect.Value) bool {
	switch v.Kind() { //nolint:exhaustive // only container and struct kinds need traversal; scalars can never hold Tags
	case reflect.Pointer, reflect.Interface:
		if v.IsNil() {
			return false
		}
		return valueHasTag(v.Elem())
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if valueHasTag(v.Index(i)) {
				return true
			}
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			if valueHasTag(v.MapIndex(key)) {
				return true
			}
		}
	case reflect.Struct:
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			if !field.IsExported() {
				continue
			}
			fieldValue := v.Field(i)
			if field.Name == "Tags" {
				if tagsContain(fieldValue) {
					return true
				}
				continue
			}
			if valueHasTag(fieldValue) {
				return true
			}
		}
	}
	return false
}

// tagsContain reports whether a []*string Tags field contains managedByAIDeckTag.
func tagsContain(tags reflect.Value) bool {
	if tags.Kind() != reflect.Slice {
		return false
	}
	for i := 0; i < tags.Len(); i++ {
		tag := tags.Index(i)
		if tag.Kind() == reflect.Pointer {
			if tag.IsNil() {
				continue
			}
			tag = tag.Elem()
		}
		if tag.Kind() == reflect.String && tag.String() == managedByAIDeckTag {
			return true
		}
	}
	return false
}
