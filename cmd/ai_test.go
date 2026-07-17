package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Kong/ai-deck-converter/convert"
	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
)

// aiGatewaySourceYAML is a minimal but complete AI Gateway 2.0 source document
// used to exercise the JSON/YAML input and output paths.
const aiGatewaySourceYAML = `models:
  - type: model
    display_name: gpt-4o
    name: gpt-4o
    capabilities: [generate]
    formats:
      - type: openai
    targets:
      - name: gpt-4o
        weight: 100
        provider: openai-main
        config:
          type: openai
          temperature: 0.7
    policies: []
    acls: {allow: [], deny: []}
    config:
      route:
        paths: [/ai]
      model:
        alias: "@openai/gpt-4o"
      balancer:
        algorithm: round-robin
        slots: 10000
providers:
  - type: openai
    display_name: OpenAI Main
    name: openai-main
    config:
      auth:
        type: basic
        headers:
          - name: Authorization
            value: "{vault://env/openai-key}"
`

// assertHasAITag verifies the decoded decK document carries the AI-managed
// select tag in its _info section.
func assertHasAITag(t *testing.T, doc map[string]interface{}) {
	t.Helper()
	info, ok := doc["_info"].(map[string]interface{})
	require.True(t, ok, "_info section should be present")
	tags, ok := info["select_tags"].([]interface{})
	require.True(t, ok, "_info.select_tags should be present")
	assert.Contains(t, tags, managedByAIDeckTag)
}

func TestContentHasmanagedByAIDeckTag(t *testing.T) {
	otherTag := "team-a"
	tests := []struct {
		name    string
		content *file.Content
		want    bool
	}{
		{
			name:    "nil content",
			content: nil,
			want:    false,
		},
		{
			name:    "empty content",
			content: &file.Content{},
			want:    false,
		},
		{
			name: "tag in _info select tags",
			content: &file.Content{
				Info: &file.Info{SelectorTags: []string{otherTag, managedByAIDeckTag}},
			},
			want: true,
		},
		{
			name: "unrelated select tag only",
			content: &file.Content{
				Info: &file.Info{SelectorTags: []string{otherTag}},
			},
			want: false,
		},
		{
			name: "tag on a top-level entity",
			content: &file.Content{
				Services: []file.FService{
					{Service: kong.Service{
						Name: kong.String("s1"),
						Tags: kong.StringSlice(otherTag, managedByAIDeckTag),
					}},
				},
			},
			want: true,
		},
		{
			name: "tag on a nested route under a service",
			content: &file.Content{
				Services: []file.FService{
					{
						Service: kong.Service{Name: kong.String("s1")},
						Routes: []*file.FRoute{
							{Route: kong.Route{
								Name: kong.String("r1"),
								Tags: kong.StringSlice(managedByAIDeckTag),
							}},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "tag on a plugin nested under a route under a service",
			content: &file.Content{
				Services: []file.FService{
					{
						Service: kong.Service{Name: kong.String("s1")},
						Routes: []*file.FRoute{
							{
								Route: kong.Route{Name: kong.String("r1")},
								Plugins: []*file.FPlugin{
									{Plugin: kong.Plugin{
										Name: kong.String("p1"),
										Tags: kong.StringSlice(managedByAIDeckTag),
									}},
								},
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "tag on a plugin nested under a service",
			content: &file.Content{
				Services: []file.FService{
					{
						Service: kong.Service{Name: kong.String("s1")},
						Plugins: []*file.FPlugin{
							{Plugin: kong.Plugin{
								Name: kong.String("p1"),
								Tags: kong.StringSlice(managedByAIDeckTag),
							}},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "tag on a consumer nested under a consumer group",
			content: &file.Content{
				ConsumerGroups: []file.FConsumerGroupObject{
					{
						ConsumerGroup: kong.ConsumerGroup{Name: kong.String("cg1")},
						Consumers: []*kong.Consumer{
							{
								Username: kong.String("c1"),
								Tags:     kong.StringSlice(managedByAIDeckTag),
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "tag on a plugin nested under a consumer group",
			content: &file.Content{
				ConsumerGroups: []file.FConsumerGroupObject{
					{
						ConsumerGroup: kong.ConsumerGroup{Name: kong.String("cg1")},
						Plugins: []*kong.ConsumerGroupPlugin{
							{
								Name: kong.String("p1"),
								Tags: kong.StringSlice(managedByAIDeckTag),
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "nested entity with multiple tags including the ai tag",
			content: &file.Content{
				Services: []file.FService{
					{
						Service: kong.Service{Name: kong.String("s1")},
						Routes: []*file.FRoute{
							{Route: kong.Route{
								Name: kong.String("r1"),
								Tags: kong.StringSlice(otherTag, managedByAIDeckTag, "team-b"),
							}},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "nested entity with multiple tags but none is the ai tag",
			content: &file.Content{
				Services: []file.FService{
					{
						Service: kong.Service{Name: kong.String("s1")},
						Routes: []*file.FRoute{
							{Route: kong.Route{
								Name: kong.String("r1"),
								Tags: kong.StringSlice(otherTag, "team-b"),
							}},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "entities without the tag",
			content: &file.Content{
				Services: []file.FService{
					{Service: kong.Service{Name: kong.String("s1"), Tags: kong.StringSlice(otherTag)}},
				},
				Plugins: []file.FPlugin{
					{Plugin: kong.Plugin{Name: kong.String("p1")}},
				},
			},
			want: false,
		},
		{
			name: "matching string in a non-tag field is not a false positive",
			content: &file.Content{
				Services: []file.FService{
					{Service: kong.Service{Name: kong.String(managedByAIDeckTag)}},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, contentHasmanagedByAIDeckTag(tt.content))
		})
	}
}

// TestAiConvertJSONInputParity verifies that a JSON source produces exactly the
// same converted decK output as the equivalent YAML source. This covers JSON
// input support for both `file ai2kong` and `ai sync`, which share convert.Convert.
func TestAiConvertJSONInputParity(t *testing.T) {
	jsonSource, err := yaml.YAMLToJSON([]byte(aiGatewaySourceYAML))
	require.NoError(t, err)

	fromYAML, _, err := convert.Convert([]byte(aiGatewaySourceYAML), convert.Options{OutputMode: "deck"})
	require.NoError(t, err)
	fromJSON, _, err := convert.Convert(jsonSource, convert.Options{OutputMode: "deck"})
	require.NoError(t, err)

	assert.True(t, bytes.Equal(fromYAML, fromJSON),
		"JSON and YAML inputs should convert to identical decK output")
}

// TestAi2KongOutputFormats drives the ai2kong command end-to-end with JSON input
// and asserts both YAML (default) and JSON output are valid and carry the AI tag,
// and that an unknown format is rejected.
func TestAi2KongOutputFormats(t *testing.T) {
	disableAnalytics = true

	dir := t.TempDir()
	jsonSource, err := yaml.YAMLToJSON([]byte(aiGatewaySourceYAML))
	require.NoError(t, err)
	srcJSON := filepath.Join(dir, "input.json")
	require.NoError(t, os.WriteFile(srcJSON, jsonSource, 0o600))
	srcYAML := filepath.Join(dir, "input.yaml")
	require.NoError(t, os.WriteFile(srcYAML, []byte(aiGatewaySourceYAML), 0o600))

	runAi2Kong := func(t *testing.T, source, output, format string) error {
		t.Helper()
		cmd := newAi2KongCmd()
		cmd.SetArgs([]string{"--source", source, "--output-file", output, "--format", format})
		return cmd.Execute()
	}

	t.Run("json input, json output", func(t *testing.T) {
		out := filepath.Join(dir, "out.json")
		require.NoError(t, runAi2Kong(t, srcJSON, out, "json"))
		data, err := os.ReadFile(out)
		require.NoError(t, err)
		var doc map[string]interface{}
		require.NoError(t, json.Unmarshal(data, &doc), "output should be valid JSON")
		assertHasAITag(t, doc)
	})

	t.Run("yaml input, yaml output", func(t *testing.T) {
		out := filepath.Join(dir, "out.yaml")
		require.NoError(t, runAi2Kong(t, srcYAML, out, "yaml"))
		data, err := os.ReadFile(out)
		require.NoError(t, err)
		var doc map[string]interface{}
		require.NoError(t, yaml.Unmarshal(data, &doc))
		assertHasAITag(t, doc)
	})

	t.Run("unknown format errors", func(t *testing.T) {
		err := runAi2Kong(t, srcYAML, filepath.Join(dir, "out.xml"), "xml")
		assert.Error(t, err)
	})
}
