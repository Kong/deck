package state

import (
	"fmt"

	memdb "github.com/hashicorp/go-memdb"

	"github.com/kong/deck/state/indexers"
	"github.com/kong/deck/utils"
)

const (
	caCertTableName = "caCert"
)

var caCertTableSchema = &memdb.TableSchema{
	Name: caCertTableName,
	Indexes: map[string]*memdb.IndexSchema{
		"id": {
			Name:    "id",
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
type CACertificatesCollection collection

// Add adds a caCert to the collection
func (k *CACertificatesCollection) Add(caCert CACertificate) error {
	// TODO abstract this check in the go-memdb library itself
	if utils.Empty(caCert.ID) {
		return errIDRequired
	}
	txn := k.db.Txn(true)
	defer txn.Abort()

	var searchBy []string
	searchBy = append(searchBy, *caCert.ID)
	if !utils.Empty(caCert.Cert) {
		searchBy = append(searchBy, *caCert.Cert)
	}
	_, err := getCACert(txn, searchBy...)
	if err == nil {
		return fmt.Errorf("inserting ca-cert %v: %w", caCert.Console(), ErrAlreadyExists)
	} else if err != ErrNotFound {
		return err
	}

	err = txn.Insert(caCertTableName, &caCert)
	if err != nil {
		return err
	}
	txn.Commit()
	return nil
}

func getCACert(txn *memdb.Txn, IDs ...string) (*CACertificate, error) {
	for _, id := range IDs {
		res, err := multiIndexLookupUsingTxn(txn, caCertTableName,
			[]string{"cert", "id"}, id)
		if err == ErrNotFound {
			continue
		}
		if err != nil {
			return nil, err
		}
		caCert, ok := res.(*CACertificate)
		if !ok {
			panic(unexpectedType)
		}
		return &CACertificate{CACertificate: *caCert.DeepCopy()}, nil
	}
	return nil, ErrNotFound
}

// Get gets a caCertificate by cert or ID.
func (k *CACertificatesCollection) Get(certOrID string) (*CACertificate, error) {
	if certOrID == "" {
		return nil, errIDRequired
	}

	txn := k.db.Txn(false)
	defer txn.Abort()
	return getCACert(txn, certOrID)
}

// Update udpates an existing caCert.
// It returns an error if the caCert is not already present.
func (k *CACertificatesCollection) Update(caCert CACertificate) error {
	// TODO abstract this check in the go-memdb library itself
	if utils.Empty(caCert.ID) {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteCACert(txn, *caCert.ID)
	if err != nil {
		return err
	}

	err = txn.Insert(caCertTableName, &caCert)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func deleteCACert(txn *memdb.Txn, certOrID string) error {
	caCert, err := getCACert(txn, certOrID)
	if err != nil {
		return err
	}

	err = txn.Delete(caCertTableName, caCert)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes a caCertificate by looking up it's cert and key.
func (k *CACertificatesCollection) Delete(certOrID string) error {
	if certOrID == "" {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteCACert(txn, certOrID)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

// GetAll gets a caCertificate by name or ID.
func (k *CACertificatesCollection) GetAll() ([]*CACertificate, error) {
	txn := k.db.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(caCertTableName, all, true)
	if err != nil {
		return nil, err
	}

	var res []*CACertificate
	for el := iter.Next(); el != nil; el = iter.Next() {
		c, ok := el.(*CACertificate)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &CACertificate{CACertificate: *c.DeepCopy()})
	}
	txn.Commit()
	return res, nil
}
