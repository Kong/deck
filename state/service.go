package state

import (
	"fmt"

	"github.com/hashicorp/go-memdb"
	"github.com/pkg/errors"
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
		all: {
			Name: all,
			Indexer: &memdb.ConditionalIndex{
				Conditional: func(v interface{}) (bool, error) {
					return true, nil
				},
			},
		},
	},
}

// AddService adds a service to KongState
func (k *KongState) AddService(service Service) error {
	txn := k.memdb.Txn(true)
	defer txn.Commit()
	err := txn.Insert(serviceTableName, &service)
	if err != nil {
		return errors.Wrap(err, "insert failed")
	}
	return nil
}

// GetService gets a service by name or ID.
func (k *KongState) GetService(nameOrID string) (*Service, error) {
	res, err := k.multiIndexLookup(serviceTableName, []string{"name", id}, nameOrID)
	if err == ErrNotFound {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, errors.Wrap(err, "service lookup failed")
	}
	if res == nil {
		fmt.Println("res is nil")
	}
	service, ok := res.(*Service)
	if !ok {
		panic("unexpected type found")
	}
	return service, nil
}

func (k *KongState) UpdateService(service Service) error {
	txn := k.memdb.Txn(true)
	defer txn.Commit()
	err := txn.Insert(serviceTableName, &service)
	if err != nil {
		return errors.Wrap(err, "update failed")
	}
	return nil
}

// DeleteService deletes a service by name or ID.
func (k *KongState) DeleteService(nameOrID string) error {
	service, err := k.GetService(nameOrID)

	if err != nil {
		return errors.Wrap(err, "looking up service")
	}

	txn := k.memdb.Txn(true)
	defer txn.Commit()

	err = txn.Delete(serviceTableName, service)
	if err != nil {
		return errors.Wrap(err, "delete failed")
	}
	return nil
}

// GetAllServices gets a service by name or ID.
func (k *KongState) GetAllServices() ([]*Service, error) {
	txn := k.memdb.Txn(false)
	defer txn.Commit()

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
		res = append(res, s)
	}
	return res, nil
}
