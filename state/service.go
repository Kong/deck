package state

import (
	memdb "github.com/hashicorp/go-memdb"
	"github.com/pkg/errors"
)

const (
	serviceTableName = "service"
)

var serviceTableSchema = &memdb.TableSchema{
	Name: serviceTableName,
	Indexes: map[string]*memdb.IndexSchema{
		id: {
			Name:   id,
			Unique: true,
			// EnforceUnique: true,
			Indexer: &memdb.StringFieldIndex{Field: "ID"},
		},
		"name": {
			Name:   "name",
			Unique: true,
			// EnforceUnique: true,
			Indexer: &memdb.StringFieldIndex{Field: "Name"},
		},
		all: allIndex,
	},
}

// ServicesCollection stores and indexes Kong Services.
type ServicesCollection collection

// Add adds a service to the collection
func (k *ServicesCollection) Add(service Service) error {
	txn := k.db.Txn(true)
	defer txn.Abort()
	err := txn.Insert(serviceTableName, &service)
	if err != nil {
		return errors.Wrap(err, "insert failed")
	}
	txn.Commit()
	return nil
}

// Get gets a service by name or ID.
func (k *ServicesCollection) Get(nameOrID string) (*Service, error) {
	res, err := multiIndexLookup(k.db, serviceTableName,
		[]string{"name", id}, nameOrID)
	if err == ErrNotFound {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, errors.Wrap(err, "service lookup failed")
	}
	if res == nil {
		return nil, ErrNotFound
	}
	service, ok := res.(*Service)
	if !ok {
		panic("unexpected type found")
	}

	return &Service{Service: *service.DeepCopy()}, nil
}

// Update udpates an existing service.
// It returns an error if the service is not already present.
func (k *ServicesCollection) Update(service Service) error {
	// TODO check if entity is already present or not, throw error if present
	// TODO abstract this in the go-memdb library itself
	txn := k.db.Txn(true)
	defer txn.Abort()
	err := txn.Insert(serviceTableName, &service)
	if err != nil {
		return errors.Wrap(err, "update failed")
	}
	txn.Commit()
	return nil
}

// Delete deletes a service by name or ID.
func (k *ServicesCollection) Delete(nameOrID string) error {
	service, err := k.Get(nameOrID)

	if err != nil {
		return errors.Wrap(err, "looking up service")
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err = txn.Delete(serviceTableName, service)
	if err != nil {
		return errors.Wrap(err, "delete failed")
	}
	txn.Commit()
	return nil
}

// GetAll gets a service by name or ID.
func (k *ServicesCollection) GetAll() ([]*Service, error) {
	txn := k.db.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(serviceTableName, all, true)
	if err != nil {
		return nil, errors.Wrapf(err, "service lookup failed")
	}

	var res []*Service
	for el := iter.Next(); el != nil; el = iter.Next() {
		s, ok := el.(*Service)
		if !ok {
			panic("unexpected type found")
		}
		res = append(res, &Service{Service: *s.DeepCopy()})
	}
	txn.Commit()
	return res, nil
}
