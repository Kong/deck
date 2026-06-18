package cmd

import (
	"context"
	"reflect"
	"testing"

	"github.com/kong/go-database-reconciler/pkg/diff"
	"github.com/kong/go-database-reconciler/pkg/dump"
	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/kong/go-database-reconciler/pkg/state"
	"github.com/kong/go-kong/kong"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetermineSelectorTag(t *testing.T) {
	type args struct {
		dumpConfig  dump.Config
		fileContent file.Content
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "both present and equal",
			args: args{
				dumpConfig:  dump.Config{SelectorTags: []string{"foo"}},
				fileContent: file.Content{Info: &file.Info{SelectorTags: []string{"foo"}}},
			},
			want:    []string{"foo"},
			wantErr: false,
		},
		{
			name: "both present and equal order invariant",
			args: args{
				dumpConfig:  dump.Config{SelectorTags: []string{"foo", "bar"}},
				fileContent: file.Content{Info: &file.Info{SelectorTags: []string{"bar", "foo"}}},
			},
			want:    []string{"bar", "foo"},
			wantErr: false,
		},
		{
			name: "both present and not equal",
			args: args{
				dumpConfig:  dump.Config{SelectorTags: []string{"bar"}},
				fileContent: file.Content{Info: &file.Info{SelectorTags: []string{"foo"}}},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "only file has tags",
			args: args{
				dumpConfig:  dump.Config{},
				fileContent: file.Content{Info: &file.Info{SelectorTags: []string{"foo"}}},
			},
			want:    []string{"foo"},
			wantErr: false,
		},
		{
			name: "only config has tags",
			args: args{
				dumpConfig:  dump.Config{SelectorTags: []string{"foo"}},
				fileContent: file.Content{Info: &file.Info{}},
			},
			want:    []string{"foo"},
			wantErr: false,
		},
		{
			name: "config has one tag and file has duplicates",
			args: args{
				dumpConfig:  dump.Config{SelectorTags: []string{"foo"}},
				fileContent: file.Content{Info: &file.Info{SelectorTags: []string{"foo", "foo"}}},
			},
			want:    []string{"foo"},
			wantErr: false,
		},
		{
			name: "config has multiple tags and file has duplicates",
			args: args{
				dumpConfig:  dump.Config{SelectorTags: []string{"foo", "bar"}},
				fileContent: file.Content{Info: &file.Info{SelectorTags: []string{"foo", "bar", "foo", "bar"}}},
			},
			want:    []string{"bar", "foo"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := determineSelectorTag(tt.args.fileContent, tt.args.dumpConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("determineSelectorTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("determineSelectorTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPerformDiff_JSONOutput(t *testing.T) {
	// Reset global jsonOutput to a known state
	jsonOutput = diff.JSONOutputObject{}
	// This is initialized in syncMain() in the actual application,
	// but we need to set it up here for testing
	jsonOutput.Changes = diff.EntityChanges{
		Creating:         []diff.EntityState{},
		Updating:         []diff.EntityState{},
		Deleting:         []diff.EntityState{},
		DroppedCreations: []diff.EntityState{},
		DroppedUpdates:   []diff.EntityState{},
		DroppedDeletions: []diff.EntityState{},
	}

	currentState, err := state.NewKongState()
	require.NoError(t, err)

	// mock target state
	targetState, err := state.NewKongState()
	require.NoError(t, err)
	service := state.Service{
		Service: kong.Service{
			ID:   kong.String("service-1"),
			Name: kong.String("Service 1"),
		},
	}
	err = targetState.Services.Add(service)
	require.NoError(t, err)

	// Calling performDiff with dry=true to avoid actual API calls
	totalOps, err := performDiff(
		context.Background(),
		currentState,
		targetState,
		true,  // dry mode
		1,     // parallelism
		0,     // delay
		nil,   // client (not used in dry mode)
		false, // isKonnect
		true,  // enabled Json output
		ApplyTypeFull,
	)

	require.NoError(t, err)
	assert.Equal(t, 1, totalOps)

	// Verify jsonOutput is populated correctly
	assert.Equal(t, int32(1), jsonOutput.Summary.Creating)
	assert.Equal(t, int32(0), jsonOutput.Summary.Updating)
	assert.Equal(t, int32(0), jsonOutput.Summary.Deleting)
	assert.Equal(t, int32(1), jsonOutput.Summary.Total)

	// Verify changes are populated
	assert.Len(t, jsonOutput.Changes.Creating, 1)
	assert.Empty(t, jsonOutput.Changes.Updating)
	assert.Empty(t, jsonOutput.Changes.Deleting)
	assert.Empty(t, jsonOutput.Changes.DroppedCreations)
	assert.Empty(t, jsonOutput.Changes.DroppedUpdates)
	assert.Empty(t, jsonOutput.Changes.DroppedDeletions)
}

func TestGetFormatFlagValue(t *testing.T) {
	newCmd := func() *cobra.Command {
		cmd := &cobra.Command{}
		cmd.Flags().String("format", "yaml", "")
		return cmd
	}

	tests := []struct {
		name      string
		setupCmd  func(*cobra.Command)
		envValue  string
		flagValue string
		want      string
	}{
		{
			name:      "returns default when neither flag nor env var is set",
			setupCmd:  func(_ *cobra.Command) {},
			flagValue: "yaml",
			want:      "yaml",
		},
		{
			name: "CLI flag takes priority over env var",
			setupCmd: func(cmd *cobra.Command) {
				_ = cmd.Flags().Set("format", "json")
			},
			envValue:  "yaml",
			flagValue: "json",
			want:      "json",
		},
		{
			name:      "env var used when flag not explicitly set",
			setupCmd:  func(_ *cobra.Command) {},
			envValue:  "json",
			flagValue: "yaml",
			want:      "json",
		},
		{
			name:      "env var is case-preserved",
			setupCmd:  func(_ *cobra.Command) {},
			envValue:  "JSON",
			flagValue: "yaml",
			want:      "JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				t.Setenv("DECK_FORMAT", tt.envValue)
			}
			cmd := newCmd()
			tt.setupCmd(cmd)
			got := getFormatFlagValue(cmd, tt.flagValue)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCheckParallelism_Invalid(t *testing.T) {
	parallelism := 0
	err := syncMain(
		context.Background(),
		[]string{"-"}, // never read — guard returns before GetContentFromFiles
		true,          // dry
		parallelism,   // parallelism
		0,             // delay
		"",            // workspace
		false,         // enableJSONOutput
		ApplyTypeFull,
	)
	require.EqualError(t, err, "--parallelism cannot be less than 1, got 0")
}

func TestCheckParallelism_Positive(t *testing.T) {
	parallelism := 1
	err := checkParallelism(parallelism)
	require.NoError(t, err)
}
