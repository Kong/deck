package state

import (
	"fmt"

	memdb "github.com/hashicorp/go-memdb"
	"github.com/kong/deck/state/indexers"
	"github.com/kong/deck/utils"
)

const (
	consumerGroupTableName = "consumerGroup"
)

var consumerGroupTableSchema = &memdb.TableSchema{
	Name: consumerGroupTableName,
	Indexes: map[string]*memdb.IndexSchema{
		"id": {
			Name:   "id",
			Unique: true,
			Indexer: &indexers.SubFieldIndexer{
				Fields: []indexers.Field{
					{
						Struct: "ConsumerGroup",
						Sub:    "ID",
					},
				},
			},
		},
		"name": {
			Name:   "name",
			Unique: true,
			Indexer: &indexers.SubFieldIndexer{
				Fields: []indexers.Field{
					{
						Struct: "ConsumerGroup",
						Sub:    "Name",
					},
				},
			},
		},
		all: allIndex,
	},
}

// consumerGroupsCollection stores and indexes Kong consumerGroups.
type ConsumerGroupsCollection collection

// Add adds an consumerGroup to the collection.
// consumerGroup.ID should not be nil else an error is thrown.
func (k *ConsumerGroupsCollection) Add(consumerGroup ConsumerGroupObject) error {
	// TODO abstract this check in the go-memdb library itself
	if utils.Empty(consumerGroup.ConsumerGroup.ID) {
		return errIDRequired
	}
	txn := k.db.Txn(true)
	defer txn.Abort()

	var searchBy []string
	searchBy = append(searchBy, *consumerGroup.ConsumerGroup.ID)
	if !utils.Empty(consumerGroup.ConsumerGroup.Name) {
		searchBy = append(searchBy, *consumerGroup.ConsumerGroup.Name)
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

func getconsumerGroup(txn *memdb.Txn, IDs ...string) (*ConsumerGroupObject, error) {
	for _, id := range IDs {
		res, err := multiIndexLookupUsingTxn(txn, consumerGroupTableName,
			[]string{"name", "id"}, id)
		if err == ErrNotFound {
			continue
		}
		if err != nil {
			return nil, err
		}

		consumerGroup, ok := res.(*ConsumerGroupObject)
		if !ok {
			panic(unexpectedType)
		}
		return &ConsumerGroupObject{
			ConsumerGroupObject: *consumerGroup.DeepCopy()}, nil
	}
	return nil, ErrNotFound
}

// Get gets an consumerGroup by name or ID.
func (k *ConsumerGroupsCollection) Get(nameOrID string) (*ConsumerGroupObject, error) {
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
func (k *ConsumerGroupsCollection) Update(consumerGroup ConsumerGroupObject) error {
	// TODO abstract this in the go-memdb library itself
	if utils.Empty(consumerGroup.ConsumerGroup.ID) {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteconsumerGroup(txn, *consumerGroup.ConsumerGroup.ID)
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

// Delete deletes an consumerGroup by it's name or ID.
func (k *ConsumerGroupsCollection) Delete(consumerGroup ConsumerGroupObject) error {
	if utils.Empty(consumerGroup.ConsumerGroup.ID) && utils.Empty(consumerGroup.ConsumerGroup.Name) {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteconsumerGroup(txn, *consumerGroup.ConsumerGroup.ID)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

// GetAll gets all consumerGroups in the state.
func (k *ConsumerGroupsCollection) GetAll() ([]*ConsumerGroupObject, error) {
	txn := k.db.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(consumerGroupTableName, all, true)
	if err != nil {
		return nil, err
	}

	var res []*ConsumerGroupObject
	for el := iter.Next(); el != nil; el = iter.Next() {
		u, ok := el.(*ConsumerGroupObject)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &ConsumerGroupObject{
			ConsumerGroupObject: *u.DeepCopy()})
	}
	txn.Commit()
	return res, nil
}
