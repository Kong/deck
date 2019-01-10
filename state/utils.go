package state

import (
	memdb "github.com/hashicorp/go-memdb"
	"github.com/pkg/errors"
)

const (
	id  = "id"
	all = "all"
)

// ErrNotFound is an error type that is
// returned when an entity is not found in the state.
var ErrNotFound = errors.New("entity not found")

var allIndex = &memdb.IndexSchema{
	Name: all,
	Indexer: &memdb.ConditionalIndex{
		Conditional: func(v interface{}) (bool, error) {
			return true, nil
		},
	},
}

// multiIndexLookup can be used to search for an entity
// based on search on multiple indexes with same key.
func multiIndexLookup(memdb *memdb.MemDB, tableName string,
	indices []string,
	args ...interface{}) (interface{}, error) {

	txn := memdb.Txn(false)
	defer txn.Abort()

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
