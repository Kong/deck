package convert

import (
	"testing"
)

func Test_isEmpty(t *testing.T) {
	tests := []struct {
		name     string
		v        interface{}
		expected bool
	}{
		{name: "nil", v: nil, expected: true},
		{name: "empty slice", v: []string{}, expected: true},
		{name: "non-empty slice", v: []string{"a"}, expected: false},
		{name: "empty array", v: [0]int{}, expected: true},
		{name: "non-empty array", v: [1]int{1}, expected: false},
		{name: "empty map", v: map[string]string{}, expected: true},
		{name: "non-empty map", v: map[string]string{"k": "v"}, expected: false},
		{name: "string", v: "hello", expected: false},
		{name: "int", v: 0, expected: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isEmpty(tt.v); got != tt.expected {
				t.Errorf("isEmpty(%v) = %v, expected %v", tt.v, got, tt.expected)
			}
		})
	}
}
