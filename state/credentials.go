package state

import (
	memdb "github.com/hashicorp/go-memdb"
	"github.com/kong/deck/state/indexers"
	"github.com/pkg/errors"
)

const (
	byConsumerID = "byConsumerID"
)

// credentialsCollection stores and indexes key-auth credentials.
type credentialsCollection struct {
	collection
	CredType string
}

func (k *credentialsCollection) TableName() string {
	return k.CredType
}

func (k *credentialsCollection) Schema() *memdb.TableSchema {
	return &memdb.TableSchema{
		Name: k.CredType,
		Indexes: map[string]*memdb.IndexSchema{
			"id": {
				Name:    "id",
				Unique:  true,
				Indexer: &memdb.StringFieldIndex{Field: "ID"},
			},
			byConsumerID: {
				Name: byConsumerID,
				Indexer: &indexers.MethodIndexer{
					Method: "GetConsumer",
				},
			},
			"id2": {
				Name: "id2",
				// TODO configurable
				Unique: true,
				Indexer: &indexers.MethodIndexer{
					Method: "GetID2",
				},
			},
			all: allIndex,
		},
	}
}

func (k *credentialsCollection) getCred(txn *memdb.Txn, IDs ...string) (entity, error) {
	for _, id := range IDs {
		res, err := multiIndexLookupUsingTxn(txn, k.CredType,
			[]string{"id", "id2"}, id)
		if err == ErrNotFound {
			continue
		}
		if err != nil {
			return nil, err
		}
		cred, ok := res.(entity)
		if !ok {
			panic(unexpectedType)
		}
		return cred, nil
	}
	return nil, ErrNotFound
}

// Add adds a key-auth credential to credentialsCollection.
func (k *credentialsCollection) Add(cred entity) error {
	if cred.GetID() == "" {
		return errIDRequired
	}
	txn := k.db.Txn(true)
	defer txn.Abort()

	// TODO detect unique constraint violation for ID2

	_, err := k.getCred(txn, cred.GetID(), cred.GetID2())
	if err == nil {
		return errors.Errorf("credential %v already exists", cred.GetID())
	} else if err != ErrNotFound {
		return err
	}

	err = txn.Insert(k.CredType, cred)
	if err != nil {
		return err
	}
	txn.Commit()
	return nil
}

// Get gets a credential by ID or endpoint key.
func (k *credentialsCollection) Get(id string) (entity, error) {
	if id == "" {
		return nil, errIDRequired
	}

	txn := k.db.Txn(false)
	defer txn.Abort()
	return k.getCred(txn, id)
}

// Update updates an existing key-auth credential.
func (k *credentialsCollection) Update(cred entity) error {
	// TODO abstract this check in the go-memdb library itself
	if cred.GetID() == "" {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := k.deleteCred(txn, cred.GetID())
	if err != nil {
		return err
	}
	err = txn.Insert(k.CredType, cred)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func (k *credentialsCollection) deleteCred(txn *memdb.Txn, nameOrID string) error {
	cred, err := k.getCred(txn, nameOrID)
	if err != nil {
		return err
	}

	err = txn.Delete(k.CredType, cred)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes a key-auth credential by key or ID.
func (k *credentialsCollection) Delete(id string) error {
	if id == "" {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := k.deleteCred(txn, id)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

// GetAll gets all key-auth credentials.
func (k *credentialsCollection) GetAll() ([]entity, error) {
	txn := k.db.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(k.CredType, all, true)
	if err != nil {
		return nil, err
	}

	var res []entity
	for el := iter.Next(); el != nil; el = iter.Next() {
		r, ok := el.(entity)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, r)
	}
	return res, nil
}

// GetAllByConsumerID returns all key-auth credentials
// belong to a Consumer with id.
func (k *credentialsCollection) GetAllByConsumerID(id string) ([]entity,
	error) {
	txn := k.db.Txn(false)
	iter, err := txn.Get(k.CredType, byConsumerID, id)
	if err != nil {
		return nil, err
	}
	var res []entity
	for el := iter.Next(); el != nil; el = iter.Next() {
		r, ok := el.(entity)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, r)
	}
	return res, nil
}
