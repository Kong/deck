package state

import (
	"reflect"
	"testing"

	"github.com/hbagdi/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

func routesCollection() *RoutesCollection {
	return state().Routes
}

func TestRoutesCollection_Add(t *testing.T) {
	type args struct {
		route Route
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "errors when ID is nil",
			args: args{
				route: Route{
					Route: kong.Route{
						Name:  kong.String("foo"),
						Hosts: kong.StringSlice("example.com"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "inserts without a name",
			args: args{
				route: Route{
					Route: kong.Route{
						ID:    kong.String("id1"),
						Hosts: kong.StringSlice("example.com"),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "inserts with a name and ID",
			args: args{
				route: Route{
					Route: kong.Route{
						ID:    kong.String("id2"),
						Name:  kong.String("bar-name"),
						Hosts: kong.StringSlice("example.com"),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "errors on re-insert when name is present",
			args: args{
				route: Route{
					Route: kong.Route{
						ID:    kong.String("id4"),
						Name:  kong.String("foo-name"),
						Hosts: kong.StringSlice("example.com"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "errors on re-insert when id is present",
			args: args{
				route: Route{
					Route: kong.Route{
						ID:    kong.String("id3"),
						Name:  kong.String("foobar-name"),
						Hosts: kong.StringSlice("example.com"),
					},
				},
			},
			wantErr: true,
		},
	}
	k := routesCollection()
	route1 := Route{
		Route: kong.Route{
			ID:    kong.String("id3"),
			Name:  kong.String("foo-name"),
			Hosts: kong.StringSlice("example.com"),
		},
	}
	k.Add(route1)
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := k.Add(tt.args.route); (err != nil) != tt.wantErr {
				t.Errorf("RoutesCollection.Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRoutesCollection_Get(t *testing.T) {
	type args struct {
		nameOrID string
	}
	route1 := Route{
		Route: kong.Route{
			ID:    kong.String("foo-id"),
			Hosts: kong.StringSlice("example.com"),
		},
	}
	route2 := Route{
		Route: kong.Route{
			ID:    kong.String("bar-id"),
			Name:  kong.String("bar-name"),
			Hosts: kong.StringSlice("example.com"),
		},
	}
	tests := []struct {
		name    string
		args    args
		want    *Route
		wantErr bool
	}{

		{
			name: "gets a route by ID",
			args: args{
				nameOrID: "foo-id",
			},
			want:    &route1,
			wantErr: false,
		},
		{
			name: "gets a route by Name",
			args: args{
				nameOrID: "bar-name",
			},
			want:    &route2,
			wantErr: false,
		},
		{
			name: "returns an ErrNotFound when no route found",
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
	k := routesCollection()
	k.Add(route1)
	k.Add(route2)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := k.Get(tt.args.nameOrID)
			if (err != nil) != tt.wantErr {
				t.Errorf("RoutesCollection.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RoutesCollection.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoutesInvalidType(t *testing.T) {
	assert := assert.New(t)

	collection := routesCollection()

	var service Service
	service.Name = kong.String("my-service")
	service.ID = kong.String("first")
	txn := collection.db.Txn(true)
	txn.Insert(routeTableName, &service)
	txn.Commit()

	assert.Panics(func() {
		collection.Get("my-service")
	})
	assert.Panics(func() {
		collection.GetAll()
	})
}

func TestRoutesCollection_Update(t *testing.T) {
	route1 := Route{
		Route: kong.Route{
			ID:    kong.String("foo-id"),
			Hosts: kong.StringSlice("example.com"),
		},
	}
	route2 := Route{
		Route: kong.Route{
			ID:    kong.String("bar-id"),
			Name:  kong.String("bar-name"),
			Hosts: kong.StringSlice("example.com"),
		},
	}
	route3 := Route{
		Route: kong.Route{
			ID:    kong.String("foo-id"),
			Name:  kong.String("name"),
			Hosts: kong.StringSlice("example.com"),
		},
	}
	type args struct {
		route Route
	}
	tests := []struct {
		name         string
		args         args
		wantErr      bool
		updatedRoute *Route
	}{
		{
			name: "update errors if route.ID is nil",
			args: args{
				route: Route{
					Route: kong.Route{
						Name: kong.String("name"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "update errors if route does not exist",
			args: args{
				route: Route{
					Route: kong.Route{
						ID: kong.String("does-not-exist"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "update succeeds when ID is supplied",
			args: args{
				route: route3,
			},
			wantErr:      false,
			updatedRoute: &route3,
		},
	}
	k := routesCollection()
	k.Add(route1)
	k.Add(route2)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//t.Parallel()
			if err := k.Update(tt.args.route); (err != nil) != tt.wantErr {
				t.Errorf("RoutesCollection.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				got, _ := k.Get(*tt.updatedRoute.ID)

				if !reflect.DeepEqual(got, tt.updatedRoute) {
					t.Errorf("update route, got = %#v, want %#v", got, tt.updatedRoute)
				}
			}
		})
	}
}

// Regression test
// to ensure that the memory reference of the pointer returned by Get()
// is different from the one stored in MemDB.
func TestRouteGetMemoryReference(t *testing.T) {
	assert := assert.New(t)
	collection := routesCollection()

	var route Route
	route.Name = kong.String("my-route")
	route.ID = kong.String("first")
	route.Hosts = kong.StringSlice("example.com", "demo.example.com")
	route.Service = &kong.Service{
		ID: kong.String("service1-id"),
	}
	assert.NotNil(route.Service)
	err := collection.Add(route)
	assert.NotNil(route.Service)
	assert.Nil(err)

	re, err := collection.Get("first")
	assert.Nil(err)
	assert.NotNil(re)
	assert.Equal("my-route", *re.Name)

	re.SNIs = kong.StringSlice("example.com", "demo.example.com")

	re, err = collection.Get("my-route")
	assert.Nil(err)
	assert.NotNil(re)
	assert.Nil(re.SNIs)
}

func TestRouteDelete(t *testing.T) {
	assert := assert.New(t)
	collection := routesCollection()

	var route Route
	route.Name = kong.String("my-route")
	route.ID = kong.String("first")
	route.Hosts = kong.StringSlice("example.com", "demo.example.com")
	route.Service = &kong.Service{
		ID: kong.String("service1-id"),
	}
	err := collection.Add(route)
	assert.Nil(err)

	re, err := collection.Get("my-route")
	assert.Nil(err)
	assert.NotNil(re)
	assert.Equal("example.com", *re.Hosts[0])

	err = collection.Delete(*re.ID)
	assert.Nil(err)

	err = collection.Delete(*re.ID)
	assert.NotNil(err)
}

func TestRouteGetAll(t *testing.T) {
	assert := assert.New(t)
	collection := routesCollection()

	var route Route
	route.Name = kong.String("my-route1")
	route.ID = kong.String("first")
	route.Hosts = kong.StringSlice("example.com", "demo.example.com")
	route.Service = &kong.Service{
		ID: kong.String("service1-id"),
	}
	err := collection.Add(route)
	assert.Nil(err)

	var route2 Route
	route2.Name = kong.String("my-route2")
	route2.ID = kong.String("second")
	route2.Hosts = kong.StringSlice("example.com", "demo.example.com")
	route2.Service = &kong.Service{
		ID: kong.String("service1-id"),
	}
	err = collection.Add(route2)
	assert.Nil(err)

	routes, err := collection.GetAll()

	assert.Nil(err)
	assert.Equal(2, len(routes))
}

func TestRouteGetAllByServiceID(t *testing.T) {
	assert := assert.New(t)
	collection := routesCollection()

	routes := []*Route{
		{
			Route: kong.Route{
				ID: kong.String("route0-id"),
			},
		},
		{
			Route: kong.Route{
				ID:   kong.String("route1-id"),
				Name: kong.String("route1-name"),
				Service: &kong.Service{
					ID: kong.String("service1-id"),
				},
			},
		},
		{
			Route: kong.Route{
				ID: kong.String("route2-id"),
				Service: &kong.Service{
					ID: kong.String("service1-id"),
				},
			},
		},
		{
			Route: kong.Route{
				ID:   kong.String("route3-id"),
				Name: kong.String("route3-name"),
				Service: &kong.Service{
					ID: kong.String("service2-id"),
				},
			},
		},
		{
			Route: kong.Route{
				ID:   kong.String("route4-id"),
				Name: kong.String("route4-name"),
				Service: &kong.Service{
					ID: kong.String("service2-id"),
				},
			},
		},
		{
			Route: kong.Route{
				ID: kong.String("route5-id"),
				Service: &kong.Service{
					ID: kong.String("service2-id"),
				},
			},
		},
	}

	for _, route := range routes {
		err := collection.Add(*route)
		assert.Nil(err)
	}

	routes, err := collection.GetAllByServiceID("service1-id")
	assert.Nil(err)
	assert.Equal(2, len(routes))

	routes, err = collection.GetAllByServiceID("service2-id")
	assert.Nil(err)
	assert.Equal(3, len(routes))
}
