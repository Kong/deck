package state

import (
	"fmt"

	memdb "github.com/hashicorp/go-memdb"
	"github.com/kong/deck/utils"
)

const (
	consumerGroupTableName = "consumerGroup"
)

var consumerGroupTableSchema = &memdb.TableSchema{
	Name: consumerGroupTableName,
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

// consumerGroupsCollection stores and indexes Kong consumerGroups.
type ConsumerGroupsCollection collection

// Add adds an consumerGroup to the collection.
// consumerGroup.ID should not be nil else an error is thrown.
func (k *ConsumerGroupsCollection) Add(consumerGroup ConsumerGroup) error {
	if utils.Empty(consumerGroup.ID) {
		return errIDRequired
	}
	txn := k.db.Txn(true)
	defer txn.Abort()

	var searchBy []string
	searchBy = append(searchBy, *consumerGroup.ID)
	if !utils.Empty(consumerGroup.Name) {
		searchBy = append(searchBy, *consumerGroup.Name)
	}
	_, err := getconsumerGroup(txn, searchBy...)
	if err == nil {
		return fmt.Errorf("inserting consumerGroup %v: %w", consumerGroup.Console(), ErrAlreadyExists)
	} else if err != ErrNotFound {
		return err
	}

	err = txn.Insert(consumerGroupTableName, &consumerGroup)
	if err != nil {
		return err
	}
	txn.Commit()
	return nil
}

func getconsumerGroup(txn *memdb.Txn, IDs ...string) (*ConsumerGroup, error) {
	for _, id := range IDs {
		res, err := multiIndexLookupUsingTxn(txn, consumerGroupTableName,
			[]string{"name", "id"}, id)
		if err == ErrNotFound {
			continue
		}
		if err != nil {
			return nil, err
		}

		consumerGroup, ok := res.(*ConsumerGroup)
		if !ok {
			panic(unexpectedType)
		}
		return &ConsumerGroup{ConsumerGroup: *consumerGroup.DeepCopy()}, nil
	}
	return nil, ErrNotFound
}

// Get gets an consumerGroup by name or ID.
func (k *ConsumerGroupsCollection) Get(nameOrID string) (*ConsumerGroup, error) {
	if nameOrID == "" {
		return nil, errIDRequired
	}

	txn := k.db.Txn(false)
	defer txn.Abort()
	consumerGroup, err := getconsumerGroup(txn, nameOrID)
	if err != nil {
		if err == ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return consumerGroup, nil
}

// Update udpates an existing consumerGroup.
func (k *ConsumerGroupsCollection) Update(consumerGroup ConsumerGroup) error {
	if utils.Empty(consumerGroup.ID) {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteconsumerGroup(txn, *consumerGroup.ID)
	if err != nil {
		return err
	}

	err = txn.Insert(consumerGroupTableName, &consumerGroup)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func deleteconsumerGroup(txn *memdb.Txn, nameOrID string) error {
	consumerGroup, err := getconsumerGroup(txn, nameOrID)
	if err != nil {
		return err
	}

	err = txn.Delete(consumerGroupTableName, consumerGroup)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes an consumerGroup by its name or ID.
func (k *ConsumerGroupsCollection) Delete(nameOrID string) error {
	if nameOrID == "" {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteconsumerGroup(txn, nameOrID)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

// GetAll gets all consumerGroups in the state.
func (k *ConsumerGroupsCollection) GetAll() ([]*ConsumerGroup, error) {
	txn := k.db.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(consumerGroupTableName, all, true)
	if err != nil {
		return nil, err
	}

	var res []*ConsumerGroup
	for el := iter.Next(); el != nil; el = iter.Next() {
		u, ok := el.(*ConsumerGroup)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &ConsumerGroup{ConsumerGroup: *u.DeepCopy()})
	}
	txn.Commit()
	return res, nil
}
