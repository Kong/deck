package state

import (
	"fmt"

	memdb "github.com/hashicorp/go-memdb"
	"github.com/kong/deck/utils"
)

const (
	licenseTableName = "license"
)

var licenseTableSchema = &memdb.TableSchema{
	Name: licenseTableName,
	Indexes: map[string]*memdb.IndexSchema{
		"id": {
			Name:    "id",
			Unique:  true,
			Indexer: &memdb.StringFieldIndex{Field: "ID"},
		},
		all: allIndex,
	},
}

// LicensesCollection stores and indexes Kong Licenses.
type LicensesCollection collection

// Add adds a license to the collection.
// license.ID should not be nil else an error is thrown.
func (k *LicensesCollection) Add(license License) error {
	if utils.Empty(license.ID) {
		return errIDRequired
	}
	txn := k.db.Txn(true)
	defer txn.Abort()

	var searchBy []string
	searchBy = append(searchBy, *license.ID)
	_, err := getLicense(txn, searchBy...)
	if err == nil {
		return fmt.Errorf("inserting license %v: %w", license.Console(), ErrAlreadyExists)
	} else if err != ErrNotFound {
		return err
	}

	err = txn.Insert(licenseTableName, &license)
	if err != nil {
		return err
	}
	txn.Commit()
	return nil
}

func getLicense(txn *memdb.Txn, IDs ...string) (*License, error) {
	for _, id := range IDs {
		res, err := multiIndexLookupUsingTxn(txn, licenseTableName,
			[]string{"id"}, id)
		if err == ErrNotFound {
			continue
		}
		if err != nil {
			return nil, err
		}

		license, ok := res.(*License)
		if !ok {
			panic(unexpectedType)
		}
		return &License{License: *license.DeepCopy()}, nil
	}
	return nil, ErrNotFound
}

// Get gets a license by ID.
func (k *LicensesCollection) Get(ID string) (*License, error) {
	if ID == "" {
		return nil, errIDRequired
	}

	txn := k.db.Txn(false)
	defer txn.Abort()
	license, err := getLicense(txn, ID)
	if err != nil {
		if err == ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return license, nil
}

// Update udpates an existing license.
func (k *LicensesCollection) Update(license License) error {
	if utils.Empty(license.ID) {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteLicense(txn, *license.ID)
	if err != nil {
		return err
	}

	err = txn.Insert(licenseTableName, &license)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func deleteLicense(txn *memdb.Txn, nameOrID string) error {
	license, err := getLicense(txn, nameOrID)
	if err != nil {
		return err
	}

	err = txn.Delete(licenseTableName, license)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes a license by its ID.
func (k *LicensesCollection) Delete(ID string) error {
	if ID == "" {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteLicense(txn, ID)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

// GetAll gets all licenses in the state.
func (k *LicensesCollection) GetAll() ([]*License, error) {
	txn := k.db.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(licenseTableName, all, true)
	if err != nil {
		return nil, err
	}

	var res []*License
	for el := iter.Next(); el != nil; el = iter.Next() {
		l, ok := el.(*License)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &License{License: *l.DeepCopy()})
	}
	txn.Commit()
	return res, nil
}
