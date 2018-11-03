package state

import (
	memdb "github.com/hashicorp/go-memdb"
	"github.com/pkg/errors"
)

// KongState is an in-memory database representation
// of Kong's configuration.
type KongState struct {
	memdb *memdb.MemDB
}

const (
	serviceTableName = "service"
	id               = "id"
	all              = "all"
)

var schema = &memdb.DBSchema{
	Tables: map[string]*memdb.TableSchema{
		serviceTableName: serviceTableSchema,
	},
}

// NewKongState creates a new in-memory KongState.
func NewKongState() (*KongState, error) {
	m, err := memdb.NewMemDB(schema)
	if err != nil {
		return nil, errors.Wrap(err, "creating new KongState")
	}
	return &KongState{
		memdb: m,
	}, nil
}

// multiIndexLookup can be used to search for an entity
// based on search on multiple indexes with same key.
func (k *KongState) multiIndexLookup(tableName string,
	indices []string,
	args ...interface{}) (interface{}, error) {

	txn := k.memdb.Txn(false)
	defer txn.Commit()

	var (
		res interface{}
		err error
	)
	for _, indexName := range indices {
		res, err = txn.First(tableName, indexName, args...)
		if err == nil {
			return res, nil
		}
		if err == memdb.ErrNotFound {
			continue
		} else {
			return nil, errors.Wrap(err, "lookup failed")
		}
	}
	return nil, errors.New("not found")
}
