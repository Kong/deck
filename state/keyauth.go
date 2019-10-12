package state

import (
	memdb "github.com/hashicorp/go-memdb"
	"github.com/hbagdi/deck/state/indexers"
	"github.com/pkg/errors"
)

const (
	keyAuthTableName           = "keyAuth"
	keyAuthsByConsumerUsername = "keyAuthsByConsumerUsername"
	keyAuthsByConsumerID       = "keyAuthsByConsumerID"
)

var keyAuthTableSchema = &memdb.TableSchema{
	Name: keyAuthTableName,
	Indexes: map[string]*memdb.IndexSchema{
		"id": {
			Name:    "id",
			Unique:  true,
			Indexer: &memdb.StringFieldIndex{Field: "ID"},
		},
		keyAuthsByConsumerUsername: {
			Name: keyAuthsByConsumerUsername,
			Indexer: &indexers.SubFieldIndexer{
				Fields: []indexers.Field{
					{
						Struct: "Consumer",
						Sub:    "Username",
					},
				},
			},
		},
		keyAuthsByConsumerID: {
			Name: keyAuthsByConsumerID,
			Indexer: &indexers.SubFieldIndexer{
				Fields: []indexers.Field{
					{
						Struct: "Consumer",
						Sub:    "ID",
					},
				},
			},
		},
		"key": {
			Name:    "key",
			Unique:  true,
			Indexer: &memdb.StringFieldIndex{Field: "Key"},
		},
		all: allIndex,
	},
}

// KeyAuthsCollection stores and indexes key-auth credentials.
type KeyAuthsCollection collection

// Add adds a key-auth credential to KeyAuthsCollection
func (k *KeyAuthsCollection) Add(keyAuth KeyAuth) error {
	txn := k.db.Txn(true)
	defer txn.Abort()
	err := txn.Insert(keyAuthTableName, &keyAuth)
	if err != nil {
		return errors.Wrap(err, "insert failed")
	}
	txn.Commit()
	return nil
}

// Get gets a key-auth credential by key or ID.
func (k *KeyAuthsCollection) Get(keyOrID string) (*KeyAuth, error) {
	res, err := multiIndexLookup(k.db, keyAuthTableName,
		[]string{"key", "id"}, keyOrID)
	if err == ErrNotFound {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, errors.Wrap(err, "keyAuth lookup failed")
	}
	if res == nil {
		return nil, ErrNotFound
	}
	keyAuth, ok := res.(*KeyAuth)
	if !ok {
		panic("unexpected type found")
	}
	return &KeyAuth{KeyAuth: *keyAuth.DeepCopy()}, nil
}

// GetAllByConsumerUsername returns all key-auth credentials
// belong to a Consumer with username.
func (k *KeyAuthsCollection) GetAllByConsumerUsername(username string) ([]*KeyAuth,
	error) {
	txn := k.db.Txn(false)
	iter, err := txn.Get(keyAuthTableName, keyAuthsByConsumerUsername, username)
	if err != nil {
		return nil, err
	}
	var res []*KeyAuth
	for el := iter.Next(); el != nil; el = iter.Next() {
		r, ok := el.(*KeyAuth)
		if !ok {
			panic("unexpected type found")
		}
		res = append(res, &KeyAuth{KeyAuth: *r.DeepCopy()})
	}
	return res, nil
}

// GetAllByConsumerID returns all key-auth credentials
// belong to a Consumer with id.
func (k *KeyAuthsCollection) GetAllByConsumerID(id string) ([]*KeyAuth,
	error) {
	txn := k.db.Txn(false)
	iter, err := txn.Get(keyAuthTableName, keyAuthsByConsumerID, id)
	if err != nil {
		return nil, err
	}
	var res []*KeyAuth
	for el := iter.Next(); el != nil; el = iter.Next() {
		r, ok := el.(*KeyAuth)
		if !ok {
			panic("unexpected type found")
		}
		res = append(res, &KeyAuth{KeyAuth: *r.DeepCopy()})
	}
	return res, nil
}

// Update updates an existing key-auth credential.
func (k *KeyAuthsCollection) Update(keyAuth KeyAuth) error {
	txn := k.db.Txn(true)
	defer txn.Abort()
	err := txn.Insert(keyAuthTableName, &keyAuth)
	if err != nil {
		return errors.Wrap(err, "update failed")
	}
	txn.Commit()
	return nil
}

// Delete deletes a key-auth credential by key or ID.
func (k *KeyAuthsCollection) Delete(keyOrID string) error {
	keyAuth, err := k.Get(keyOrID)

	if err != nil {
		return errors.Wrap(err, "looking up keyAuth")
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err = txn.Delete(keyAuthTableName, keyAuth)
	if err != nil {
		return errors.Wrap(err, "delete failed")
	}
	txn.Commit()
	return nil
}

// GetAll gets all key-auth credentials.
func (k *KeyAuthsCollection) GetAll() ([]*KeyAuth, error) {
	txn := k.db.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(keyAuthTableName, all, true)
	if err != nil {
		return nil, errors.Wrapf(err, "keyAuth lookup failed")
	}

	var res []*KeyAuth
	for el := iter.Next(); el != nil; el = iter.Next() {
		r, ok := el.(*KeyAuth)
		if !ok {
			panic("unexpected type found")
		}
		res = append(res, &KeyAuth{KeyAuth: *r.DeepCopy()})
	}
	txn.Commit()
	return res, nil
}
