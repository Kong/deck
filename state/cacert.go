package state

import (
	memdb "github.com/hashicorp/go-memdb"
	"github.com/hbagdi/deck/state/indexers"
	"github.com/pkg/errors"
)

const (
	caCertTableName = "caCert"
)

var caCertTableSchema = &memdb.TableSchema{
	Name: caCertTableName,
	Indexes: map[string]*memdb.IndexSchema{
		id: {
			Name:    id,
			Unique:  true,
			Indexer: &memdb.StringFieldIndex{Field: "ID"},
		},
		"cert": {
			Name:   "cert",
			Unique: true,
			Indexer: &indexers.MD5FieldsIndexer{
				Fields: []string{"Cert"},
			},
		},
		all: allIndex,
	},
}

// CACertificatesCollection stores and indexes Kong CACertificates.
type CACertificatesCollection struct {
	memdb *memdb.MemDB
}

// NewCACertificatesCollection instantiates a CACertificatesCollection.
func NewCACertificatesCollection() (*CACertificatesCollection, error) {
	var schema = &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			caCertTableName: caCertTableSchema,
		},
	}
	m, err := memdb.NewMemDB(schema)
	if err != nil {
		return nil, errors.Wrap(err, "creating new CACertificateCollection")
	}
	return &CACertificatesCollection{
		memdb: m,
	}, nil
}

// Add adds a caCert to the collection
func (k *CACertificatesCollection) Add(caCert CACertificate) error {
	txn := k.memdb.Txn(true)
	defer txn.Abort()
	err := txn.Insert(caCertTableName, &caCert)
	if err != nil {
		return errors.Wrap(err, "insert failed")
	}
	txn.Commit()
	return nil
}

// Get gets a caCertificate by cert or ID.
func (k *CACertificatesCollection) Get(certOrID string) (*CACertificate, error) {
	res, err := multiIndexLookup(k.memdb, caCertTableName,
		[]string{"id", "cert"}, certOrID)
	if err == ErrNotFound {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, errors.Wrap(err, "caCertificate lookup failed")
	}
	if res == nil {
		return nil, ErrNotFound
	}
	c, ok := res.(*CACertificate)
	if !ok {
		panic("unexpected type found")
	}
	return &CACertificate{CACertificate: *c.DeepCopy()}, nil
}

// Update udpates an existing caCert.
// It returns an error if the caCert is not already present.
func (k *CACertificatesCollection) Update(caCert CACertificate) error {
	// TODO check if entity is already present or not, throw error if present
	// TODO abstract this in the go-memdb library itself
	txn := k.memdb.Txn(true)
	defer txn.Abort()
	err := txn.Insert(caCertTableName, &caCert)
	if err != nil {
		return errors.Wrap(err, "update failed")
	}
	txn.Commit()
	return nil
}

// Delete deletes a caCertificate by looking up it's cert and key.
func (k *CACertificatesCollection) Delete(certOrID string) error {
	caCert, err := k.Get(certOrID)

	if err != nil {
		return errors.Wrap(err, "looking up caCert")
	}

	txn := k.memdb.Txn(true)
	defer txn.Abort()

	err = txn.Delete(caCertTableName, caCert)
	if err != nil {
		return errors.Wrap(err, "delete failed")
	}
	txn.Commit()
	return nil
}

// GetAll gets a caCertificate by name or ID.
func (k *CACertificatesCollection) GetAll() ([]*CACertificate, error) {
	txn := k.memdb.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(caCertTableName, all, true)
	if err != nil {
		return nil, errors.Wrapf(err, "caCertificate lookup failed")
	}

	var res []*CACertificate
	for el := iter.Next(); el != nil; el = iter.Next() {
		c, ok := el.(*CACertificate)
		if !ok {
			panic("unexpected type found")
		}
		res = append(res, &CACertificate{CACertificate: *c.DeepCopy()})
	}
	txn.Commit()
	return res, nil
}
