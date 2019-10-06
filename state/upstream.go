package state

import (
	memdb "github.com/hashicorp/go-memdb"
	"github.com/hbagdi/deck/utils"
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
type UpstreamsCollection collection

// Add adds an upstream to the collection.
// upstream.ID should not be nil else an error is thrown.
func (k *UpstreamsCollection) Add(upstream Upstream) error {
	// TODO abstract this check in the go-memdb library itself
	if utils.Empty(upstream.ID) {
		return errIDRequired
	}
	txn := k.db.Txn(true)
	defer txn.Abort()

	var searchBy []string
	searchBy = append(searchBy, *upstream.ID)
	if !utils.Empty(upstream.Name) {
		searchBy = append(searchBy, *upstream.Name)
	}
	_, err := getUpstream(txn, searchBy...)
	if err == nil {
		return ErrAlreadyExists
	} else if err != ErrNotFound {
		return err
	}

	err = txn.Insert(upstreamTableName, &upstream)
	if err != nil {
		return err
	}
	txn.Commit()
	return nil
}

func getUpstream(txn *memdb.Txn, IDs ...string) (*Upstream, error) {
	for _, id := range IDs {
		res, err := multiIndexLookupUsingTxn(txn, upstreamTableName,
			[]string{"name", "id"}, id)
		if err == ErrNotFound {
			continue
		}
		if err != nil {
			return nil, err
		}

		upstream, ok := res.(*Upstream)
		if !ok {
			panic(unexpectedType)
		}
		return &Upstream{Upstream: *upstream.DeepCopy()}, nil
	}
	return nil, ErrNotFound
}

// Get gets an upstream by name or ID.
func (k *UpstreamsCollection) Get(nameOrID string) (*Upstream, error) {
	if nameOrID == "" {
		return nil, errIDRequired
	}

	txn := k.db.Txn(false)
	defer txn.Abort()
	upstream, err := getUpstream(txn, nameOrID)
	if err != nil {
		if err == ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return upstream, nil
}

// Update udpates an existing upstream.
func (k *UpstreamsCollection) Update(upstream Upstream) error {
	// TODO abstract this in the go-memdb library itself
	if utils.Empty(upstream.ID) {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteUpstream(txn, *upstream.ID)
	if err != nil {
		return err
	}

	err = txn.Insert(upstreamTableName, &upstream)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func deleteUpstream(txn *memdb.Txn, nameOrID string) error {
	upstream, err := getUpstream(txn, nameOrID)
	if err != nil {
		return err
	}

	err = txn.Delete(upstreamTableName, upstream)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes an upstream by it's name or ID.
func (k *UpstreamsCollection) Delete(nameOrID string) error {
	if nameOrID == "" {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteUpstream(txn, nameOrID)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

// GetAll gets all upstreams in the state.
func (k *UpstreamsCollection) GetAll() ([]*Upstream, error) {
	txn := k.db.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(upstreamTableName, all, true)
	if err != nil {
		return nil, err
	}

	var res []*Upstream
	for el := iter.Next(); el != nil; el = iter.Next() {
		u, ok := el.(*Upstream)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &Upstream{Upstream: *u.DeepCopy()})
	}
	txn.Commit()
	return res, nil
}
