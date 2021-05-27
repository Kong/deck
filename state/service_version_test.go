package state

import (
	"reflect"
	"testing"

	"github.com/kong/deck/konnect"
	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

func serviceVersionCollection() *ServiceVersionsCollection {
	return state().ServiceVersions
}

func TestServiceVersionCollection_Add(t *testing.T) {
	type args struct {
		serviceVersion ServiceVersion
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "errors when ID is nil",
			args: args{
				serviceVersion: ServiceVersion{
					ServiceVersion: konnect.ServiceVersion{
						Version: kong.String("foo"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "errors without a version",
			args: args{
				serviceVersion: ServiceVersion{
					ServiceVersion: konnect.ServiceVersion{
						ID: kong.String("id1"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "errors without a ServicePackage",
			args: args{
				serviceVersion: ServiceVersion{
					ServiceVersion: konnect.ServiceVersion{
						ID:      kong.String("id1"),
						Version: kong.String("bar-name"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "inserts with all valid details",
			args: args{
				serviceVersion: ServiceVersion{
					ServiceVersion: konnect.ServiceVersion{
						ID:      kong.String("id2"),
						Version: kong.String("bar-name"),
						ServicePackage: &konnect.ServicePackage{
							ID: kong.String("id1"),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "errors on re-insert when version is already present",
			args: args{
				serviceVersion: ServiceVersion{
					ServiceVersion: konnect.ServiceVersion{
						ID:      kong.String("id4"),
						Version: kong.String("foo-name"),
						ServicePackage: &konnect.ServicePackage{
							ID: kong.String("id1"),
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "errors on re-insert when id is present",
			args: args{
				serviceVersion: ServiceVersion{
					ServiceVersion: konnect.ServiceVersion{
						ID:      kong.String("id3"),
						Version: kong.String("foobar-name"),
						ServicePackage: &konnect.ServicePackage{
							ID: kong.String("id1"),
						},
					},
				},
			},
			wantErr: true,
		},
	}
	k := serviceVersionCollection()
	sv1 := ServiceVersion{
		ServiceVersion: konnect.ServiceVersion{
			ID:      kong.String("id3"),
			Version: kong.String("foo-name"),
			ServicePackage: &konnect.ServicePackage{
				ID: kong.String("id1"),
			},
		},
	}
	k.Add(sv1)
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := k.Add(tt.args.serviceVersion); (err != nil) != tt.wantErr {
				t.Errorf("ServiceVersionCollection.Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestServiceVersionCollection_Get(t *testing.T) {
	type args struct {
		nameOrID  string
		packageID string
	}
	sv1 := ServiceVersion{
		ServiceVersion: konnect.ServiceVersion{
			ID:      kong.String("foo-id"),
			Version: kong.String("foo-name"),
			ServicePackage: &konnect.ServicePackage{
				ID: kong.String("id1"),
			},
		},
	}
	sv2 := ServiceVersion{
		ServiceVersion: konnect.ServiceVersion{
			ID:      kong.String("bar-id"),
			Version: kong.String("bar-name"),
			ServicePackage: &konnect.ServicePackage{
				ID: kong.String("id1"),
			},
		},
	}
	tests := []struct {
		name    string
		args    args
		want    *ServiceVersion
		wantErr bool
	}{

		{
			name: "gets a serviceVersion by package and version ID",
			args: args{
				nameOrID:  "foo-id",
				packageID: "id1",
			},
			want:    &sv1,
			wantErr: false,
		},
		{
			name: "returns an error when only version is specified",
			args: args{
				nameOrID: "bar-name",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "returns an ErrNotFound when no serviceVersion found",
			args: args{
				nameOrID:  "baz-id",
				packageID: "id1",
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
	k := serviceVersionCollection()
	k.Add(sv1)
	k.Add(sv2)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := k.Get(tt.args.packageID, tt.args.nameOrID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ServiceVersionCollection.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ServiceVersionCollection.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServiceVersionCollection_Update(t *testing.T) {
	sv1 := ServiceVersion{
		ServiceVersion: konnect.ServiceVersion{
			ID:      kong.String("foo-id"),
			Version: kong.String("foo-name"),
			ServicePackage: &konnect.ServicePackage{
				ID: kong.String("id1"),
			},
		},
	}
	sv2 := ServiceVersion{
		ServiceVersion: konnect.ServiceVersion{
			ID:      kong.String("bar-id"),
			Version: kong.String("bar-name"),
			ServicePackage: &konnect.ServicePackage{
				ID: kong.String("id1"),
			},
		},
	}
	sv3 := ServiceVersion{
		ServiceVersion: konnect.ServiceVersion{
			ID:      kong.String("foo-id"),
			Version: kong.String("new-foo-name"),
			ServicePackage: &konnect.ServicePackage{
				ID: kong.String("id1"),
			},
		},
	}
	type args struct {
		serviceVersion ServiceVersion
	}
	tests := []struct {
		name           string
		args           args
		wantErr        bool
		updatedVersion *ServiceVersion
	}{
		{
			name: "update errors if serviceVersion.ID is nil",
			args: args{
				serviceVersion: ServiceVersion{
					ServiceVersion: konnect.ServiceVersion{
						Version: kong.String("name"),
						ServicePackage: &konnect.ServicePackage{
							ID: kong.String("id1"),
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "update errors if serviceVersion does not exist",
			args: args{
				serviceVersion: ServiceVersion{
					ServiceVersion: konnect.ServiceVersion{
						ID: kong.String("does-not-exist"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "update succeeds when ID is supplied",
			args: args{
				serviceVersion: sv3,
			},
			wantErr:        false,
			updatedVersion: &sv3,
		},
	}
	k := serviceVersionCollection()
	k.Add(sv1)
	k.Add(sv2)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			if err := k.Update(tt.args.serviceVersion); (err != nil) != tt.wantErr {
				t.Errorf("ServiceVersionCollection.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				got, _ := k.Get(*tt.updatedVersion.ServicePackage.ID, *tt.updatedVersion.ID)

				if !reflect.DeepEqual(got, tt.updatedVersion) {
					t.Errorf("update serviceVersion, got = %#v, want %#v", got, tt.updatedVersion)
				}
			}
		})
	}
}

func TestServiceVersionDelete(t *testing.T) {
	assert := assert.New(t)
	collection := serviceVersionCollection()

	var serviceVersion ServiceVersion
	serviceVersion.Version = kong.String("my-serviceVersion")
	serviceVersion.ID = kong.String("first")
	serviceVersion.ServicePackage = &konnect.ServicePackage{
		ID: kong.String("package-id1"),
	}
	err := collection.Add(serviceVersion)
	assert.Nil(err)

	re, err := collection.Get("package-id1", "my-serviceVersion")
	assert.Nil(err)
	assert.NotNil(re)

	err = collection.Delete("package-id1", *re.ID)
	assert.Nil(err)

	err = collection.Delete("package-id1", *re.ID)
	assert.NotNil(err)
}

func TestServiceVersionGetAll(t *testing.T) {
	assert := assert.New(t)
	collection := serviceVersionCollection()

	var serviceVersion ServiceVersion
	serviceVersion.Version = kong.String("my-sv1")
	serviceVersion.ID = kong.String("first")
	serviceVersion.ServicePackage = &konnect.ServicePackage{
		ID: kong.String("id1"),
	}
	err := collection.Add(serviceVersion)
	assert.Nil(err)

	var sv2 ServiceVersion
	sv2.Version = kong.String("my-sv2")
	sv2.ID = kong.String("second")
	sv2.ServicePackage = &konnect.ServicePackage{
		ID: kong.String("id1"),
	}
	err = collection.Add(sv2)
	assert.Nil(err)

	serviceVersions, err := collection.GetAll()

	assert.Nil(err)
	assert.Equal(2, len(serviceVersions))
}

func TestServiceVersionGetAllByServiceID(t *testing.T) {
	assert := assert.New(t)
	collection := serviceVersionCollection()

	serviceVersions := []*ServiceVersion{
		{
			ServiceVersion: konnect.ServiceVersion{
				ID:      kong.String("sv1-id"),
				Version: kong.String("sv1-name"),
				ServicePackage: &konnect.ServicePackage{
					ID: kong.String("id1"),
				},
			},
		},
		{
			ServiceVersion: konnect.ServiceVersion{
				ID:      kong.String("sv2-id"),
				Version: kong.String("sv2-name"),
				ServicePackage: &konnect.ServicePackage{
					ID: kong.String("id1"),
				},
			},
		},
		{
			ServiceVersion: konnect.ServiceVersion{
				ID:      kong.String("sv3-id"),
				Version: kong.String("sv3-name"),
				ServicePackage: &konnect.ServicePackage{
					ID: kong.String("id2"),
				},
			},
		},
		{
			ServiceVersion: konnect.ServiceVersion{
				ID:      kong.String("sv4-id"),
				Version: kong.String("sv4-name"),
				ServicePackage: &konnect.ServicePackage{
					ID: kong.String("id2"),
				},
			},
		},
		{
			ServiceVersion: konnect.ServiceVersion{
				ID:      kong.String("sv5-id"),
				Version: kong.String("sv5-name"),
				ServicePackage: &konnect.ServicePackage{
					ID: kong.String("id2"),
				},
			},
		},
	}

	for _, serviceVersion := range serviceVersions {
		err := collection.Add(*serviceVersion)
		assert.Nil(err)
	}

	serviceVersions, err := collection.GetAllByServicePackageID("id1")
	assert.Nil(err)
	assert.Equal(2, len(serviceVersions))

	serviceVersions, err = collection.GetAllByServicePackageID("id2")
	assert.Nil(err)
	assert.Equal(3, len(serviceVersions))
}
