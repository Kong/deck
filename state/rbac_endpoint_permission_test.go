package state

import (
	"reflect"
	"testing"

	"github.com/kong/go-kong/kong"
	"github.com/stretchr/testify/assert"
)

func rbacEndpointPermissionsCollection() *RBACEndpointPermissionsCollection {
	return state().RBACEndpointPermissions
}

func TestRBACEndpointPermissionsCollection_Add(t *testing.T) {
	type args struct {
		rbacEndpointPermission RBACEndpointPermission
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "errors when role is nil",
			args: args{
				rbacEndpointPermission: RBACEndpointPermission{
					RBACEndpointPermission: kong.RBACEndpointPermission{
						Workspace: kong.String("*"),
						Actions:   kong.StringSlice("read"),
						Endpoint:  kong.String("/foo"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "inserts without a workspace, endpoint, and role",
			args: args{
				rbacEndpointPermission: RBACEndpointPermission{
					RBACEndpointPermission: kong.RBACEndpointPermission{
						Workspace: kong.String("*"),
						Endpoint:  kong.String("/foo"),
						Actions:   kong.StringSlice("read"),
						Role:      &kong.RBACRole{ID: kong.String("1234")},
					},
				},
			},
			wantErr: false,
		},
	}
	k := rbacEndpointPermissionsCollection()
	rbacEndpointPermission1 := RBACEndpointPermission{
		RBACEndpointPermission: kong.RBACEndpointPermission{
			Workspace: kong.String("*"),
			Endpoint:  kong.String("*"),
			Actions:   kong.StringSlice("read"),
			Role:      &kong.RBACRole{ID: kong.String("1234")},
		},
	}
	k.Add(rbacEndpointPermission1)
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := k.Add(tt.args.rbacEndpointPermission); (err != nil) != tt.wantErr {
				t.Errorf("RBACEndpointPermissionsCollection.Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRBACEndpointPermissionsCollection_Get(t *testing.T) {
	type args struct {
		nameOrID string
	}
	rbacEndpointPermission1 := RBACEndpointPermission{
		RBACEndpointPermission: kong.RBACEndpointPermission{
			Workspace: kong.String("*"),
			Endpoint:  kong.String("/foo"),
			Actions:   kong.StringSlice("read"),
			Role:      &kong.RBACRole{ID: kong.String("1234")},
		},
	}
	rbacEndpointPermission2 := RBACEndpointPermission{
		RBACEndpointPermission: kong.RBACEndpointPermission{
			Workspace: kong.String("*"),
			Endpoint:  kong.String("/bar"),
			Actions:   kong.StringSlice("read"),
			Role:      &kong.RBACRole{ID: kong.String("1234")},
		},
	}
	tests := []struct {
		name    string
		args    args
		want    *RBACEndpointPermission
		wantErr bool
	}{
		{
			name: "gets a rbacEndpointPermission by ID",
			args: args{
				nameOrID: rbacEndpointPermission1.FriendlyName(),
			},
			want:    &rbacEndpointPermission1,
			wantErr: false,
		},
		{
			name: "gets a rbacEndpointPermission by Name",
			args: args{
				nameOrID: rbacEndpointPermission2.FriendlyName(),
			},
			want:    &rbacEndpointPermission2,
			wantErr: false,
		},
		{
			name: "returns an ErrNotFound when no rbacEndpointPermission found",
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
	k := rbacEndpointPermissionsCollection()
	k.Add(rbacEndpointPermission1)
	k.Add(rbacEndpointPermission2)
	for _, tt := range tests {
		tc := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := k.Get(tc.args.nameOrID)
			if (err != nil) != tc.wantErr {
				t.Errorf("RBACEndpointPermissionsCollection.Get() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("RBACEndpointPermissionsCollection.Get() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRBACEndpointPermissionsCollection_Update(t *testing.T) {
	rbacEndpointPermission1 := RBACEndpointPermission{
		RBACEndpointPermission: kong.RBACEndpointPermission{
			Workspace: kong.String("*"),
			Endpoint:  kong.String("/foo"),
			Actions:   kong.StringSlice("read"),
			Role:      &kong.RBACRole{ID: kong.String("1234")},
		},
	}
	rbacEndpointPermission2 := RBACEndpointPermission{
		RBACEndpointPermission: kong.RBACEndpointPermission{
			Workspace: kong.String("*"),
			Endpoint:  kong.String("/bar"),
			Actions:   kong.StringSlice("read"),
			Role:      &kong.RBACRole{ID: kong.String("1234")},
		},
	}
	rbacEndpointPermission3 := RBACEndpointPermission{
		RBACEndpointPermission: kong.RBACEndpointPermission{
			Workspace: kong.String("*"),
			Endpoint:  kong.String("/foo"),
			Actions:   kong.StringSlice("read"),
			Role:      &kong.RBACRole{ID: kong.String("1234")},
			Comment:   kong.String("updated!"),
		},
	}
	type args struct {
		rbacEndpointPermission RBACEndpointPermission
	}
	tests := []struct {
		name                          string
		args                          args
		wantErr                       bool
		updatedRBACEndpointPermission *RBACEndpointPermission
	}{
		{
			name: "update errors if rbacEndpointPermission does not exist",
			args: args{
				rbacEndpointPermission: RBACEndpointPermission{
					RBACEndpointPermission: kong.RBACEndpointPermission{
						Workspace: kong.String("foo"),
						Endpoint:  kong.String("bad"),
						Actions:   kong.StringSlice("read"),
						Role:      &kong.RBACRole{ID: kong.String("1234")},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "update succeeds when ID is supplied",
			args: args{
				rbacEndpointPermission: rbacEndpointPermission3,
			},
			wantErr:                       false,
			updatedRBACEndpointPermission: &rbacEndpointPermission3,
		},
	}
	k := rbacEndpointPermissionsCollection()
	k.Add(rbacEndpointPermission1)
	k.Add(rbacEndpointPermission2)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			if err := k.Update(tt.args.rbacEndpointPermission); (err != nil) != tt.wantErr {
				t.Errorf("RBACEndpointPermissionsCollection.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				got, _ := k.Get(tt.updatedRBACEndpointPermission.FriendlyName())

				if !reflect.DeepEqual(got, tt.updatedRBACEndpointPermission) {
					t.Errorf("update rbacEndpointPermission, got = %#v, want %#v", got, tt.updatedRBACEndpointPermission)
				}
			}
		})
	}
}

func TestRBACEndpointPermissionDelete(t *testing.T) {
	assert := assert.New(t)
	collection := rbacEndpointPermissionsCollection()

	rbacEndpointPermission := RBACEndpointPermission{RBACEndpointPermission: kong.RBACEndpointPermission{
		Workspace: kong.String("*"),
		Endpoint:  kong.String("/foo"),
		Actions:   kong.StringSlice("read"),
		Role:      &kong.RBACRole{ID: kong.String("1234")},
	}}

	err := collection.Add(rbacEndpointPermission)
	assert.Nil(err)

	re, err := collection.Get(rbacEndpointPermission.FriendlyName())
	assert.Nil(err)
	assert.NotNil(re)

	err = collection.Delete(re.FriendlyName())
	assert.Nil(err)

	err = collection.Delete(re.FriendlyName())
	assert.NotNil(err)
}

func TestRBACEndpointPermissionGetAll(t *testing.T) {
	assert := assert.New(t)
	collection := rbacEndpointPermissionsCollection()

	rbacEndpointPermission := RBACEndpointPermission{RBACEndpointPermission: kong.RBACEndpointPermission{
		Workspace: kong.String("*"),
		Endpoint:  kong.String("/first"),
		Actions:   kong.StringSlice("read"),
		Role:      &kong.RBACRole{ID: kong.String("1234")},
	}}

	err := collection.Add(rbacEndpointPermission)
	assert.Nil(err)

	rbacEndpointPermission2 := RBACEndpointPermission{RBACEndpointPermission: kong.RBACEndpointPermission{
		Workspace: kong.String("*"),
		Endpoint:  kong.String("/second"),
		Actions:   kong.StringSlice("read"),
		Role:      &kong.RBACRole{ID: kong.String("1234")},
	}}

	err = collection.Add(rbacEndpointPermission2)
	assert.Nil(err)

	rbacEndpointPermissions, err := collection.GetAll()

	assert.Nil(err)
	assert.Equal(2, len(rbacEndpointPermissions))
}

func TestRBACEndpointPermissionGetAllByServiceID(t *testing.T) {
	assert := assert.New(t)
	collection := rbacEndpointPermissionsCollection()

	rbacEndpointPermissions := []*RBACEndpointPermission{
		{RBACEndpointPermission: kong.RBACEndpointPermission{
			Workspace: kong.String("*"),
			Endpoint:  kong.String("/first"),
			Actions:   kong.StringSlice("read"),
			Role:      &kong.RBACRole{ID: kong.String("1234")},
		}},
		{RBACEndpointPermission: kong.RBACEndpointPermission{
			Workspace: kong.String("*"),
			Endpoint:  kong.String("/second"),
			Actions:   kong.StringSlice("read"),
			Role:      &kong.RBACRole{ID: kong.String("1234")},
		}},
		{RBACEndpointPermission: kong.RBACEndpointPermission{
			Workspace: kong.String("*"),
			Endpoint:  kong.String("/third"),
			Actions:   kong.StringSlice("read"),
			Role:      &kong.RBACRole{ID: kong.String("1234")},
		}},
		{RBACEndpointPermission: kong.RBACEndpointPermission{
			Workspace: kong.String("*"),
			Endpoint:  kong.String("/fourth"),
			Actions:   kong.StringSlice("read"),
			Role:      &kong.RBACRole{ID: kong.String("4321")},
		}},
		{RBACEndpointPermission: kong.RBACEndpointPermission{
			Workspace: kong.String("*"),
			Endpoint:  kong.String("/fifth"),
			Actions:   kong.StringSlice("read"),
			Role:      &kong.RBACRole{ID: kong.String("4321")},
		}},
	}

	for _, rbacEndpointPermission := range rbacEndpointPermissions {
		err := collection.Add(*rbacEndpointPermission)
		assert.Nil(err)
	}

	rbacEndpointPermissions, err := collection.GetAllByRoleID("1234")
	assert.Nil(err)
	assert.Equal(3, len(rbacEndpointPermissions))

	rbacEndpointPermissions, err = collection.GetAllByRoleID("4321")
	assert.Nil(err)
	assert.Equal(2, len(rbacEndpointPermissions))
}
