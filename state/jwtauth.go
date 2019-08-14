package state

import (
	memdb "github.com/hashicorp/go-memdb"
	"github.com/hbagdi/deck/state/indexers"
	"github.com/pkg/errors"
)

const (
	jwtAuthTableName           = "jwtAuth"
	jwtAuthsByConsumerUsername = "jwtAuthsByConsumerUsername"
	jwtAuthsByConsumerID       = "jwtAuthsByConsumerID"
)

var jwtAuthTableSchema = &memdb.TableSchema{
	Name: jwtAuthTableName,
	Indexes: map[string]*memdb.IndexSchema{
		id: {
			Name:    id,
			Unique:  true,
			Indexer: &memdb.StringFieldIndex{Field: "ID"},
		},
		jwtAuthsByConsumerUsername: {
			Name: jwtAuthsByConsumerUsername,
			Indexer: &indexers.SubFieldIndexer{
				Fields: []indexers.Field{
					{
						Struct: "Consumer",
						Sub:    "Username",
					},
				},
			},
		},
		jwtAuthsByConsumerID: {
			Name: jwtAuthsByConsumerID,
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

// JWTAuthsCollection stores and indexes key-auth credentials.
type JWTAuthsCollection struct {
	memdb *memdb.MemDB
}

// NewJWTAuthsCollection instantiates a JWTAuthsCollection.
func NewJWTAuthsCollection() (*JWTAuthsCollection, error) {
	var schema = &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			jwtAuthTableName: jwtAuthTableSchema,
		},
	}
	m, err := memdb.NewMemDB(schema)
	if err != nil {
		return nil, errors.Wrap(err, "creating new JWTAuthsCollection")
	}
	return &JWTAuthsCollection{
		memdb: m,
	}, nil
}

// Add adds a key-auth credential to JWTAuthsCollection
func (k *JWTAuthsCollection) Add(jwtAuth JWTAuth) error {
	txn := k.memdb.Txn(true)
	defer txn.Abort()
	err := txn.Insert(jwtAuthTableName, &jwtAuth)
	if err != nil {
		return errors.Wrap(err, "insert failed")
	}
	txn.Commit()
	return nil
}

// Get gets a key-auth credential by key or ID.
func (k *JWTAuthsCollection) Get(keyOrID string) (*JWTAuth, error) {
	res, err := multiIndexLookup(k.memdb, jwtAuthTableName,
		[]string{"key", id}, keyOrID)
	if err == ErrNotFound {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, errors.Wrap(err, "jwtAuth lookup failed")
	}
	if res == nil {
		return nil, ErrNotFound
	}
	jwtAuth, ok := res.(*JWTAuth)
	if !ok {
		panic("unexpected type found")
	}
	return &JWTAuth{JWTAuth: *jwtAuth.DeepCopy()}, nil
}

// GetAllByConsumerUsername returns all key-auth credentials
// belong to a Consumer with username.
func (k *JWTAuthsCollection) GetAllByConsumerUsername(username string) ([]*JWTAuth,
	error) {
	txn := k.memdb.Txn(false)
	iter, err := txn.Get(jwtAuthTableName, jwtAuthsByConsumerUsername, username)
	if err != nil {
		return nil, err
	}
	var res []*JWTAuth
	for el := iter.Next(); el != nil; el = iter.Next() {
		r, ok := el.(*JWTAuth)
		if !ok {
			panic("unexpected type found")
		}
		res = append(res, &JWTAuth{JWTAuth: *r.DeepCopy()})
	}
	return res, nil
}

// GetAllByConsumerID returns all key-auth credentials
// belong to a Consumer with id.
func (k *JWTAuthsCollection) GetAllByConsumerID(id string) ([]*JWTAuth,
	error) {
	txn := k.memdb.Txn(false)
	iter, err := txn.Get(jwtAuthTableName, jwtAuthsByConsumerID, id)
	if err != nil {
		return nil, err
	}
	var res []*JWTAuth
	for el := iter.Next(); el != nil; el = iter.Next() {
		r, ok := el.(*JWTAuth)
		if !ok {
			panic("unexpected type found")
		}
		res = append(res, &JWTAuth{JWTAuth: *r.DeepCopy()})
	}
	return res, nil
}

// Update updates an existing key-auth credential.
func (k *JWTAuthsCollection) Update(jwtAuth JWTAuth) error {
	txn := k.memdb.Txn(true)
	defer txn.Abort()
	err := txn.Insert(jwtAuthTableName, &jwtAuth)
	if err != nil {
		return errors.Wrap(err, "update failed")
	}
	txn.Commit()
	return nil
}

// Delete deletes a key-auth credential by key or ID.
func (k *JWTAuthsCollection) Delete(keyOrID string) error {
	jwtAuth, err := k.Get(keyOrID)

	if err != nil {
		return errors.Wrap(err, "looking up jwtAuth")
	}

	txn := k.memdb.Txn(true)
	defer txn.Abort()

	err = txn.Delete(jwtAuthTableName, jwtAuth)
	if err != nil {
		return errors.Wrap(err, "delete failed")
	}
	txn.Commit()
	return nil
}

// GetAll gets all key-auth credentials.
func (k *JWTAuthsCollection) GetAll() ([]*JWTAuth, error) {
	txn := k.memdb.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(jwtAuthTableName, all, true)
	if err != nil {
		return nil, errors.Wrapf(err, "jwtAuth lookup failed")
	}

	var res []*JWTAuth
	for el := iter.Next(); el != nil; el = iter.Next() {
		r, ok := el.(*JWTAuth)
		if !ok {
			panic("unexpected type found")
		}
		res = append(res, &JWTAuth{JWTAuth: *r.DeepCopy()})
	}
	txn.Commit()
	return res, nil
}
