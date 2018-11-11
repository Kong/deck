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
	routeTableName   = "route"
	id               = "id"
	all              = "all"
)

var ErrNotFound = errors.New("entity not found")

var schema = &memdb.DBSchema{
	Tables: map[string]*memdb.TableSchema{
		serviceTableName: serviceTableSchema,
		routeTableName:   routeTableSchema,
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

	for _, indexName := range indices {
		res, err := txn.First(tableName, indexName, args...)
		if res == nil && err == nil {
			continue
		}
		if err != nil {
			return nil, err
		}
		if res != nil {
			return res, nil
		}
	}
	return nil, ErrNotFound
}
