package state

import (
	memdb "github.com/hashicorp/go-memdb"
	"github.com/hbagdi/deck/state/indexers"
	"github.com/pkg/errors"
)

const (
	basicAuthTableName           = "basicAuth"
	basicAuthsByConsumerUsername = "basicAuthsByConsumerUsername"
	basicAuthsByConsumerID       = "basicAuthsByConsumerID"
)

var basicAuthTableSchema = &memdb.TableSchema{
	Name: basicAuthTableName,
	Indexes: map[string]*memdb.IndexSchema{
		id: {
			Name:    id,
			Unique:  true,
			Indexer: &memdb.StringFieldIndex{Field: "ID"},
		},
		basicAuthsByConsumerUsername: {
			Name: basicAuthsByConsumerUsername,
			Indexer: &indexers.SubFieldIndexer{
				Fields: []indexers.Field{
					{
						Struct: "Consumer",
						Sub:    "Username",
					},
				},
			},
		},
		basicAuthsByConsumerID: {
			Name: basicAuthsByConsumerID,
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

// BasicAuthsCollection stores and indexes basic-auth credentials.
type BasicAuthsCollection struct {
	memdb *memdb.MemDB
}

// NewBasicAuthsCollection instantiates a BasicAuthsCollection.
func NewBasicAuthsCollection() (*BasicAuthsCollection, error) {
	var schema = &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			basicAuthTableName: basicAuthTableSchema,
		},
	}
	m, err := memdb.NewMemDB(schema)
	if err != nil {
		return nil, errors.Wrap(err, "creating new BasicAuthsCollection")
	}
	return &BasicAuthsCollection{
		memdb: m,
	}, nil
}

// Add adds a basic-auth credential to BasicAuthsCollection
func (k *BasicAuthsCollection) Add(basicAuth BasicAuth) error {
	txn := k.memdb.Txn(true)
	defer txn.Abort()
	err := txn.Insert(basicAuthTableName, &basicAuth)
	if err != nil {
		return errors.Wrap(err, "insert failed")
	}
	txn.Commit()
	return nil
}

// Get gets a basic-auth credential by  or ID.
func (k *BasicAuthsCollection) Get(usernameOrID string) (*BasicAuth, error) {
	res, err := multiIndexLookup(k.memdb, basicAuthTableName,
		[]string{"username", id}, usernameOrID)
	if err == ErrNotFound {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, errors.Wrap(err, "basicAuth lookup failed")
	}
	if res == nil {
		return nil, ErrNotFound
	}
	basicAuth, ok := res.(*BasicAuth)
	if !ok {
		panic("unexpected type found")
	}
	return &BasicAuth{BasicAuth: *basicAuth.DeepCopy()}, nil
}

// GetAllByConsumerUsername returns all basic-auth credentials
// belong to a Consumer with username.
func (k *BasicAuthsCollection) GetAllByConsumerUsername(username string) ([]*BasicAuth,
	error) {
	txn := k.memdb.Txn(false)
	iter, err := txn.Get(basicAuthTableName, basicAuthsByConsumerUsername, username)
	if err != nil {
		return nil, err
	}
	var res []*BasicAuth
	for el := iter.Next(); el != nil; el = iter.Next() {
		r, ok := el.(*BasicAuth)
		if !ok {
			panic("unexpected type found")
		}
		res = append(res, &BasicAuth{BasicAuth: *r.DeepCopy()})
	}
	return res, nil
}

// GetAllByConsumerID returns all basic-auth credentials
// belong to a Consumer with id.
func (k *BasicAuthsCollection) GetAllByConsumerID(id string) ([]*BasicAuth,
	error) {
	txn := k.memdb.Txn(false)
	iter, err := txn.Get(basicAuthTableName, basicAuthsByConsumerID, id)
	if err != nil {
		return nil, err
	}
	var res []*BasicAuth
	for el := iter.Next(); el != nil; el = iter.Next() {
		r, ok := el.(*BasicAuth)
		if !ok {
			panic("unexpected type found")
		}
		res = append(res, &BasicAuth{BasicAuth: *r.DeepCopy()})
	}
	return res, nil
}

// Update updates an existing basic-auth credential.
func (k *BasicAuthsCollection) Update(basicAuth BasicAuth) error {
	txn := k.memdb.Txn(true)
	defer txn.Abort()
	err := txn.Insert(basicAuthTableName, &basicAuth)
	if err != nil {
		return errors.Wrap(err, "update failed")
	}
	txn.Commit()
	return nil
}

// Delete deletes a basic-auth credential by key or ID.
func (k *BasicAuthsCollection) Delete(usernameOrID string) error {
	basicAuth, err := k.Get(usernameOrID)

	if err != nil {
		return errors.Wrap(err, "looking up basicAuth")
	}

	txn := k.memdb.Txn(true)
	defer txn.Abort()

	err = txn.Delete(basicAuthTableName, basicAuth)
	if err != nil {
		return errors.Wrap(err, "delete failed")
	}
	txn.Commit()
	return nil
}

// GetAll gets all basic-auth credentials.
func (k *BasicAuthsCollection) GetAll() ([]*BasicAuth, error) {
	txn := k.memdb.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(basicAuthTableName, all, true)
	if err != nil {
		return nil, errors.Wrapf(err, "basicAuth lookup failed")
	}

	var res []*BasicAuth
	for el := iter.Next(); el != nil; el = iter.Next() {
		r, ok := el.(*BasicAuth)
		if !ok {
			panic("unexpected type found")
		}
		res = append(res, &BasicAuth{BasicAuth: *r.DeepCopy()})
	}
	txn.Commit()
	return res, nil
}
