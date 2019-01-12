package state

import (
	memdb "github.com/hashicorp/go-memdb"
	"github.com/hbagdi/deck/state/indexers"
	"github.com/pkg/errors"
)

const (
	certificateTableName = "certificate"
)

var certificateTableSchema = &memdb.TableSchema{
	Name: certificateTableName,
	Indexes: map[string]*memdb.IndexSchema{
		id: {
			Name:    id,
			Unique:  true,
			Indexer: &memdb.StringFieldIndex{Field: "ID"},
		},
		"certkey": {
			Name:   "certkey",
			Unique: true,
			Indexer: &indexers.MD5FieldsIndexer{
				Fields: []string{"Cert", "Key"},
			},
		},
		all: allIndex,
	},
}

// CertificatesCollection stores and indexes Kong Certificates.
type CertificatesCollection struct {
	memdb *memdb.MemDB
}

// NewCertificatesCollection instantiates a CertificatesCollection.
func NewCertificatesCollection() (*CertificatesCollection, error) {
	var schema = &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			certificateTableName: certificateTableSchema,
		},
	}
	m, err := memdb.NewMemDB(schema)
	if err != nil {
		return nil, errors.Wrap(err, "creating new CertificateCollection")
	}
	return &CertificatesCollection{
		memdb: m,
	}, nil
}

// Add adds a certificate to the collection
func (k *CertificatesCollection) Add(certificate Certificate) error {
	txn := k.memdb.Txn(true)
	defer txn.Abort()
	err := txn.Insert(certificateTableName, &certificate)
	if err != nil {
		return errors.Wrap(err, "insert failed")
	}
	txn.Commit()
	return nil
}

// Get gets a certificate by name or ID.
func (k *CertificatesCollection) Get(id string) (*Certificate, error) {
	res, err := multiIndexLookup(k.memdb, certificateTableName,
		[]string{"id"}, id)
	if err == ErrNotFound {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, errors.Wrap(err, "certificate lookup failed")
	}
	if res == nil {
		return nil, ErrNotFound
	}
	c, ok := res.(*Certificate)
	if !ok {
		panic("unexpected type found")
	}
	return &Certificate{Certificate: *c.DeepCopy()}, nil
}

// GetByCertKey gets a certificate with
// the same key and cert from the collection.
func (k *CertificatesCollection) GetByCertKey(cert,
	key string) (*Certificate, error) {
	txn := k.memdb.Txn(false)
	defer txn.Abort()

	res, err := txn.First(certificateTableName, "certkey", cert, key)
	if err == ErrNotFound {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, errors.Wrap(err, "certificate lookup failed")
	}
	if res == nil {
		return nil, ErrNotFound
	}
	c, ok := res.(*Certificate)
	if !ok {
		panic("unexpected type found")
	}
	return &Certificate{Certificate: *c.DeepCopy()}, nil
}

// Update udpates an exisitng certificate.
// It returns an error if the certificate is not already present.
func (k *CertificatesCollection) Update(certificate Certificate) error {
	// TODO check if entity is already present or not, throw error if present
	// TODO abstract this in the go-memdb library itself
	txn := k.memdb.Txn(true)
	defer txn.Abort()
	err := txn.Insert(certificateTableName, &certificate)
	if err != nil {
		return errors.Wrap(err, "update failed")
	}
	txn.Commit()
	return nil
}

// Delete deletes a certificate by ID.
func (k *CertificatesCollection) Delete(ID string) error {
	certificate, err := k.Get(ID)

	if err != nil {
		return errors.Wrap(err, "looking up certificate")
	}

	txn := k.memdb.Txn(true)
	defer txn.Abort()

	err = txn.Delete(certificateTableName, certificate)
	if err != nil {
		return errors.Wrap(err, "delete failed")
	}
	txn.Commit()
	return nil
}

// DeleteByCertKey deletes a certificate by looking up it's cert and key.
func (k *CertificatesCollection) DeleteByCertKey(cert, key string) error {
	certificate, err := k.GetByCertKey(cert, key)

	if err != nil {
		return errors.Wrap(err, "looking up certificate")
	}

	txn := k.memdb.Txn(true)
	defer txn.Abort()

	err = txn.Delete(certificateTableName, certificate)
	if err != nil {
		return errors.Wrap(err, "delete failed")
	}
	txn.Commit()
	return nil
}

// GetAll gets a certificate by name or ID.
func (k *CertificatesCollection) GetAll() ([]*Certificate, error) {
	txn := k.memdb.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(certificateTableName, all, true)
	if err != nil {
		return nil, errors.Wrapf(err, "certificate lookup failed")
	}

	var res []*Certificate
	for el := iter.Next(); el != nil; el = iter.Next() {
		c, ok := el.(*Certificate)
		if !ok {
			panic("unexpected type found")
		}
		res = append(res, &Certificate{Certificate: *c.DeepCopy()})
	}
	txn.Commit()
	return res, nil
}
