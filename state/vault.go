package state

import (
	"errors"
	"fmt"

	memdb "github.com/hashicorp/go-memdb"
	"github.com/kong/deck/utils"
)

const (
	vaultTableName = "vault"
)

var vaultTableSchema = &memdb.TableSchema{
	Name: vaultTableName,
	Indexes: map[string]*memdb.IndexSchema{
		"id": {
			Name:    "id",
			Unique:  true,
			Indexer: &memdb.StringFieldIndex{Field: "ID"},
		},
		"prefix": {
			Name:    "prefix",
			Unique:  true,
			Indexer: &memdb.StringFieldIndex{Field: "Prefix"},
		},
		all: allIndex,
	},
}

// VaultsCollection stores and indexes Kong Vaults.
type VaultsCollection collection

// Add adds a vault to the collection.
// vault.ID should not be nil else an error is thrown.
func (k *VaultsCollection) Add(vault Vault) error {
	if utils.Empty(vault.ID) {
		return errIDRequired
	}
	txn := k.db.Txn(true)
	defer txn.Abort()

	var searchBy []string
	searchBy = append(searchBy, *vault.ID)
	if !utils.Empty(vault.Prefix) {
		searchBy = append(searchBy, *vault.Prefix)
	}
	_, err := getVault(txn, searchBy...)
	if err == nil {
		return fmt.Errorf("inserting vault %v: %w", vault.Console(), ErrAlreadyExists)
	} else if !errors.Is(err, ErrNotFound) {
		return err
	}

	err = txn.Insert(vaultTableName, &vault)
	if err != nil {
		return err
	}
	txn.Commit()
	return nil
}

func getVault(txn *memdb.Txn, IDs ...string) (*Vault, error) {
	for _, id := range IDs {
		res, err := multiIndexLookupUsingTxn(txn, vaultTableName,
			[]string{"prefix", "id"}, id)
		if errors.Is(err, ErrNotFound) {
			continue
		}
		if err != nil {
			return nil, err
		}

		vault, ok := res.(*Vault)
		if !ok {
			panic(unexpectedType)
		}
		return &Vault{Vault: *vault.DeepCopy()}, nil
	}
	return nil, ErrNotFound
}

// Get gets a vault by prefix or ID.
func (k *VaultsCollection) Get(prefixOrID string) (*Vault, error) {
	if prefixOrID == "" {
		return nil, errIDRequired
	}

	txn := k.db.Txn(false)
	defer txn.Abort()
	vault, err := getVault(txn, prefixOrID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return vault, nil
}

// Update udpates an existing vault.
func (k *VaultsCollection) Update(vault Vault) error {
	if utils.Empty(vault.ID) {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteVault(txn, *vault.ID)
	if err != nil {
		return err
	}

	err = txn.Insert(vaultTableName, &vault)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func deleteVault(txn *memdb.Txn, nameOrID string) error {
	vault, err := getVault(txn, nameOrID)
	if err != nil {
		return err
	}

	err = txn.Delete(vaultTableName, vault)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes a vault by its prefix or ID.
func (k *VaultsCollection) Delete(prefixOrID string) error {
	if prefixOrID == "" {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteVault(txn, prefixOrID)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

// GetAll gets all vaults in the state.
func (k *VaultsCollection) GetAll() ([]*Vault, error) {
	txn := k.db.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(vaultTableName, all, true)
	if err != nil {
		return nil, err
	}

	var res []*Vault
	for el := iter.Next(); el != nil; el = iter.Next() {
		v, ok := el.(*Vault)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &Vault{Vault: *v.DeepCopy()})
	}
	txn.Commit()
	return res, nil
}
