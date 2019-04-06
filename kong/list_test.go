package kong

import (
	"reflect"
	"testing"
)

func TestConstructQueryString(t *testing.T) {

}

func Test_constructQueryString(t *testing.T) {
	type args struct {
		opt *ListOpt
	}
	tests := []struct {
		name string
		args args
		want qs
	}{
		{
			"nil opt", args{}, qs{},
		},
		{
			"empty opt", args{opt: &ListOpt{}}, qs{},
		},
		{
			"size", args{opt: &ListOpt{Size: 42}}, qs{Size: 42},
		},
		{
			"offset", args{opt: &ListOpt{Offset: "42"}}, qs{Offset: "42"},
		},
		{
			"Single tag",
			args{opt: &ListOpt{Tags: StringSlice("tag1")}},
			qs{Tags: "tag1"},
		},
		{
			"Multiple AND tags",
			args{opt: &ListOpt{Tags: StringSlice("tag1", "tag2", "tag3")}},
			qs{Tags: "tag1/tag2/tag3"},
		},
		{
			"Multiple AND tags",
			args{opt: &ListOpt{Tags: StringSlice("tag1", "tag2", "tag3"),
				MatchAllTags: true}},
			qs{Tags: "tag1,tag2,tag3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := constructQueryString(tt.args.opt); !reflect.DeepEqual(got,
				tt.want) {
				t.Errorf("constructQueryString() = %v, want %v", got, tt.want)
			}
		})
	}
}
