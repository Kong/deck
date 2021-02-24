package state

import (
	"fmt"

	memdb "github.com/hashicorp/go-memdb"
	"github.com/kong/deck/utils"
)

const (
	rbacRoleTableName = "rbac-role"
)

var rbacRoleTableSchema = &memdb.TableSchema{
	Name: rbacRoleTableName,
	Indexes: map[string]*memdb.IndexSchema{
		"id": {
			Name:    "id",
			Unique:  true,
			Indexer: &memdb.StringFieldIndex{Field: "ID"},
		},
		"name": {
			Name:    "name",
			Unique:  true,
			Indexer: &memdb.StringFieldIndex{Field: "Name"},
		},
		all: allIndex,
	},
}

// RBACRolesCollection stores and indexes Kong RBACRoles.
type RBACRolesCollection collection

// Add adds a rbacRole into RBACRolesCollection
// rbacRole.ID should not be nil else an error is thrown.
func (k *RBACRolesCollection) Add(rbacRole RBACRole) error {
	// TODO abstract this check in the go-memdb library itself
	if utils.Empty(rbacRole.ID) {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	var searchBy []string
	searchBy = append(searchBy, *rbacRole.ID)
	if !utils.Empty(rbacRole.Name) {
		searchBy = append(searchBy, *rbacRole.Name)
	}
	_, err := getRBACRole(txn, searchBy...)
	if err == nil {
		return fmt.Errorf("inserting rbacRole %v: %w", rbacRole.Console(), ErrAlreadyExists)
	} else if err != ErrNotFound {
		return err
	}

	err = txn.Insert(rbacRoleTableName, &rbacRole)
	if err != nil {
		return err
	}
	txn.Commit()
	return nil
}

func getRBACRole(txn *memdb.Txn, IDs ...string) (*RBACRole, error) {
	for _, id := range IDs {
		res, err := multiIndexLookupUsingTxn(txn, rbacRoleTableName,
			[]string{"name", "id"}, id)
		if err == ErrNotFound {
			continue
		}
		if err != nil {
			return nil, err
		}

		rbacRole, ok := res.(*RBACRole)
		if !ok {
			panic(unexpectedType)
		}
		return &RBACRole{RBACRole: *rbacRole.DeepCopy()}, nil
	}
	return nil, ErrNotFound
}

// Get gets a rbacRole by name or ID.
func (k *RBACRolesCollection) Get(nameOrID string) (*RBACRole, error) {
	if nameOrID == "" {
		return nil, errIDRequired
	}

	txn := k.db.Txn(false)
	defer txn.Abort()
	rbacRole, err := getRBACRole(txn, nameOrID)
	if err != nil {
		return nil, err
	}
	return rbacRole, nil
}

// Update updates a rbacRole
func (k *RBACRolesCollection) Update(rbacRole RBACRole) error {
	// TODO abstract this check in the go-memdb library itself
	if utils.Empty(rbacRole.ID) {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteRBACRole(txn, *rbacRole.ID)
	if err != nil {
		return err
	}

	err = txn.Insert(rbacRoleTableName, &rbacRole)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func deleteRBACRole(txn *memdb.Txn, nameOrID string) error {
	rbacRole, err := getRBACRole(txn, nameOrID)
	if err != nil {
		return err
	}

	err = txn.Delete(rbacRoleTableName, rbacRole)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes a rbacRole by name or ID.
func (k *RBACRolesCollection) Delete(nameOrID string) error {
	if nameOrID == "" {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteRBACRole(txn, nameOrID)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

// GetAll gets a rbacRole by name or ID.
func (k *RBACRolesCollection) GetAll() ([]*RBACRole, error) {
	txn := k.db.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(rbacRoleTableName, all, true)
	if err != nil {
		return nil, err
	}

	var res []*RBACRole
	for el := iter.Next(); el != nil; el = iter.Next() {
		r, ok := el.(*RBACRole)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &RBACRole{RBACRole: *r.DeepCopy()})
	}
	txn.Commit()
	return res, nil
}
