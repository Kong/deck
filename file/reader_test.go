package file

import (
	"reflect"
	"testing"
)

func Test_ensureJSON(t *testing.T) {
	type args struct {
		m map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		{
			"empty array is kept as is",
			args{map[string]interface{}{
				"foo": []interface{}{},
			}},
			map[string]interface{}{
				"foo": []interface{}{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ensureJSON(tt.args.m); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ensureJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}
