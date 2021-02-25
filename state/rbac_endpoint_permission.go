package state

import (
	"fmt"

	memdb "github.com/hashicorp/go-memdb"
	"github.com/kong/deck/state/indexers"
	"github.com/kong/deck/utils"
	"github.com/pkg/errors"
)

const (
	rbacEndpointPermissionTableName = "rbac-endpointpermission"
	rbacEndpointPermissionsByRoleID = "rbacEndpointPermissionsByRoleID"
)

var errInvalidRole = errors.New("role.ID is required in rbacEndpointPermission")
var errEndpointRequired = errors.New("endpoint is required in rbacEndpointPermission")
var errWorkspaceRequired = errors.New("workspace is required in rbacEndpointPermission")
var rbacEndpointPermissionTableSchema = &memdb.TableSchema{
	Name: rbacEndpointPermissionTableName,
	Indexes: map[string]*memdb.IndexSchema{
		// ID in the case of an RBACEndpointPermission is a composite key of role ID, workspace, and endpoint
		"id": {
			Name:    "id",
			Unique:  true,
			Indexer: &memdb.StringFieldIndex{Field: "ID"},
		},
		all: allIndex,
		// foreign
		rbacEndpointPermissionsByRoleID: {
			Name: rbacEndpointPermissionsByRoleID,
			Indexer: &indexers.SubFieldIndexer{
				Fields: []indexers.Field{
					{
						Struct: "Role",
						Sub:    "ID",
					},
				},
			},
		},
	},
}

func validateRoleForRBACEndpointPermission(rbacEndpointPermission *RBACEndpointPermission) error {
	if rbacEndpointPermission.Role == nil ||
		utils.Empty(rbacEndpointPermission.Role.ID) {
		return errInvalidRole
	}
	return nil
}

// RBACEndpointPermissionsCollection stores and indexes Kong RBACEndpointPermissions.
type RBACEndpointPermissionsCollection collection

// Add adds a rbacEndpointPermission into RBACEndpointPermissionsCollection
// rbacEndpointPermission.Endpoint should not be nil else an error is thrown.
func (k *RBACEndpointPermissionsCollection) Add(rbacEndpointPermission RBACEndpointPermission) error {
	// TODO abstract this check in the go-memdb library itself
	if utils.Empty(rbacEndpointPermission.Endpoint) {
		return errEndpointRequired
	}

	if utils.Empty(rbacEndpointPermission.Workspace) {
		return errWorkspaceRequired
	}

	if err := validateRoleForRBACEndpointPermission(&rbacEndpointPermission); err != nil {
		return err
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	var searchBy []string
	searchBy = append(searchBy, rbacEndpointPermission.Identifier())

	_, err := getRBACEndpointPermission(txn, searchBy...)
	if err == nil {
		return fmt.Errorf("inserting rbacEndpointPermission %v: %w", rbacEndpointPermission.Console(), ErrAlreadyExists)
	} else if err != ErrNotFound {
		return err
	}
	rbacEndpointPermission.ID = rbacEndpointPermission.Identifier()
	err = txn.Insert(rbacEndpointPermissionTableName, &rbacEndpointPermission)
	if err != nil {
		return err
	}
	txn.Commit()
	return nil
}

func getRBACEndpointPermission(txn *memdb.Txn, IDs ...string) (*RBACEndpointPermission, error) {
	for _, id := range IDs {
		res, err := multiIndexLookupUsingTxn(txn, rbacEndpointPermissionTableName,
			[]string{"id"}, id)
		if err == ErrNotFound {
			continue
		}
		if err != nil {
			return nil, err
		}
		rbacEndpointPermission, ok := res.(*RBACEndpointPermission)
		if !ok {
			panic(unexpectedType)
		}
		return &RBACEndpointPermission{RBACEndpointPermission: *rbacEndpointPermission.DeepCopy()}, nil
	}
	return nil, ErrNotFound
}

// Get gets a rbacEndpointPermission by name or ID.
func (k *RBACEndpointPermissionsCollection) Get(nameOrID string) (*RBACEndpointPermission, error) {
	if nameOrID == "" {
		return nil, errIDRequired
	}

	txn := k.db.Txn(false)
	defer txn.Abort()
	rbacEndpointPermission, err := getRBACEndpointPermission(txn, nameOrID)
	if err != nil {
		return nil, err
	}
	return rbacEndpointPermission, nil
}

// Update updates a rbacEndpointPermission
func (k *RBACEndpointPermissionsCollection) Update(rbacEndpointPermission RBACEndpointPermission) error {
	if utils.Empty(rbacEndpointPermission.Endpoint) {
		return errEndpointRequired
	}
	if utils.Empty(rbacEndpointPermission.Workspace) {
		return errWorkspaceRequired
	}
	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteRBACEndpointPermission(txn, rbacEndpointPermission.Identifier())
	if err != nil {
		return err
	}

	rbacEndpointPermission.ID = rbacEndpointPermission.Identifier()
	err = txn.Insert(rbacEndpointPermissionTableName, &rbacEndpointPermission)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func deleteRBACEndpointPermission(txn *memdb.Txn, nameOrID string) error {
	rbacEndpointPermission, err := getRBACEndpointPermission(txn, nameOrID)
	if err != nil {
		return err
	}
	rbacEndpointPermission.ID = rbacEndpointPermission.Identifier()
	err = txn.Delete(rbacEndpointPermissionTableName, rbacEndpointPermission)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes a rbacEndpointPermission by name or ID.
func (k *RBACEndpointPermissionsCollection) Delete(endpointIdentifier string) error {
	if endpointIdentifier == "" {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteRBACEndpointPermission(txn, endpointIdentifier)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

// GetAll gets a rbacEndpointPermission by name or ID.
func (k *RBACEndpointPermissionsCollection) GetAll() ([]*RBACEndpointPermission, error) {
	txn := k.db.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(rbacEndpointPermissionTableName, all, true)
	if err != nil {
		return nil, err
	}

	var res []*RBACEndpointPermission
	for el := iter.Next(); el != nil; el = iter.Next() {
		r, ok := el.(*RBACEndpointPermission)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &RBACEndpointPermission{RBACEndpointPermission: *r.DeepCopy()})
	}
	txn.Commit()
	return res, nil
}

// GetAllByRoleID returns all endpoint permissions by referencing a role
// by its id.
func (k *RBACEndpointPermissionsCollection) GetAllByRoleID(id string) ([]*RBACEndpointPermission,
	error) {
	txn := k.db.Txn(false)
	iter, err := txn.Get(rbacEndpointPermissionTableName, rbacEndpointPermissionsByRoleID, id)
	if err != nil {
		return nil, err
	}
	var res []*RBACEndpointPermission
	for el := iter.Next(); el != nil; el = iter.Next() {
		r, ok := el.(*RBACEndpointPermission)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &RBACEndpointPermission{RBACEndpointPermission: *r.DeepCopy()})
	}
	return res, nil
}
