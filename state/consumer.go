package state

import (
	"fmt"

	memdb "github.com/hashicorp/go-memdb"
	"github.com/kong/deck/utils"
)

const (
	consumerTableName = "consumer"
)

var consumerTableSchema = &memdb.TableSchema{
	Name: consumerTableName,
	Indexes: map[string]*memdb.IndexSchema{
		"id": {
			Name:    "id",
			Unique:  true,
			Indexer: &memdb.StringFieldIndex{Field: "ID"},
		},
		"Username": {
			Name:         "Username",
			Unique:       true,
			Indexer:      &memdb.StringFieldIndex{Field: "Username"},
			AllowMissing: true,
		},
		all: allIndex,
	},
}

// ConsumersCollection stores and indexes Kong Consumers.
type ConsumersCollection collection

// Add adds a consumer to the collection
// An error is thrown if consumer.ID is empty.
func (k *ConsumersCollection) Add(consumer Consumer) error {
	if utils.Empty(consumer.ID) {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	var searchBy []string
	searchBy = append(searchBy, *consumer.ID)
	if !utils.Empty(consumer.Username) {
		searchBy = append(searchBy, *consumer.Username)
	}
	_, err := getConsumer(txn, searchBy...)
	if err == nil {
		return fmt.Errorf("inserting consumer %v: %w", consumer.Console(), ErrAlreadyExists)
	} else if err != ErrNotFound {
		return err
	}

	err = txn.Insert(consumerTableName, &consumer)
	if err != nil {
		return err
	}
	txn.Commit()
	return nil
}

func getConsumer(txn *memdb.Txn, IDs ...string) (*Consumer, error) {
	for _, id := range IDs {
		res, err := multiIndexLookupUsingTxn(txn, consumerTableName,
			[]string{"Username", "id"}, id)
		if err == ErrNotFound {
			continue
		}
		if err != nil {
			return nil, err
		}
		consumer, ok := res.(*Consumer)
		if !ok {
			panic(unexpectedType)
		}
		return &Consumer{Consumer: *consumer.DeepCopy()}, nil
	}
	return nil, ErrNotFound
}

// Get gets a consumer by name or ID.
func (k *ConsumersCollection) Get(userNameOrID string) (*Consumer, error) {
	if userNameOrID == "" {
		return nil, errIDRequired
	}

	txn := k.db.Txn(false)
	defer txn.Abort()
	return getConsumer(txn, userNameOrID)
}

// Update udpates an existing consumer.
// It returns an error if the consumer is not already present.
func (k *ConsumersCollection) Update(consumer Consumer) error {
	// TODO abstract this in the go-memdb library itself
	if utils.Empty(consumer.ID) {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteConsumer(txn, *consumer.ID)
	if err != nil {
		return err
	}

	err = txn.Insert(consumerTableName, &consumer)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func deleteConsumer(txn *memdb.Txn, userNameOrID string) error {
	consumer, err := getConsumer(txn, userNameOrID)
	if err != nil {
		return err
	}

	err = txn.Delete(consumerTableName, consumer)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes a consumer by name or ID.
func (k *ConsumersCollection) Delete(userNameOrID string) error {
	if userNameOrID == "" {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteConsumer(txn, userNameOrID)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

// GetAll gets a consumer by name or ID.
func (k *ConsumersCollection) GetAll() ([]*Consumer, error) {
	txn := k.db.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(consumerTableName, all, true)
	if err != nil {
		return nil, err
	}

	var res []*Consumer
	for el := iter.Next(); el != nil; el = iter.Next() {
		s, ok := el.(*Consumer)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &Consumer{Consumer: *s.DeepCopy()})
	}
	txn.Commit()
	return res, nil
}
