package state

import (
	"errors"
	"fmt"

	memdb "github.com/hashicorp/go-memdb"
	"github.com/kong/deck/utils"
)

const (
	keyTableName = "key"
)

var keyTableSchema = &memdb.TableSchema{
	Name: keyTableName,
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

// KeysCollection stores and indexes Kong Keys.
type KeysCollection collection

// Add adds a key to the collection.
// key.ID should not be nil else an error is thrown.
func (k *KeysCollection) Add(key Key) error {
	if utils.Empty(key.ID) {
		return errIDRequired
	}
	txn := k.db.Txn(true)
	defer txn.Abort()

	var searchBy []string
	searchBy = append(searchBy, *key.ID)
	if !utils.Empty(key.Name) {
		searchBy = append(searchBy, *key.Name)
	}
	_, err := getKey(txn, searchBy...)
	if err == nil {
		return fmt.Errorf("inserting key %v: %w", key.Console(), ErrAlreadyExists)
	} else if !errors.Is(err, ErrNotFound) {
		return err
	}

	err = txn.Insert(keyTableName, &key)
	if err != nil {
		return err
	}
	txn.Commit()
	return nil
}

func getKey(txn *memdb.Txn, IDs ...string) (*Key, error) {
	for _, id := range IDs {
		res, err := multiIndexLookupUsingTxn(txn, keyTableName,
			[]string{"name", "id"}, id)
		if errors.Is(err, ErrNotFound) {
			continue
		}
		if err != nil {
			return nil, err
		}

		key, ok := res.(*Key)
		if !ok {
			panic(unexpectedType)
		}
		return &Key{Key: *key.DeepCopy()}, nil
	}
	return nil, ErrNotFound
}

// Get gets a key by name or ID.
func (k *KeysCollection) Get(nameOrID string) (*Key, error) {
	if nameOrID == "" {
		return nil, errIDRequired
	}

	txn := k.db.Txn(false)
	defer txn.Abort()
	key, err := getKey(txn, nameOrID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return key, nil
}

// Update udpates an existing key.
func (k *KeysCollection) Update(key Key) error {
	if utils.Empty(key.ID) {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteKey(txn, *key.ID)
	if err != nil {
		return err
	}

	err = txn.Insert(keyTableName, &key)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func deleteKey(txn *memdb.Txn, nameOrID string) error {
	key, err := getKey(txn, nameOrID)
	if err != nil {
		return err
	}

	err = txn.Delete(keyTableName, key)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes a key by its name or ID.
func (k *KeysCollection) Delete(nameOrID string) error {
	if nameOrID == "" {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteKey(txn, nameOrID)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

// GetAll gets all key in the state.
func (k *KeysCollection) GetAll() ([]*Key, error) {
	txn := k.db.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(keyTableName, all, true)
	if err != nil {
		return nil, err
	}

	var res []*Key
	for el := iter.Next(); el != nil; el = iter.Next() {
		k, ok := el.(*Key)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &Key{Key: *k.DeepCopy()})
	}
	txn.Commit()
	return res, nil
}
