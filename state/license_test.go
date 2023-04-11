package state

import (
	"reflect"
	"testing"

	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

func licensesCollection() *LicensesCollection {
	return state().Licenses
}

func TestLicensesCollection_Add(t *testing.T) {
	type args struct {
		license License
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "errors when ID is nil",
			args: args{
				license: License{
					License: kong.License{
						Payload: kong.String("example"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "inserts with a payload and ID",
			args: args{
				license: License{
					License: kong.License{
						ID:      kong.String("id2"),
						Payload: kong.String("example"),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "errors on re-insert by ID",
			args: args{
				license: License{
					License: kong.License{
						ID:      kong.String("dup"),
						Payload: kong.String("example"),
					},
				},
			},
			wantErr: true,
		},
	}
	k := licensesCollection()
	lic1 := License{
		License: kong.License{
			ID:      kong.String("dup"),
			Payload: kong.String("example"),
		},
	}
	k.Add(lic1)
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := k.Add(tt.args.license); (err != nil) != tt.wantErr {
				t.Errorf("LicensesCollection.Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLicensesCollection_Get(t *testing.T) {
	type args struct {
		nameOrID string
	}
	lic1 := License{
		License: kong.License{
			ID:      kong.String("foo-id"),
			Payload: kong.String("example"),
		},
	}
	lic2 := License{
		License: kong.License{
			ID:      kong.String("bar-id"),
			Payload: kong.String("example"),
		},
	}
	tests := []struct {
		name    string
		args    args
		want    *License
		wantErr bool
	}{
		{
			name: "gets a license by ID",
			args: args{
				nameOrID: "foo-id",
			},
			want:    &lic1,
			wantErr: false,
		},
		{
			name: "returns an ErrNotFound when no license found",
			args: args{
				nameOrID: "baz-id",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "returns an error when ID is empty",
			args: args{
				nameOrID: "",
			},
			want:    nil,
			wantErr: true,
		},
	}
	k := licensesCollection()
	k.Add(lic1)
	k.Add(lic2)
	for _, tt := range tests {
		tc := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := k.Get(tc.args.nameOrID)
			if (err != nil) != tc.wantErr {
				t.Errorf("LicensesCollection.Get() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestLicensesCollection_Update(t *testing.T) {
	lic1 := License{
		License: kong.License{
			ID:      kong.String("foo-id"),
			Payload: kong.String("example"),
		},
	}
	lic2 := License{
		License: kong.License{
			ID:      kong.String("bar-id"),
			Payload: kong.String("example"),
		},
	}
	lic3 := License{
		License: kong.License{
			ID:      kong.String("foo-id"),
			Payload: kong.String("example"),
		},
	}
	type args struct {
		license License
	}
	tests := []struct {
		name           string
		args           args
		wantErr        bool
		updatedLicense *License
	}{
		{
			name: "update errors if license.ID is nil",
			args: args{
				license: License{
					License: kong.License{},
				},
			},
			wantErr: true,
		},
		{
			name: "update errors if license does not exist",
			args: args{
				license: License{
					License: kong.License{
						ID:      kong.String("does-not-exist"),
						Payload: kong.String("example"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "update succeeds when ID is supplied",
			args: args{
				license: lic3,
			},
			wantErr:        false,
			updatedLicense: &lic3,
		},
	}
	k := licensesCollection()
	k.Add(lic1)
	k.Add(lic2)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			if err := k.Update(tt.args.license); (err != nil) != tt.wantErr {
				t.Errorf("LicensesCollection.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				got, _ := k.Get(*tt.updatedLicense.ID)

				if !reflect.DeepEqual(got, tt.updatedLicense) {
					t.Errorf("update license, got = %#v, want %#v", got, tt.updatedLicense)
				}
			}
		})
	}
}

func TestLicenseUpdate(t *testing.T) {
	assert := assert.New(t)
	k := licensesCollection()
	lic1 := License{
		License: kong.License{
			ID:      kong.String("foo-id"),
			Payload: kong.String("example"),
		},
	}
	assert.Nil(k.Add(lic1))

	lic1.Payload = kong.String("other")
	assert.Nil(k.Update(lic1))

	r, err := k.Get("foo-id")
	assert.Nil(err)
	assert.NotNil(r)
	assert.Equal(*r.Payload, "other")
}

func TestLicenseDelete(t *testing.T) {
	assert := assert.New(t)
	collection := licensesCollection()

	var license License
	license.ID = kong.String("first")
	license.Payload = kong.String("example")
	err := collection.Add(license)
	assert.NoError(err)

	err = collection.Delete("does-not-exist")
	assert.Error(err)
	err = collection.Delete("first")
	assert.NoError(err)

	err = collection.Delete("first")
	assert.Error(err)

	err = collection.Delete("")
	assert.Error(err)
}

func TestLicenseGetAll(t *testing.T) {
	assert := assert.New(t)
	collection := licensesCollection()

	licenses := []License{
		{
			License: kong.License{
				ID:      kong.String("first"),
				Payload: kong.String("example"),
			},
		},
		{
			License: kong.License{
				ID:      kong.String("second"),
				Payload: kong.String("example"),
			},
		},
	}
	for _, s := range licenses {
		assert.Nil(collection.Add(s))
	}

	allLicenses, err := collection.GetAll()

	assert.NoError(err)
	assert.Len(allLicenses, len(licenses))
}
