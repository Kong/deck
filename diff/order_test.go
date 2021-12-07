package diff

import (
	"reflect"
	"testing"

	"github.com/kong/deck/types"
)

func Test_reverse(t *testing.T) {
	type args struct {
		src [][]types.EntityType
	}
	tests := []struct {
		name string
		args args
		want [][]types.EntityType
	}{
		{
			name: "doesn't panic on empty slice",
			args: args{
				src: nil,
			},
			want: [][]types.EntityType{},
		},
		{
			name: "doesn't panic on empty slice",
			args: args{
				src: [][]types.EntityType{
					{"foo"},
					{"bar"},
					{"baz", "fubar"},
				},
			},
			want: [][]types.EntityType{
				{"baz", "fubar"},
				{"bar"},
				{"foo"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := reverse(tt.args.src); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("reverse() = %v, want %v", got, tt.want)
			}
		})
	}
}
