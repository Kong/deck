package state

import (
	"errors"
	"fmt"

	"github.com/hashicorp/go-memdb"

	"github.com/kong/deck/state/indexers"
	"github.com/kong/deck/utils"
)

const (
	serviceVersionTableName    = "service-version"
	versionsByServicePackageID = "serviceVersionsByServicePackageID"
)

var errInvalidPackage = fmt.Errorf("servicePackage.ID is required in ServiceVersion")

var serviceVersionTableSchema = &memdb.TableSchema{
	Name: serviceVersionTableName,
	Indexes: map[string]*memdb.IndexSchema{
		"id": {
			Name:    "id",
			Unique:  true,
			Indexer: &memdb.StringFieldIndex{Field: "ID"},
		},
		all: allIndex,
		// foreign
		versionsByServicePackageID: {
			Name: versionsByServicePackageID,
			Indexer: &indexers.SubFieldIndexer{
				Fields: []indexers.Field{
					{
						Struct: "ServicePackage",
						Sub:    "ID",
					},
				},
			},
		},
	},
}

func validatePackage(version ServiceVersion) error {
	if version.ServicePackage == nil ||
		utils.Empty(version.ServicePackage.ID) {
		return errInvalidPackage
	}
	return nil
}

// ServiceVersionsCollection stores and indexes Service Versions.
type ServiceVersionsCollection collection

// Add adds a serviceVersion into ServiceVersionsCollection
// serviceVersion.ID should not be nil else an error is thrown.
func (k *ServiceVersionsCollection) Add(serviceVersion ServiceVersion) error {
	// TODO abstract this check in the go-memdb library itself
	if utils.Empty(serviceVersion.ID) {
		return errIDRequired
	}

	if err := validatePackage(serviceVersion); err != nil {
		return err
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	var searchBy []string
	searchBy = append(searchBy, *serviceVersion.ID)
	if !utils.Empty(serviceVersion.Version) {
		searchBy = append(searchBy, *serviceVersion.Version)
	}
	_, err := getServiceVersion(txn, *serviceVersion.ServicePackage.ID, searchBy...)
	if err == nil {
		return fmt.Errorf("inserting serviceVersion %v: %w", serviceVersion.Console(), ErrAlreadyExists)
	} else if !errors.Is(err, ErrNotFound) {
		return err
	}

	err = txn.Insert(serviceVersionTableName, &serviceVersion)
	if err != nil {
		return err
	}
	txn.Commit()
	return nil
}

func getServiceVersion(txn *memdb.Txn, packageID string, IDs ...string) (*ServiceVersion, error) {
	if packageID == "" {
		return nil, fmt.Errorf("packageID is required")
	}
	versions, err := getAllByPackageID(txn, packageID)
	if err != nil {
		return nil, err
	}

	for _, id := range IDs {
		for _, version := range versions {
			if id == *version.ID || id == *version.Version {
				return &ServiceVersion{ServiceVersion: *version.DeepCopy()}, nil
			}
		}
	}
	return nil, ErrNotFound
}

// Get gets a Service Version by name or ID.
func (k *ServiceVersionsCollection) Get(packageID, nameOrID string) (*ServiceVersion, error) {
	if nameOrID == "" {
		return nil, errIDRequired
	}

	txn := k.db.Txn(false)
	defer txn.Abort()
	serviceVersion, err := getServiceVersion(txn, packageID, nameOrID)
	if err != nil {
		return nil, err
	}
	return serviceVersion, nil
}

// Update updates a Service Version.
func (k *ServiceVersionsCollection) Update(serviceVersion ServiceVersion) error {
	// TODO abstract this check in the go-memdb library itself
	if utils.Empty(serviceVersion.ID) {
		return errIDRequired
	}
	if err := validatePackage(serviceVersion); err != nil {
		return err
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteServiceVersion(txn, *serviceVersion.ServicePackage.ID, *serviceVersion.ID)
	if err != nil {
		return err
	}

	err = txn.Insert(serviceVersionTableName, &serviceVersion)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func deleteServiceVersion(txn *memdb.Txn, packageID, nameOrID string) error {
	serviceVersion, err := getServiceVersion(txn, packageID, nameOrID)
	if err != nil {
		return err
	}

	err = txn.Delete(serviceVersionTableName, serviceVersion)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes a serviceVersion by name or ID.
func (k *ServiceVersionsCollection) Delete(packageID, nameOrID string) error {
	if nameOrID == "" {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteServiceVersion(txn, packageID, nameOrID)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

// GetAll gets all serviceVersios.
func (k *ServiceVersionsCollection) GetAll() ([]*ServiceVersion, error) {
	txn := k.db.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(serviceVersionTableName, all, true)
	if err != nil {
		return nil, err
	}

	var res []*ServiceVersion
	for el := iter.Next(); el != nil; el = iter.Next() {
		s, ok := el.(*ServiceVersion)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &ServiceVersion{ServiceVersion: *s.DeepCopy()})
	}
	txn.Commit()
	return res, nil
}

func getAllByPackageID(txn *memdb.Txn, packageID string) ([]*ServiceVersion, error) {
	iter, err := txn.Get(serviceVersionTableName, versionsByServicePackageID, packageID)
	if err != nil {
		return nil, err
	}

	var versions []*ServiceVersion
	for el := iter.Next(); el != nil; el = iter.Next() {
		v, ok := el.(*ServiceVersion)
		if !ok {
			panic(unexpectedType)
		}
		versions = append(versions, &ServiceVersion{ServiceVersion: *v.DeepCopy()})
	}
	return versions, nil
}

// GetAllByServicePackageID returns all serviceVersions for a servicePackage id.
func (k *ServiceVersionsCollection) GetAllByServicePackageID(id string) ([]*ServiceVersion,
	error,
) {
	txn := k.db.Txn(false)
	return getAllByPackageID(txn, id)
}
