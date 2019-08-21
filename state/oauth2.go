package state

import (
	memdb "github.com/hashicorp/go-memdb"
	"github.com/hbagdi/deck/state/indexers"
	"github.com/pkg/errors"
)

const (
	oauth2TableName               = "oauth2Cred"
	oauth2CredsByConsumerUsername = "oauth2CredsByConsumerUsername"
	oauth2CredsByConsumerID       = "oauth2CredsByConsumerID"
)

var oauth2CredTableSchema = &memdb.TableSchema{
	Name: oauth2TableName,
	Indexes: map[string]*memdb.IndexSchema{
		id: {
			Name:    id,
			Unique:  true,
			Indexer: &memdb.StringFieldIndex{Field: "ID"},
		},
		oauth2CredsByConsumerUsername: {
			Name: oauth2CredsByConsumerUsername,
			Indexer: &indexers.SubFieldIndexer{
				Fields: []indexers.Field{
					{
						Struct: "Consumer",
						Sub:    "Username",
					},
				},
			},
		},
		oauth2CredsByConsumerID: {
			Name: oauth2CredsByConsumerID,
			Indexer: &indexers.SubFieldIndexer{
				Fields: []indexers.Field{
					{
						Struct: "Consumer",
						Sub:    "ID",
					},
				},
			},
		},
		"ClientID": {
			Name:    "ClientID",
			Unique:  true,
			Indexer: &memdb.StringFieldIndex{Field: "ClientID"},
		},
		all: allIndex,
	},
}

// Oauth2CredsCollection stores and indexes oauth2 credentials.
type Oauth2CredsCollection struct {
	memdb *memdb.MemDB
}

// NewOauth2CredsCollection instantiates a Oauth2CredsCollection.
func NewOauth2CredsCollection() (*Oauth2CredsCollection, error) {
	var schema = &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			oauth2TableName: oauth2CredTableSchema,
		},
	}
	m, err := memdb.NewMemDB(schema)
	if err != nil {
		return nil, errors.Wrap(err, "creating new Oauth2CredsCollection")
	}
	return &Oauth2CredsCollection{
		memdb: m,
	}, nil
}

// Add adds an oauth2 credential to Oauth2CredsCollection
func (k *Oauth2CredsCollection) Add(oauth2Cred Oauth2Credential) error {
	txn := k.memdb.Txn(true)
	defer txn.Abort()
	err := txn.Insert(oauth2TableName, &oauth2Cred)
	if err != nil {
		return errors.Wrap(err, "insert failed")
	}
	txn.Commit()
	return nil
}

// Get gets an oauth2 credential by client_id or ID.
func (k *Oauth2CredsCollection) Get(clientIDorID string) (*Oauth2Credential, error) {
	res, err := multiIndexLookup(k.memdb, oauth2TableName,
		[]string{"ClientID", id}, clientIDorID)
	if err == ErrNotFound {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, errors.Wrap(err, "oauth2Cred lookup failed")
	}
	if res == nil {
		return nil, ErrNotFound
	}
	oauth2Cred, ok := res.(*Oauth2Credential)
	if !ok {
		panic("unexpected type found")
	}
	return &Oauth2Credential{Oauth2Credential: *oauth2Cred.DeepCopy()}, nil
}

// GetAllByConsumerUsername returns all oauth2 credentials
// belong to a Consumer with username.
func (k *Oauth2CredsCollection) GetAllByConsumerUsername(
	username string) ([]*Oauth2Credential, error) {
	txn := k.memdb.Txn(false)
	iter, err := txn.Get(oauth2TableName, oauth2CredsByConsumerUsername,
		username)
	if err != nil {
		return nil, err
	}
	var res []*Oauth2Credential
	for el := iter.Next(); el != nil; el = iter.Next() {
		r, ok := el.(*Oauth2Credential)
		if !ok {
			panic("unexpected type found")
		}
		res = append(res, &Oauth2Credential{Oauth2Credential: *r.DeepCopy()})
	}
	return res, nil
}

// GetAllByConsumerID returns all oauth2 credentials
// belong to a Consumer with id.
func (k *Oauth2CredsCollection) GetAllByConsumerID(id string) ([]*Oauth2Credential,
	error) {
	txn := k.memdb.Txn(false)
	iter, err := txn.Get(oauth2TableName, oauth2CredsByConsumerID, id)
	if err != nil {
		return nil, err
	}
	var res []*Oauth2Credential
	for el := iter.Next(); el != nil; el = iter.Next() {
		r, ok := el.(*Oauth2Credential)
		if !ok {
			panic("unexpected type found")
		}
		res = append(res, &Oauth2Credential{Oauth2Credential: *r.DeepCopy()})
	}
	return res, nil
}

// Update updates an existing oauth2 credential.
func (k *Oauth2CredsCollection) Update(oauth2Cred Oauth2Credential) error {
	txn := k.memdb.Txn(true)
	defer txn.Abort()
	err := txn.Insert(oauth2TableName, &oauth2Cred)
	if err != nil {
		return errors.Wrap(err, "update failed")
	}
	txn.Commit()
	return nil
}

// Delete deletes an oauth2 credential by client_id or ID.
func (k *Oauth2CredsCollection) Delete(clientIDorID string) error {
	oauth2Cred, err := k.Get(clientIDorID)

	if err != nil {
		return errors.Wrap(err, "looking up oauth2Cred")
	}

	txn := k.memdb.Txn(true)
	defer txn.Abort()

	err = txn.Delete(oauth2TableName, oauth2Cred)
	if err != nil {
		return errors.Wrap(err, "delete failed")
	}
	txn.Commit()
	return nil
}

// GetAll gets all oauth2 credentials.
func (k *Oauth2CredsCollection) GetAll() ([]*Oauth2Credential, error) {
	txn := k.memdb.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(oauth2TableName, all, true)
	if err != nil {
		return nil, errors.Wrapf(err, "oauth2Cred lookup failed")
	}

	var res []*Oauth2Credential
	for el := iter.Next(); el != nil; el = iter.Next() {
		r, ok := el.(*Oauth2Credential)
		if !ok {
			panic("unexpected type found")
		}
		res = append(res, &Oauth2Credential{Oauth2Credential: *r.DeepCopy()})
	}
	txn.Commit()
	return res, nil
}
