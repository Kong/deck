package state

import (
	"fmt"

	memdb "github.com/hashicorp/go-memdb"
	"github.com/kong/deck/state/indexers"
	"github.com/kong/deck/utils"
)

const (
	allAvailableCertificateTableName = "allavailablecertificate"
)

var allAvailableCertificateTableSchema = &memdb.TableSchema{
	Name: allAvailableCertificateTableName,
	Indexes: map[string]*memdb.IndexSchema{
		"id": {
			Name:    "id",
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
type AllAvailableCertificatesCollection collection

func getAllAvailableCertificate(txn *memdb.Txn, id string) (*Certificate, error) {
	res, err := multiIndexLookupUsingTxn(txn, allAvailableCertificateTableName,
		[]string{"id"}, id)
	if err != nil {
		return nil, err
	}

	c, ok := res.(*Certificate)
	if !ok {
		panic(unexpectedType)
	}
	return &Certificate{Certificate: *c.DeepCopy()}, nil
}

// Get gets a certificate by ID.
func (k *AllAvailableCertificatesCollection) Get(id string) (*Certificate, error) {
	if id == "" {
		return nil, errIDRequired
	}

	txn := k.db.Txn(false)
	defer txn.Abort()
	certificate, err := getAllAvailableCertificate(txn, id)
	if err != nil {
		return nil, err
	}
	return certificate, nil
}

// Add adds a certificate to the collection
func (k *AllAvailableCertificatesCollection) Add(certificate Certificate) error {
	// TODO abstract this check in the go-memdb library itself
	if utils.Empty(certificate.ID) {
		return errIDRequired
	}
	if err := validateCert(certificate); err != nil {
		return err
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	_, err := getAllAvailableCertificate(txn, *certificate.ID)
	if err == nil {
		return fmt.Errorf("inserting certificate %v: %w", certificate.Console(), ErrAlreadyExists)
	} else if err != ErrNotFound {
		return err
	}

	err = txn.Insert(allAvailableCertificateTableName, &certificate)
	if err != nil {
		return err
	}
	txn.Commit()
	return nil
}
