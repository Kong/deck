package utils

import (
	"net/http"
	"reflect"
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
			if got := CleanAddress(tt.args.address); got != tt.want {
				t.Errorf("cleanAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseHeaders(t *testing.T) {
	type args struct {
		headers []string
	}
	tests := []struct {
		name    string
		args    args
		want    http.Header
		wantErr bool
	}{
		{
			name: "nil headers returns without an error",
			args: args{
				headers: nil,
			},
			want:    http.Header{},
			wantErr: false,
		},
		{
			name: "empty headers returns without an error",
			args: args{
				headers: []string{},
			},
			want:    http.Header{},
			wantErr: false,
		},
		{
			name: "headers returns without an error",
			args: args{
				headers: []string{
					"foo:bar",
					"baz:fubar",
				},
			},
			want: http.Header{
				"Foo": []string{"bar"},
				"Baz": []string{"fubar"},
			},
			wantErr: false,
		},
		{
			name: "invalid headers value returns an error",
			args: args{
				headers: []string{
					"fubar",
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseHeaders(tt.args.headers)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseHeaders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseHeaders() = %v, want %v", got, tt.want)
			}
		})
	}
}
