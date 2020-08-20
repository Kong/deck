package state

import (
	memdb "github.com/hashicorp/go-memdb"
	"github.com/kong/deck/state/indexers"
	"github.com/kong/deck/utils"
	"github.com/pkg/errors"
)

const (
	sniTableName = "sni"
	snisByCertID = "snisByCertID"
)

var errInvalidCert = errors.New("certificate.ID is required in sni")

var sniTableSchema = &memdb.TableSchema{
	Name: sniTableName,
	Indexes: map[string]*memdb.IndexSchema{
		"id": {
			Name:    "id",
			Unique:  true,
			Indexer: &memdb.StringFieldIndex{Field: "ID"},
		},
		"name": {
			Name:         "name",
			Unique:       true,
			Indexer:      &memdb.StringFieldIndex{Field: "Name"},
			AllowMissing: true,
		},
		all: allIndex,
		// foreign
		snisByCertID: {
			Name: snisByCertID,
			Indexer: &indexers.SubFieldIndexer{
				Fields: []indexers.Field{
					{
						Struct: "Certificate",
						Sub:    "ID",
					},
				},
			},
		},
	},
}

func validateCertForSNI(sni *SNI) error {
	if sni.Certificate == nil ||
		utils.Empty(sni.Certificate.ID) {
		return errInvalidCert
	}
	return nil
}

// SNIsCollection stores and indexes Kong SNIs.
type SNIsCollection collection

// Add adds a sni into SNIsCollection
// sni.ID should not be nil else an error is thrown.
func (k *SNIsCollection) Add(sni SNI) error {
	// TODO abstract this check in the go-memdb library itself
	if utils.Empty(sni.ID) {
		return errIDRequired
	}

	if err := validateCertForSNI(&sni); err != nil {
		return err
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	var searchBy []string
	searchBy = append(searchBy, *sni.ID)
	if !utils.Empty(sni.Name) {
		searchBy = append(searchBy, *sni.Name)
	}
	_, err := getSNI(txn, searchBy...)
	if err == nil {
		return errors.Errorf("sni %v already exists", sni.Console())
	} else if err != ErrNotFound {
		return err
	}

	err = txn.Insert(sniTableName, &sni)
	if err != nil {
		return err
	}
	txn.Commit()
	return nil
}

func getSNI(txn *memdb.Txn, IDs ...string) (*SNI, error) {
	for _, id := range IDs {
		res, err := multiIndexLookupUsingTxn(txn, sniTableName,
			[]string{"name", "id"}, id)
		if err == ErrNotFound {
			continue
		}
		if err != nil {
			return nil, err
		}

		sni, ok := res.(*SNI)
		if !ok {
			panic(unexpectedType)
		}
		return &SNI{SNI: *sni.DeepCopy()}, nil
	}
	return nil, ErrNotFound
}

// Get gets a sni by name or ID.
func (k *SNIsCollection) Get(nameOrID string) (*SNI, error) {
	if nameOrID == "" {
		return nil, errIDRequired
	}

	txn := k.db.Txn(false)
	defer txn.Abort()
	sni, err := getSNI(txn, nameOrID)
	if err != nil {
		return nil, err
	}
	return sni, nil
}

// Update updates a sni
func (k *SNIsCollection) Update(sni SNI) error {
	// TODO abstract this check in the go-memdb library itself
	if utils.Empty(sni.ID) {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteSNI(txn, *sni.ID)
	if err != nil {
		return err
	}

	err = txn.Insert(sniTableName, &sni)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func deleteSNI(txn *memdb.Txn, nameOrID string) error {
	sni, err := getSNI(txn, nameOrID)
	if err != nil {
		return err
	}

	err = txn.Delete(sniTableName, sni)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes a sni by name or ID.
func (k *SNIsCollection) Delete(nameOrID string) error {
	if nameOrID == "" {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteSNI(txn, nameOrID)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

// GetAll gets a sni by name or ID.
func (k *SNIsCollection) GetAll() ([]*SNI, error) {
	txn := k.db.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(sniTableName, all, true)
	if err != nil {
		return nil, err
	}

	var res []*SNI
	for el := iter.Next(); el != nil; el = iter.Next() {
		r, ok := el.(*SNI)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &SNI{SNI: *r.DeepCopy()})
	}
	txn.Commit()
	return res, nil
}

// GetAllByCertID returns all routes referencing a service
// by its id.
func (k *SNIsCollection) GetAllByCertID(id string) ([]*SNI,
	error) {
	txn := k.db.Txn(false)
	iter, err := txn.Get(sniTableName, snisByCertID, id)
	if err != nil {
		return nil, err
	}
	var res []*SNI
	for el := iter.Next(); el != nil; el = iter.Next() {
		r, ok := el.(*SNI)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &SNI{SNI: *r.DeepCopy()})
	}
	return res, nil
}
