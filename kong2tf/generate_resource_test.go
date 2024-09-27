package kong2tf

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateComplexLayout(t *testing.T) {
	jsonInput := `{
		"field1": "basic_string",
		"field2": ["list", "of", "strings"],
		"field3": {
			"nested_field1": "nested_string",
			"nested_field2": ["list", "of", "nested", "strings"],
			"nested_field3": {
				"nested_nested_field1": "nested_nested_string",
				"nested_nested_field2": ["list", "of", "nested", "nested", "strings"],
				"nested_nested_field3": {
					"nested_nested_nested_field1": "nested_nested_nested_string"
				}
			}
		},
		"field4": []
	}`

	var entity map[string]any
	err := json.Unmarshal([]byte(jsonInput), &entity)
	require.NoError(t, err)

	expected := `resource "konnect_entity_type" "some_service_name" {
  field1 = "basic_string"
  field2 = ["list", "of", "strings"]
  field3 = {
    nested_field1 = "nested_string"
    nested_field2 = ["list", "of", "nested", "strings"]
    nested_field3 = {
      nested_nested_field1 = "nested_nested_string"
      nested_nested_field2 = ["list", "of", "nested", "nested", "strings"]
      nested_nested_field3 = {
        nested_nested_nested_field1 = "nested_nested_nested_string"
      }
    }
  }
  field4 = []	

  service = {
    id = konnect_gateway_service.some_service.id
  }
  control_plane_id = var.control_plane_id
}
`

	result := generateResource("entity_type", "name", entity, map[string]string{
		"service": "some_service",
	}, importConfig{
		controlPlaneID: new(string),
		importValues:   map[string]*string{},
	}, []string{})

	require.Equal(t, strings.Fields(expected), strings.Fields(result))
}

func TestGenerateResourceWithoutParent(t *testing.T) {
	entity := map[string]any{
		"field1": "value1",
		"field2": "value2",
	}
	parents := map[string]string{}

	expected := `resource "konnect_entity_type" "name" {
	field1 = "value1"
	field2 = "value2"

	control_plane_id = var.control_plane_id
}
`

	result := generateResource("entity_type", "name", entity, parents, importConfig{
		controlPlaneID: new(string),
		importValues:   map[string]*string{},
	}, []string{})
	require.Equal(t, strings.Fields(expected), strings.Fields(result))
}

func TestGenerateResourceWithParent(t *testing.T) {
	entity := map[string]any{
		"field1": "value1",
		"field2": "value2",
	}
	parents := map[string]string{
		"parent1": "parent_value1",
	}

	expected := `resource "konnect_entity_type" "parent_value1_name" {
	field1 = "value1"
	field2 = "value2"

	parent1 = {
		id = konnect_gateway_parent1.parent_value1.id
	}
	control_plane_id = var.control_plane_id
}
`

	result := generateResource("entity_type", "name", entity, parents, importConfig{
		controlPlaneID: new(string),
		importValues:   map[string]*string{},
	}, []string{})
	require.Equal(t, strings.Fields(expected), strings.Fields(result))
}

func TestGenerateRelationship(t *testing.T) {
	entity := map[string]any{
		"field1": "value1",
		"field2": "value2",
	}

	relations := map[string]string{
		"relation1": "relation_value1",
		"relation2": "relation_value2",
	}

	expected := `resource "konnect_entity_type" "name" {
	relation1_id = konnect_gateway_relation1.relation_value1.id
	relation2_id = konnect_gateway_relation2.relation_value2.id
	control_plane_id = var.control_plane_id
}
`

	result := generateRelationship("entity_type", "name", relations, entity, importConfig{
		controlPlaneID: new(string),
		importValues:   map[string]*string{},
	})
	require.Equal(t, strings.Fields(expected), strings.Fields(result))
}

func TestLifecycle(t *testing.T) {
	entity := map[string]any{
		"field1": "value1",
		"field2": "value2",
	}

	expected := `resource "konnect_entity_type" "name" {
  field1 = "value1"
  field2 = "value2"

  control_plane_id = var.control_plane_id
  lifecycle {
    ignore_changes = [
      ignore_this_one
    ]
  }

}
`

	result := generateResource("entity_type", "name", entity, map[string]string{}, importConfig{
		controlPlaneID: new(string),
		importValues:   map[string]*string{},
	}, []string{
		"ignore_this_one",
	})
	require.Equal(t, strings.Fields(expected), strings.Fields(result))
}

func TestImports(t *testing.T) {
	jsonInput := `{
		"id": "some_id",
		"field1": "basic_string",
		"field2": {
			"nested": "subkey"
		}
	}`

	var entity map[string]any
	err := json.Unmarshal([]byte(jsonInput), &entity)
	require.NoError(t, err)

	expected := `resource "konnect_entity_type" "name" {
  field1 = "basic_string"
  field2 = {
    nested = "subkey"
  }

  control_plane_id = var.control_plane_id
}

import {
  to = konnect_entity_type.name
  id = "{\"id\": \"some_id\", \"field2_nested\": \"subkey\", \"control_plane_id\": \"abc-123\"}"
}`

	cpID := new(string)
	*cpID = "abc-123"

	result := generateResource("entity_type", "name", entity, map[string]string{}, importConfig{
		controlPlaneID: cpID,
		importValues: map[string]*string{
			"id":            func() *string { s := "some_id"; return &s }(),
			"field2_nested": func() *string { s := "subkey"; return &s }(),
		},
	}, []string{})
	require.Equal(t, strings.Fields(expected), strings.Fields(result))
}
