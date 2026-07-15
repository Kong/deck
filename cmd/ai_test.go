package cmd

import (
	"testing"

	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

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
