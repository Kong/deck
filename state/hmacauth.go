package state

import (
	memdb "github.com/hashicorp/go-memdb"
	"github.com/hbagdi/deck/state/indexers"
	"github.com/pkg/errors"
)

const (
	hmacAuthTableName           = "hmacAuth"
	hmacAuthsByConsumerUsername = "hmacAuthsByConsumerUsername"
	hmacAuthsByConsumerID       = "hmacAuthsByConsumerID"
)

var hmacAuthTableSchema = &memdb.TableSchema{
	Name: hmacAuthTableName,
	Indexes: map[string]*memdb.IndexSchema{
		id: {
			Name:    id,
			Unique:  true,
			Indexer: &memdb.StringFieldIndex{Field: "ID"},
		},
		hmacAuthsByConsumerUsername: {
			Name: hmacAuthsByConsumerUsername,
			Indexer: &indexers.SubFieldIndexer{
				Fields: []indexers.Field{
					{
						Struct: "Consumer",
						Sub:    "Username",
					},
				},
			},
		},
		hmacAuthsByConsumerID: {
			Name: hmacAuthsByConsumerID,
			Indexer: &indexers.SubFieldIndexer{
				Fields: []indexers.Field{
					{
						Struct: "Consumer",
						Sub:    "ID",
					},
				},
			},
		},
		"username": {
			Name:    "username",
			Unique:  true,
			Indexer: &memdb.StringFieldIndex{Field: "Username"},
		},
		all: allIndex,
	},
}

// HMACAuthsCollection stores and indexes hmac-auth credentials.
type HMACAuthsCollection collection

// Add adds a hmac-auth credential to HMACAuthsCollection
func (k *HMACAuthsCollection) Add(hmacAuth HMACAuth) error {
	txn := k.db.Txn(true)
	defer txn.Abort()
	err := txn.Insert(hmacAuthTableName, &hmacAuth)
	if err != nil {
		return errors.Wrap(err, "insert failed")
	}
	txn.Commit()
	return nil
}

// Get gets a hmac-auth credential by  or ID.
func (k *HMACAuthsCollection) Get(usernameOrID string) (*HMACAuth, error) {
	res, err := multiIndexLookup(k.db, hmacAuthTableName,
		[]string{"username", id}, usernameOrID)
	if err == ErrNotFound {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, errors.Wrap(err, "hmacAuth lookup failed")
	}
	if res == nil {
		return nil, ErrNotFound
	}
	hmacAuth, ok := res.(*HMACAuth)
	if !ok {
		panic("unexpected type found")
	}
	return &HMACAuth{HMACAuth: *hmacAuth.DeepCopy()}, nil
}

// GetAllByConsumerUsername returns all hmac-auth credentials
// belong to a Consumer with username.
func (k *HMACAuthsCollection) GetAllByConsumerUsername(username string) ([]*HMACAuth,
	error) {
	txn := k.db.Txn(false)
	iter, err := txn.Get(hmacAuthTableName, hmacAuthsByConsumerUsername, username)
	if err != nil {
		return nil, err
	}
	var res []*HMACAuth
	for el := iter.Next(); el != nil; el = iter.Next() {
		r, ok := el.(*HMACAuth)
		if !ok {
			panic("unexpected type found")
		}
		res = append(res, &HMACAuth{HMACAuth: *r.DeepCopy()})
	}
	return res, nil
}

// GetAllByConsumerID returns all hmac-auth credentials
// belong to a Consumer with id.
func (k *HMACAuthsCollection) GetAllByConsumerID(id string) ([]*HMACAuth,
	error) {
	txn := k.db.Txn(false)
	iter, err := txn.Get(hmacAuthTableName, hmacAuthsByConsumerID, id)
	if err != nil {
		return nil, err
	}
	var res []*HMACAuth
	for el := iter.Next(); el != nil; el = iter.Next() {
		r, ok := el.(*HMACAuth)
		if !ok {
			panic("unexpected type found")
		}
		res = append(res, &HMACAuth{HMACAuth: *r.DeepCopy()})
	}
	return res, nil
}

// Update updates an existing hmac-auth credential.
func (k *HMACAuthsCollection) Update(hmacAuth HMACAuth) error {
	txn := k.db.Txn(true)
	defer txn.Abort()
	err := txn.Insert(hmacAuthTableName, &hmacAuth)
	if err != nil {
		return errors.Wrap(err, "update failed")
	}
	txn.Commit()
	return nil
}

// Delete deletes a hmac-auth credential by key or ID.
func (k *HMACAuthsCollection) Delete(usernameOrID string) error {
	hmacAuth, err := k.Get(usernameOrID)

	if err != nil {
		return errors.Wrap(err, "looking up hmacAuth")
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err = txn.Delete(hmacAuthTableName, hmacAuth)
	if err != nil {
		return errors.Wrap(err, "delete failed")
	}
	txn.Commit()
	return nil
}

// GetAll gets all hmac-auth credentials.
func (k *HMACAuthsCollection) GetAll() ([]*HMACAuth, error) {
	txn := k.db.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(hmacAuthTableName, all, true)
	if err != nil {
		return nil, errors.Wrapf(err, "hmacAuth lookup failed")
	}

	var res []*HMACAuth
	for el := iter.Next(); el != nil; el = iter.Next() {
		r, ok := el.(*HMACAuth)
		if !ok {
			panic("unexpected type found")
		}
		res = append(res, &HMACAuth{HMACAuth: *r.DeepCopy()})
	}
	txn.Commit()
	return res, nil
}
