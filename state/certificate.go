package state

import (
	memdb "github.com/hashicorp/go-memdb"
	"github.com/kong/deck/state/indexers"
	"github.com/kong/deck/utils"
	"github.com/pkg/errors"
)

const (
	certificateTableName = "certificate"
)

var certificateTableSchema = &memdb.TableSchema{
	Name: certificateTableName,
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

func validateCert(certificate Certificate) error {
	if utils.Empty(certificate.Key) {
		return errors.New("certificate's Key cannot be empty")
	}
	if utils.Empty(certificate.Cert) {
		return errors.New("certificate's Cert cannot be empty")
	}
	return nil
}

// CertificatesCollection stores and indexes Kong Certificates.
type CertificatesCollection collection

// Add adds a certificate to the collection
func (k *CertificatesCollection) Add(certificate Certificate) error {
	// TODO abstract this check in the go-memdb library itself
	if utils.Empty(certificate.ID) {
		return errIDRequired
	}
	if err := validateCert(certificate); err != nil {
		return err
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	_, err := getCertificate(txn, *certificate.ID)
	if err == nil {
		return errors.Errorf("certificate %v already exists", certificate.Console())
	} else if err != ErrNotFound {
		return err
	}

	err = txn.Insert(certificateTableName, &certificate)
	if err != nil {
		return err
	}
	txn.Commit()
	return nil
}

func getCertificate(txn *memdb.Txn, id string) (*Certificate, error) {
	res, err := multiIndexLookupUsingTxn(txn, certificateTableName,
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
func (k *CertificatesCollection) Get(id string) (*Certificate, error) {
	if id == "" {
		return nil, errIDRequired
	}

	txn := k.db.Txn(false)
	defer txn.Abort()
	certificate, err := getCertificate(txn, id)
	if err != nil {
		return nil, err
	}
	return certificate, nil
}

func getCertificateByCertKey(txn *memdb.Txn, cert, key string) (*Certificate, error) {
	res, err := txn.First(certificateTableName, "certkey", cert, key)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, ErrNotFound
	}
	c, ok := res.(*Certificate)
	if !ok {
		panic(unexpectedType)
	}
	return &Certificate{Certificate: *c.DeepCopy()}, nil
}

// GetByCertKey gets a certificate with
// the same key and cert from the collection.
func (k *CertificatesCollection) GetByCertKey(cert,
	key string) (*Certificate, error) {
	if cert == "" || key == "" {
		return nil, errors.New("cert/key cannot be empty string")
	}

	txn := k.db.Txn(false)
	defer txn.Abort()

	return getCertificateByCertKey(txn, cert, key)
}

// Update udpates an existing certificate.
// It returns an error if the certificate is not already present.
func (k *CertificatesCollection) Update(certificate Certificate) error {
	if utils.Empty(certificate.ID) {
		return errIDRequired
	}
	if err := validateCert(certificate); err != nil {
		return err
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteCertificate(txn, *certificate.ID)
	if err != nil {
		return err
	}

	err = txn.Insert(certificateTableName, &certificate)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func deleteCertificate(txn *memdb.Txn, id string) error {
	cert, err := getCertificate(txn, id)
	if err != nil {
		return err
	}

	err = txn.Delete(certificateTableName, cert)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes a certificate by ID.
func (k *CertificatesCollection) Delete(id string) error {
	if id == "" {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteCertificate(txn, id)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

// DeleteByCertKey deletes a certificate by looking up it's cert and key.
func (k *CertificatesCollection) DeleteByCertKey(cert, key string) error {
	if cert == "" || key == "" {
		return errors.New("cert/key cannot be empty string")
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	certificate, err := getCertificateByCertKey(txn, cert, key)
	if err != nil {
		return err
	}
	err = deleteCertificate(txn, *certificate.ID)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

// GetAll gets a certificate by name or ID.
func (k *CertificatesCollection) GetAll() ([]*Certificate, error) {
	txn := k.db.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(certificateTableName, all, true)
	if err != nil {
		return nil, err
	}

	var res []*Certificate
	for el := iter.Next(); el != nil; el = iter.Next() {
		c, ok := el.(*Certificate)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &Certificate{Certificate: *c.DeepCopy()})
	}
	txn.Commit()
	return res, nil
}
