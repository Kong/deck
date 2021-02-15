package state

import (
	"reflect"
	"testing"

	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

func rbacRolesCollection() *RBACRolesCollection {
	return state().RBACRoles
}

func TestRBACRolesCollection_Add(t *testing.T) {
	type args struct {
		rbacRole RBACRole
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "errors when ID is nil",
			args: args{
				rbacRole: RBACRole{
					RBACRole: kong.RBACRole{
						Name: kong.String("foo"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "inserts without a name",
			args: args{
				rbacRole: RBACRole{
					RBACRole: kong.RBACRole{
						ID: kong.String("id1"),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "inserts with a name and ID",
			args: args{
				rbacRole: RBACRole{
					RBACRole: kong.RBACRole{
						ID:   kong.String("id2"),
						Name: kong.String("bar-name"),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "errors on re-insert when name is present",
			args: args{
				rbacRole: RBACRole{
					RBACRole: kong.RBACRole{
						ID:   kong.String("id4"),
						Name: kong.String("foo-name"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "errors on re-insert when id is present",
			args: args{
				rbacRole: RBACRole{
					RBACRole: kong.RBACRole{
						ID:   kong.String("id3"),
						Name: kong.String("foobar-name"),
					},
				},
			},
			wantErr: true,
		},
	}
	k := rbacRolesCollection()
	rbacRole1 := RBACRole{
		RBACRole: kong.RBACRole{
			ID:   kong.String("id3"),
			Name: kong.String("foo-name"),
		},
	}
	k.Add(rbacRole1)
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := k.Add(tt.args.rbacRole); (err != nil) != tt.wantErr {
				t.Errorf("RBACRolesCollection.Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRBACRolesCollection_Get(t *testing.T) {
	type args struct {
		nameOrID string
	}
	rbacRole1 := RBACRole{
		RBACRole: kong.RBACRole{
			ID: kong.String("foo-id"),
		},
	}
	rbacRole2 := RBACRole{
		RBACRole: kong.RBACRole{
			ID:   kong.String("bar-id"),
			Name: kong.String("bar-name"),
		},
	}
	tests := []struct {
		name    string
		args    args
		want    *RBACRole
		wantErr bool
	}{

		{
			name: "gets a rbacRole by ID",
			args: args{
				nameOrID: "foo-id",
			},
			want:    &rbacRole1,
			wantErr: false,
		},
		{
			name: "gets a rbacRole by Name",
			args: args{
				nameOrID: "bar-name",
			},
			want:    &rbacRole2,
			wantErr: false,
		},
		{
			name: "returns an ErrNotFound when no rbacRole found",
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
	k := rbacRolesCollection()
	k.Add(rbacRole1)
	k.Add(rbacRole2)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := k.Get(tt.args.nameOrID)
			if (err != nil) != tt.wantErr {
				t.Errorf("RBACRolesCollection.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RBACRolesCollection.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRBACRolesCollection_Update(t *testing.T) {
	rbacRole1 := RBACRole{
		RBACRole: kong.RBACRole{
			ID: kong.String("foo-id"),
		},
	}
	rbacRole2 := RBACRole{
		RBACRole: kong.RBACRole{
			ID:   kong.String("bar-id"),
			Name: kong.String("bar-name"),
		},
	}
	rbacRole3 := RBACRole{
		RBACRole: kong.RBACRole{
			ID:   kong.String("foo-id"),
			Name: kong.String("name"),
		},
	}
	type args struct {
		rbacRole RBACRole
	}
	tests := []struct {
		name            string
		args            args
		wantErr         bool
		updatedRBACRole *RBACRole
	}{
		{
			name: "update errors if rbacRole.ID is nil",
			args: args{
				rbacRole: RBACRole{
					RBACRole: kong.RBACRole{
						Name: kong.String("name"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "update errors if rbacRole does not exist",
			args: args{
				rbacRole: RBACRole{
					RBACRole: kong.RBACRole{
						ID: kong.String("does-not-exist"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "update succeeds when ID is supplied",
			args: args{
				rbacRole: rbacRole3,
			},
			wantErr:         false,
			updatedRBACRole: &rbacRole3,
		},
	}
	k := rbacRolesCollection()
	k.Add(rbacRole1)
	k.Add(rbacRole2)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//t.Parallel()
			if err := k.Update(tt.args.rbacRole); (err != nil) != tt.wantErr {
				t.Errorf("RBACRolesCollection.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				got, _ := k.Get(*tt.updatedRBACRole.ID)

				if !reflect.DeepEqual(got, tt.updatedRBACRole) {
					t.Errorf("update rbacRole, got = %#v, want %#v", got, tt.updatedRBACRole)
				}
			}
		})
	}
}

// Regression test
// to ensure that the memory reference of the pointer returned by Get()
// is different from the one stored in MemDB.
// func TestRBACRoleGetMemoryReference(t *testing.T) {
// 	assert := assert.New(t)
// 	collection := rbacRolesCollection()

// 	var rbacRole RBACRole
// 	rbacRole.Name = kong.String("my-rbacRole")
// 	rbacRole.ID = kong.String("first")
// 	rbacRole.Hosts = kong.StringSlice("example.com", "demo.example.com")
// 	rbacRole.Service = &kong.Service{
// 		ID: kong.String("service1-id"),
// 	}
// 	assert.NotNil(rbacRole.Service)
// 	err := collection.Add(rbacRole)
// 	assert.NotNil(rbacRole.Service)
// 	assert.Nil(err)

// 	re, err := collection.Get("first")
// 	assert.Nil(err)
// 	assert.NotNil(re)
// 	assert.Equal("my-rbacRole", *re.Name)

// 	re.SNIs = kong.StringSlice("example.com", "demo.example.com")

// 	re, err = collection.Get("my-rbacRole")
// 	assert.Nil(err)
// 	assert.NotNil(re)
// 	assert.Nil(re.SNIs)
// }

func TestRBACRoleDelete(t *testing.T) {
	assert := assert.New(t)
	collection := rbacRolesCollection()

	var rbacRole RBACRole
	rbacRole.Name = kong.String("my-rbacRole")
	rbacRole.ID = kong.String("first")

	err := collection.Add(rbacRole)
	assert.Nil(err)

	re, err := collection.Get("my-rbacRole")
	assert.Nil(err)
	assert.NotNil(re)

	err = collection.Delete(*re.ID)
	assert.Nil(err)

	err = collection.Delete(*re.ID)
	assert.NotNil(err)
}

func TestRBACRoleGetAll(t *testing.T) {
	assert := assert.New(t)
	collection := rbacRolesCollection()

	var rbacRole RBACRole
	rbacRole.Name = kong.String("my-rbacRole1")
	rbacRole.ID = kong.String("first")

	err := collection.Add(rbacRole)
	assert.Nil(err)

	var rbacRole2 RBACRole
	rbacRole2.Name = kong.String("my-rbacRole2")
	rbacRole2.ID = kong.String("second")

	err = collection.Add(rbacRole2)
	assert.Nil(err)

	rbacRoles, err := collection.GetAll()

	assert.Nil(err)
	assert.Equal(2, len(rbacRoles))
}

// func TestRBACRoleGetAllByServiceID(t *testing.T) {
// 	assert := assert.New(t)
// 	collection := rbacRolesCollection()

// 	rbacRoles := []*RBACRole{
// 		{
// 			RBACRole: kong.RBACRole{
// 				ID: kong.String("rbacRole0-id"),
// 			},
// 		},
// 		{
// 			RBACRole: kong.RBACRole{
// 				ID:   kong.String("rbacRole1-id"),
// 				Name: kong.String("rbacRole1-name"),
// 			},
// 		},
// 		{
// 			RBACRole: kong.RBACRole{
// 				ID: kong.String("rbacRole2-id"),
// 			},
// 		},
// 		{
// 			RBACRole: kong.RBACRole{
// 				ID:   kong.String("rbacRole3-id"),
// 				Name: kong.String("rbacRole3-name"),
// 			},
// 		},
// 		{
// 			RBACRole: kong.RBACRole{
// 				ID:   kong.String("rbacRole4-id"),
// 				Name: kong.String("rbacRole4-name"),
// 			},
// 		},
// 		{
// 			RBACRole: kong.RBACRole{
// 				ID: kong.String("rbacRole5-id"),
// 			},
// 		},
// 	}

// 	for _, rbacRole := range rbacRoles {
// 		err := collection.Add(*rbacRole)
// 		assert.Nil(err)
// 	}

// 	rbacRoles, err := collection.GetAllByServiceID("service1-id")
// 	assert.Nil(err)
// 	assert.Equal(2, len(rbacRoles))

// 	rbacRoles, err = collection.GetAllByServiceID("service2-id")
// 	assert.Nil(err)
// 	assert.Equal(3, len(rbacRoles))
// }
