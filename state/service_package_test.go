package state

import (
	"reflect"
	"testing"

	"github.com/kong/deck/konnect"
	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

func servicePackagesCollection() *ServicePackagesCollection {
	return state().ServicePackages
}

func TestServicePackagesCollection_Add(t *testing.T) {
	type args struct {
		servicePackage ServicePackage
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "errors when ID is nil",
			args: args{
				servicePackage: ServicePackage{
					ServicePackage: konnect.ServicePackage{
						Name: kong.String("foo"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "errors out without a name",
			args: args{
				servicePackage: ServicePackage{
					ServicePackage: konnect.ServicePackage{
						ID: kong.String("id1"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "inserts with a name and ID",
			args: args{
				servicePackage: ServicePackage{
					ServicePackage: konnect.ServicePackage{
						ID:   kong.String("id2"),
						Name: kong.String("foo-name"),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "errors on re-insert by ID",
			args: args{
				servicePackage: ServicePackage{
					ServicePackage: konnect.ServicePackage{
						ID:   kong.String("id3"),
						Name: kong.String("foo-name"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "errors on re-insert by Name",
			args: args{
				servicePackage: ServicePackage{
					ServicePackage: konnect.ServicePackage{
						ID:   kong.String("new-id"),
						Name: kong.String("bar-name"),
					},
				},
			},
			wantErr: true,
		},
	}
	k := servicePackagesCollection()
	svc1 := ServicePackage{
		ServicePackage: konnect.ServicePackage{
			ID:   kong.String("id3"),
			Name: kong.String("bar-name"),
		},
	}
	k.Add(svc1)
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := k.Add(tt.args.servicePackage); (err != nil) != tt.wantErr {
				t.Errorf("ServicePackageCollection.Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestServicePackagesCollection_Get(t *testing.T) {
	type args struct {
		nameOrID string
	}
	svc1 := ServicePackage{
		ServicePackage: konnect.ServicePackage{
			ID:   kong.String("foo-id"),
			Name: kong.String("foo-name"),
		},
	}
	svc2 := ServicePackage{
		ServicePackage: konnect.ServicePackage{
			ID:   kong.String("bar-id"),
			Name: kong.String("bar-name"),
		},
	}
	tests := []struct {
		name    string
		args    args
		want    *ServicePackage
		wantErr bool
	}{
		{
			name: "gets a servicePackage by ID",
			args: args{
				nameOrID: "foo-id",
			},
			want:    &svc1,
			wantErr: false,
		},
		{
			name: "gets a servicePackage by Name",
			args: args{
				nameOrID: "bar-name",
			},
			want:    &svc2,
			wantErr: false,
		},
		{
			name: "returns an ErrNotFound when no servicePackage found",
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
	k := servicePackagesCollection()
	k.Add(svc1)
	k.Add(svc2)
	for _, tt := range tests {
		tc := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := k.Get(tc.args.nameOrID)
			if (err != nil) != tc.wantErr {
				t.Errorf("ServicePackageCollection.Get() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("ServicePackageCollection.Get() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestServicePackagesCollection_Update(t *testing.T) {
	svc1 := ServicePackage{
		ServicePackage: konnect.ServicePackage{
			ID:   kong.String("foo-id"),
			Name: kong.String("foo-name"),
		},
	}
	svc2 := ServicePackage{
		ServicePackage: konnect.ServicePackage{
			ID:   kong.String("bar-id"),
			Name: kong.String("bar-name"),
		},
	}
	svc3 := ServicePackage{
		ServicePackage: konnect.ServicePackage{
			ID:   kong.String("foo-id"),
			Name: kong.String("name"),
		},
	}
	type args struct {
		servicePackage ServicePackage
	}
	tests := []struct {
		name           string
		args           args
		wantErr        bool
		updatedService *ServicePackage
	}{
		{
			name: "update errors if servicePackage.ID is nil",
			args: args{
				servicePackage: ServicePackage{
					ServicePackage: konnect.ServicePackage{
						Name: kong.String("name"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "update errors if servicePackage does not exist",
			args: args{
				servicePackage: ServicePackage{
					ServicePackage: konnect.ServicePackage{
						ID: kong.String("does-not-exist"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "update succeeds when ID is supplied",
			args: args{
				servicePackage: svc3,
			},
			wantErr:        false,
			updatedService: &svc3,
		},
	}
	k := servicePackagesCollection()
	k.Add(svc1)
	k.Add(svc2)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			if err := k.Update(tt.args.servicePackage); (err != nil) != tt.wantErr {
				t.Errorf("ServicePackageCollection.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				got, _ := k.Get(*tt.updatedService.ID)

				if !reflect.DeepEqual(got, tt.updatedService) {
					t.Errorf("update servicePackage, got = %#v, want %#v", got, tt.updatedService)
				}
			}
		})
	}
}

func TestServicePackageUpdate(t *testing.T) {
	assert := assert.New(t)
	k := servicePackagesCollection()
	svc1 := ServicePackage{
		ServicePackage: konnect.ServicePackage{
			ID:   kong.String("foo-id"),
			Name: kong.String("foo-name"),
		},
	}
	assert.Nil(k.Add(svc1))

	svc1.Name = kong.String("bar-name")
	assert.Nil(k.Update(svc1))

	r, err := k.Get("foo-id")
	assert.Nil(err)
	assert.NotNil(r)

	r, err = k.Get("bar-name")
	assert.Nil(err)
	assert.NotNil(r)

	r, err = k.Get("foo-name")
	assert.NotNil(err)
	assert.Nil(r)
}

func TestServicePackagesInvalidType(t *testing.T) {
	assert := assert.New(t)
	collection := servicePackagesCollection()

	var route Route
	route.Name = kong.String("my-route")
	route.ID = kong.String("first")
	txn := collection.db.Txn(true)
	txn.Insert(servicePackageTableName, &route)
	txn.Commit()

	assert.Panics(func() {
		collection.Get("my-route")
	})
	assert.Panics(func() {
		collection.GetAll()
	})
}

func TestServicePackageDelete(t *testing.T) {
	assert := assert.New(t)
	collection := servicePackagesCollection()

	var servicePackage ServicePackage
	servicePackage.ID = kong.String("first-id")
	servicePackage.Name = kong.String("first-name")
	err := collection.Add(servicePackage)
	assert.Nil(err)

	err = collection.Delete("does-not-exist")
	assert.NotNil(err)
	err = collection.Delete("first-id")
	assert.Nil(err)

	err = collection.Delete("first-name")
	assert.NotNil(err)

	err = collection.Delete("")
	assert.NotNil(err)
}

func TestServicePackageGetAll(t *testing.T) {
	assert := assert.New(t)
	collection := servicePackagesCollection()

	services := []ServicePackage{
		{
			ServicePackage: konnect.ServicePackage{
				ID:   kong.String("first"),
				Name: kong.String("my-service1"),
			},
		},
		{
			ServicePackage: konnect.ServicePackage{
				ID:   kong.String("second"),
				Name: kong.String("my-service2"),
			},
		},
	}
	for _, s := range services {
		assert.Nil(collection.Add(s))
	}

	allServices, err := collection.GetAll()

	assert.Nil(err)
	assert.Equal(len(services), len(allServices))
}

// Regression test
// to ensure that the memory reference of the pointer returned by Get()
// is different from the one stored in MemDB.
func TestServicePackagesGetAllMemoryReference(t *testing.T) {
	assert := assert.New(t)
	collection := servicePackagesCollection()

	services := []ServicePackage{
		{
			ServicePackage: konnect.ServicePackage{
				ID:          kong.String("first"),
				Name:        kong.String("my-service1"),
				Description: kong.String("service1-desc"),
			},
		},
		{
			ServicePackage: konnect.ServicePackage{
				ID:          kong.String("second"),
				Name:        kong.String("my-service2"),
				Description: kong.String("service2-desc"),
			},
		},
	}
	for _, s := range services {
		assert.Nil(collection.Add(s))
	}

	allServices, err := collection.GetAll()
	assert.Nil(err)
	assert.Equal(len(services), len(allServices))

	allServices[0].Description = kong.String("new-service1-desc")
	allServices[1].Description = kong.String("new-service2-desc")

	servicePackage, err := collection.Get("my-service1")
	assert.Nil(err)
	assert.NotNil(servicePackage)
	assert.Equal("service1-desc", *servicePackage.Description)
}
