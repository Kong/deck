package state

import (
	"fmt"

	memdb "github.com/hashicorp/go-memdb"
	"github.com/kong/deck/utils"
)

const (
	servicePackageTableName = "service-package"
)

var servicePackageTableSchema = &memdb.TableSchema{
	Name: servicePackageTableName,
	Indexes: map[string]*memdb.IndexSchema{
		"id": {
			Name:    "id",
			Unique:  true,
			Indexer: &memdb.StringFieldIndex{Field: "ID"},
		},
		"name": {
			Name:    "name",
			Unique:  true,
			Indexer: &memdb.StringFieldIndex{Field: "Name"},
		},
		all: allIndex,
	},
}

// ServicePackagesCollection stores and indexes Kong Services.
type ServicePackagesCollection collection

// Add adds a servicePackage to the collection.
// service.ID should not be nil else an error is thrown.
func (k *ServicePackagesCollection) Add(servicePackage ServicePackage) error {
	// TODO abstract this check in the go-memdb library itself
	if utils.Empty(servicePackage.ID) {
		return errIDRequired
	}
	txn := k.db.Txn(true)
	defer txn.Abort()

	var searchBy []string
	searchBy = append(searchBy, *servicePackage.ID)
	if !utils.Empty(servicePackage.Name) {
		searchBy = append(searchBy, *servicePackage.Name)
	}
	_, err := getServicePackage(txn, searchBy...)
	if err == nil {
		return fmt.Errorf("inserting servicePackage %v: %w", servicePackage.Console(), ErrAlreadyExists)
	} else if err != ErrNotFound {
		return err
	}

	err = txn.Insert(servicePackageTableName, &servicePackage)
	if err != nil {
		return err
	}
	txn.Commit()
	return nil
}

func getServicePackage(txn *memdb.Txn, IDs ...string) (*ServicePackage, error) {
	for _, id := range IDs {
		res, err := multiIndexLookupUsingTxn(txn, servicePackageTableName,
			[]string{"name", "id"}, id)
		if err == ErrNotFound {
			continue
		}
		if err != nil {
			return nil, err
		}
		servicePackage, ok := res.(*ServicePackage)
		if !ok {
			panic(unexpectedType)
		}
		return &ServicePackage{ServicePackage: *servicePackage.DeepCopy()}, nil
	}
	return nil, ErrNotFound
}

// Get gets a servicePackage by name or ID.
func (k *ServicePackagesCollection) Get(nameOrID string) (*ServicePackage, error) {
	if nameOrID == "" {
		return nil, errIDRequired
	}

	txn := k.db.Txn(false)
	defer txn.Abort()
	return getServicePackage(txn, nameOrID)
}

// Update udpates an existing service.
// It returns an error if the servicePackage is not already present.
func (k *ServicePackagesCollection) Update(servicePackage ServicePackage) error {
	// TODO abstract this check in the go-memdb library itself
	if utils.Empty(servicePackage.ID) {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteServicePackage(txn, *servicePackage.ID)
	if err != nil {
		return err
	}

	err = txn.Insert(servicePackageTableName, &servicePackage)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func deleteServicePackage(txn *memdb.Txn, nameOrID string) error {
	servicePackage, err := getServicePackage(txn, nameOrID)
	if err != nil {
		return err
	}

	err = txn.Delete(servicePackageTableName, servicePackage)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes a servicePackage by name or ID.
func (k *ServicePackagesCollection) Delete(nameOrID string) error {
	if nameOrID == "" {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteServicePackage(txn, nameOrID)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

// GetAll returns all the servicePackages.
func (k *ServicePackagesCollection) GetAll() ([]*ServicePackage, error) {
	txn := k.db.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(servicePackageTableName, all, true)
	if err != nil {
		return nil, err
	}

	var res []*ServicePackage
	for el := iter.Next(); el != nil; el = iter.Next() {
		s, ok := el.(*ServicePackage)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &ServicePackage{ServicePackage: *s.DeepCopy()})
	}
	txn.Commit()
	return res, nil
}
