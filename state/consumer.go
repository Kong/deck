package state

import (
	memdb "github.com/hashicorp/go-memdb"
	"github.com/pkg/errors"
)

const (
	consumerTableName = "consumer"
)

var consumerTableSchema = &memdb.TableSchema{
	Name: consumerTableName,
	Indexes: map[string]*memdb.IndexSchema{
		id: {
			Name:   id,
			Unique: true,
			// EnforceUnique: true,
			Indexer: &memdb.StringFieldIndex{Field: "ID"},
		},
		"Username": {
			Name:   "Username",
			Unique: true,
			// EnforceUnique: true,
			Indexer: &memdb.StringFieldIndex{Field: "Username"},
		},
		all: allIndex,
	},
}

// ConsumersCollection stores and indexes Kong Consumers.
type ConsumersCollection struct {
	memdb *memdb.MemDB
}

// NewConsumersCollection instantiates a ConsumersCollection.
func NewConsumersCollection() (*ConsumersCollection, error) {
	var schema = &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			consumerTableName: consumerTableSchema,
		},
	}
	m, err := memdb.NewMemDB(schema)
	if err != nil {
		return nil, errors.Wrap(err, "creating new ConsumerCollection")
	}
	return &ConsumersCollection{
		memdb: m,
	}, nil
}

// Add adds a consumer to the collection
func (k *ConsumersCollection) Add(consumer Consumer) error {
	txn := k.memdb.Txn(true)
	defer txn.Abort()
	err := txn.Insert(consumerTableName, &consumer)
	if err != nil {
		return errors.Wrap(err, "insert failed")
	}
	txn.Commit()
	return nil
}

// Get gets a consumer by name or ID.
func (k *ConsumersCollection) Get(userNameOrID string) (*Consumer, error) {
	res, err := multiIndexLookup(k.memdb, consumerTableName,
		[]string{"Username", id}, userNameOrID)
	if err == ErrNotFound {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, errors.Wrap(err, "consumer lookup failed")
	}
	if res == nil {
		return nil, ErrNotFound
	}
	consumer, ok := res.(*Consumer)
	if !ok {
		panic("unexpected type found")
	}

	return &Consumer{Consumer: *consumer.DeepCopy()}, nil
}

// Update udpates an exisitng consumer.
// It returns an error if the consumer is not already present.
func (k *ConsumersCollection) Update(consumer Consumer) error {
	// TODO check if entity is already present or not, throw error if present
	// TODO abstract this in the go-memdb library itself
	txn := k.memdb.Txn(true)
	defer txn.Abort()
	err := txn.Insert(consumerTableName, &consumer)
	if err != nil {
		return errors.Wrap(err, "update failed")
	}
	txn.Commit()
	return nil
}

// Delete deletes a consumer by name or ID.
func (k *ConsumersCollection) Delete(usernameOrID string) error {
	consumer, err := k.Get(usernameOrID)

	if err != nil {
		return errors.Wrap(err, "looking up consumer")
	}

	txn := k.memdb.Txn(true)
	defer txn.Abort()

	err = txn.Delete(consumerTableName, consumer)
	if err != nil {
		return errors.Wrap(err, "delete failed")
	}
	txn.Commit()
	return nil
}

// GetAll gets a consumer by name or ID.
func (k *ConsumersCollection) GetAll() ([]*Consumer, error) {
	txn := k.memdb.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(consumerTableName, all, true)
	if err != nil {
		return nil, errors.Wrapf(err, "consumer lookup failed")
	}

	var res []*Consumer
	for el := iter.Next(); el != nil; el = iter.Next() {
		s, ok := el.(*Consumer)
		if !ok {
			panic("unexpected type found")
		}
		res = append(res, &Consumer{Consumer: *s.DeepCopy()})
	}
	txn.Commit()
	return res, nil
}
