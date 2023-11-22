package cmd

import (
	"reflect"
	"testing"

	"github.com/kong/go-database-reconciler/pkg/dump"
	"github.com/kong/go-database-reconciler/pkg/file"
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
