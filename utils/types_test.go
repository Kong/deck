package utils

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestErrArrayString(t *testing.T) {
	assert := assert.New(t)
	var err ErrArray
	assert.Equal("nil", err.Error())

	err.Errors = append(err.Errors, errors.New("foo failed"))

	assert.Equal(err.Error(), "1 errors occurred:\n\tfoo failed\n")

	err.Errors = append(err.Errors, errors.New("bar failed"))

	assert.Equal(err.Error(), "2 errors occurred:\n\tfoo failed\n\tbar failed\n")
}

func Test_cleanAddress(t *testing.T) {
	type args struct {
		address string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			args: args{
				address: "foo",
			},
			want: "foo",
		},
		{
			args: args{
				address: "http://localhost:8001",
			},
			want: "http://localhost:8001",
		},
		{
			args: args{
				address: "http://localhost:8001/",
			},
			want: "http://localhost:8001",
		},
		{
			args: args{
				address: "http://localhost:8001//",
			},
			want: "http://localhost:8001",
		},
		{
			args: args{
				address: "https://subdomain.example.com///",
			},
			want: "https://subdomain.example.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanAddress(tt.args.address); got != tt.want {
				t.Errorf("cleanAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}
