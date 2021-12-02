package state

import (
	"fmt"

	memdb "github.com/hashicorp/go-memdb"
	"github.com/kong/deck/utils"
)

const (
	developerTableName = "developer"
)

var developerTableSchema = &memdb.TableSchema{
	Name: developerTableName,
	Indexes: map[string]*memdb.IndexSchema{
		"id": {
			Name:    "id",
			Unique:  true,
			Indexer: &memdb.StringFieldIndex{Field: "ID"},
		},
		"Email": {
			Name:         "Email",
			Unique:       true,
			Indexer:      &memdb.StringFieldIndex{Field: "Email"},
			AllowMissing: true,
		},
		all: allIndex,
	},
}

// DevelopersCollection stores and indexes Kong Developers.
type DevelopersCollection collection

// Add adds a developer to the collection
// An error is thrown if developer.ID is empty.
func (k *DevelopersCollection) Add(developer Developer) error {
	if utils.Empty(developer.ID) {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	var searchBy []string
	searchBy = append(searchBy, *developer.ID)
	if !utils.Empty(developer.Email) {
		searchBy = append(searchBy, *developer.Email)
	}
	_, err := getDeveloper(txn, searchBy...)
	if err == nil {
		return fmt.Errorf("inserting developer %v: %w", developer.Console(), ErrAlreadyExists)
	} else if err != ErrNotFound {
		return err
	}

	err = txn.Insert(developerTableName, &developer)
	if err != nil {
		return err
	}
	txn.Commit()
	return nil
}

func getDeveloper(txn *memdb.Txn, IDs ...string) (*Developer, error) {
	for _, id := range IDs {
		res, err := multiIndexLookupUsingTxn(txn, developerTableName,
			[]string{"Email", "id"}, id)
		if err == ErrNotFound {
			continue
		}
		if err != nil {
			return nil, err
		}
		developer, ok := res.(*Developer)
		if !ok {
			panic(unexpectedType)
		}
		return &Developer{Developer: *developer.DeepCopy()}, nil
	}
	return nil, ErrNotFound
}

// Get gets a developer by email or ID.
func (k *DevelopersCollection) Get(userEmailOrID string) (*Developer, error) {
	if userEmailOrID == "" {
		return nil, errIDRequired
	}

	txn := k.db.Txn(false)
	defer txn.Abort()
	return getDeveloper(txn, userEmailOrID)
}

// Update udpates an existing developer.
// It returns an error if the developer is not already present.
func (k *DevelopersCollection) Update(developer Developer) error {
	// TODO abstract this in the go-memdb library itself
	if utils.Empty(developer.ID) {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteDeveloper(txn, *developer.ID)
	if err != nil {
		return err
	}

	err = txn.Insert(developerTableName, &developer)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func deleteDeveloper(txn *memdb.Txn, userEmailOrID string) error {
	developer, err := getDeveloper(txn, userEmailOrID)
	if err != nil {
		return err
	}

	err = txn.Delete(developerTableName, developer)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes a developer by email or ID.
func (k *DevelopersCollection) Delete(userEmailOrID string) error {
	if userEmailOrID == "" {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteDeveloper(txn, userEmailOrID)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

// GetAll gets a developer by email or ID.
func (k *DevelopersCollection) GetAll() ([]*Developer, error) {
	txn := k.db.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(developerTableName, all, true)
	if err != nil {
		return nil, err
	}

	var res []*Developer
	for el := iter.Next(); el != nil; el = iter.Next() {
		s, ok := el.(*Developer)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &Developer{Developer: *s.DeepCopy()})
	}
	txn.Commit()
	return res, nil
}
