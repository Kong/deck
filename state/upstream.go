package state

import (
	memdb "github.com/hashicorp/go-memdb"
	"github.com/pkg/errors"
)

const (
	upstreamTableName = "upstream"
)

var upstreamTableSchema = &memdb.TableSchema{
	Name: upstreamTableName,
	Indexes: map[string]*memdb.IndexSchema{
		id: {
			Name:    id,
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

// UpstreamsCollection stores and indexes Kong Upstreams.
type UpstreamsCollection struct {
	memdb *memdb.MemDB
}

// NewUpstreamsCollection instantiates a UpstreamsCollection.
func NewUpstreamsCollection() (*UpstreamsCollection, error) {
	var schema = &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			upstreamTableName: upstreamTableSchema,
		},
	}
	m, err := memdb.NewMemDB(schema)
	if err != nil {
		return nil, errors.Wrap(err, "creating new UpstreamCollection")
	}
	return &UpstreamsCollection{
		memdb: m,
	}, nil
}

// Add adds an upstream to the collection.
func (k *UpstreamsCollection) Add(upstream Upstream) error {
	txn := k.memdb.Txn(true)
	defer txn.Abort()
	err := txn.Insert(upstreamTableName, &upstream)
	if err != nil {
		return errors.Wrap(err, "insert failed")
	}
	txn.Commit()
	return nil
}

// Get gets an upstream by name or ID.
func (k *UpstreamsCollection) Get(nameOrID string) (*Upstream, error) {
	res, err := multiIndexLookup(k.memdb, upstreamTableName,
		[]string{"name", id}, nameOrID)
	if err == ErrNotFound {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, errors.Wrap(err, "upstream lookup failed")
	}
	if res == nil {
		return nil, ErrNotFound
	}
	u, ok := res.(*Upstream)
	if !ok {
		panic("unexpected type found")
	}
	return &Upstream{Upstream: *u.DeepCopy()}, nil
}

// Update udpates an exisitng upstream.
func (k *UpstreamsCollection) Update(upstream Upstream) error {
	// TODO check if entity is already present or not, throw error if present
	// TODO abstract this in the go-memdb library itself
	txn := k.memdb.Txn(true)
	defer txn.Abort()
	err := txn.Insert(upstreamTableName, &upstream)
	if err != nil {
		return errors.Wrap(err, "update failed")
	}
	txn.Commit()
	return nil
}

// Delete deletes an upstream by it's name or ID.
func (k *UpstreamsCollection) Delete(nameOrID string) error {
	upstream, err := k.Get(nameOrID)

	if err != nil {
		return errors.Wrap(err, "looking up upstream")
	}

	txn := k.memdb.Txn(true)
	defer txn.Abort()

	err = txn.Delete(upstreamTableName, upstream)
	if err != nil {
		return errors.Wrap(err, "delete failed")
	}
	txn.Commit()
	return nil
}

// GetAll gets all upstreams in the state.
func (k *UpstreamsCollection) GetAll() ([]*Upstream, error) {
	txn := k.memdb.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(upstreamTableName, all, true)
	if err != nil {
		return nil, errors.Wrapf(err, "upstream lookup failed")
	}

	var res []*Upstream
	for el := iter.Next(); el != nil; el = iter.Next() {
		u, ok := el.(*Upstream)
		if !ok {
			panic("unexpected type found")
		}
		res = append(res, &Upstream{Upstream: *u.DeepCopy()})
	}
	txn.Commit()
	return res, nil
}
