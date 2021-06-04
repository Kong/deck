package utils

import (
	"testing"

	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

func TestZeroOutID(t *testing.T) {
	type args struct {
		obj     interface{}
		altName *string
		withID  bool
	}
	tests := []struct {
		name        string
		args        args
		expectedObj interface{}
	}{
		{
			name: "zeros out ID when name is set",
			args: args{
				obj: &kong.Service{
					ID:   kong.String("foo-id"),
					Name: kong.String("foo-name"),
				},
				altName: kong.String("Name"),
				withID:  false,
			},
			expectedObj: &kong.Service{
				Name: kong.String("foo-name"),
			},
		},
		{
			name: "does not error out if ID is already zero value",
			args: args{
				obj: &kong.Service{
					Name: kong.String("foo-name"),
				},
				altName: kong.String("Name"),
				withID:  false,
			},
			expectedObj: &kong.Service{
				Name: kong.String("foo-name"),
			},
		},
		{
			name: "does not error out if provided value is not a pointer",
			args: args{
				obj: kong.Service{
					ID:   kong.String("foo-id"),
					Name: kong.String("foo-name"),
				},
				altName: kong.String("Name"),
				withID:  false,
			},
			expectedObj: kong.Service{
				ID:   kong.String("foo-id"),
				Name: kong.String("foo-name"),
			},
		},
		{
			name: "does not zero out ID when withID is set to true",
			args: args{
				obj: &kong.Service{
					ID:   kong.String("foo-id"),
					Name: kong.String("foo-name"),
				},
				altName: kong.String("Name"),
				withID:  true,
			},
			expectedObj: &kong.Service{
				ID:   kong.String("foo-id"),
				Name: kong.String("foo-name"),
			},
		},
		{
			name: "does not zero out ID when altName is not provided",
			args: args{
				obj: &kong.Service{
					ID:   kong.String("foo-id"),
					Name: kong.String("foo-name"),
				},
				withID: false,
			},
			expectedObj: &kong.Service{
				ID:   kong.String("foo-id"),
				Name: kong.String("foo-name"),
			},
		},
	}
	t.Parallel()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ZeroOutID(tt.args.obj, tt.args.altName, tt.args.withID)
			assert.Equal(t, tt.expectedObj, tt.args.obj)
		})
	}
}

func TestZeroOutTimestamps(t *testing.T) {
	type args struct {
		obj interface{}
	}
	tests := []struct {
		name        string
		args        args
		expectedObj interface{}
	}{
		{
			name: "clears timestamps when set",
			args: args{
				obj: &kong.Service{
					ID:        kong.String("foo-id"),
					Name:      kong.String("foo-name"),
					CreatedAt: kong.Int(42),
					UpdatedAt: kong.Int(42),
				},
			},
			expectedObj: &kong.Service{
				ID:   kong.String("foo-id"),
				Name: kong.String("foo-name"),
			},
		},
	}
	t.Parallel()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ZeroOutTimestamps(tt.args.obj)
			assert.Equal(t, tt.expectedObj, tt.args.obj)
		})
	}
}
