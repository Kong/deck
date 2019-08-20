package state

import (
	"github.com/hashicorp/go-memdb"
	"github.com/hbagdi/deck/state/indexers"
	"github.com/pkg/errors"
)

const (
	targetTableName       = "target"
	targetsByUpstreamName = "targetsByUpstreamName"
	targetsByUpstreamID   = "targetsByUpstreamID"
)

var errInvalidUpstream = errors.New("upstream with ID and name is required in target")

var targetTableSchema = &memdb.TableSchema{
	Name: targetTableName,
	Indexes: map[string]*memdb.IndexSchema{
		id: {
			Name:    id,
			Unique:  true,
			Indexer: &memdb.StringFieldIndex{Field: "ID"},
		},
		targetsByUpstreamName: {
			Name: targetsByUpstreamName,
			Indexer: &indexers.SubFieldIndexer{
				Fields: []indexers.Field{
					{
						Struct: "Upstream",
						Sub:    "Name",
					},
				},
			},
		},
		targetsByUpstreamID: {
			Name: targetsByUpstreamID,
			Indexer: &indexers.SubFieldIndexer{
				Fields: []indexers.Field{
					{
						Struct: "Upstream",
						Sub:    "ID",
					},
				},
			},
		},
		"upstreamNameTarget": {
			Name:   "upstreamNameTarget",
			Unique: true,
			Indexer: &indexers.SubFieldIndexer{
				Fields: []indexers.Field{
					{
						Struct: "Upstream",
						Sub:    "Name",
					}, {
						Struct: "Target",
						Sub:    "Target",
					},
				},
			},
		},
		all: allIndex,
	},
}

// TargetsCollection stores and indexes Kong Upstreams.
type TargetsCollection struct {
	memdb *memdb.MemDB
}

// NewTargetsCollection instantiates a TargetsCollection.
func NewTargetsCollection() (*TargetsCollection, error) {
	var schema = &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			targetTableName: targetTableSchema,
		},
	}
	m, err := memdb.NewMemDB(schema)
	if err != nil {
		return nil, errors.Wrap(err, "creating new TargetCollection")
	}
	return &TargetsCollection{
		memdb: m,
	}, nil
}

// Add adds a target to TargetsCollection.
func (k *TargetsCollection) Add(target Target) error {
	if err := k.validateTarget(&target); err != nil {
		return err
	}

	txn := k.memdb.Txn(true)
	defer txn.Abort()
	err := txn.Insert(targetTableName, &target)
	if err != nil {
		return errors.Wrap(err, "insert failed")
	}
	txn.Commit()
	return nil
}

// Get gets a target by ID.
func (k *TargetsCollection) Get(ID string) (*Target, error) {
	res, err := multiIndexLookup(k.memdb, targetTableName, []string{id}, ID)
	if err == ErrNotFound {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, errors.Wrap(err, "target lookup failed")
	}
	if res == nil {
		return nil, ErrNotFound
	}
	t, ok := res.(*Target)
	if !ok {
		panic("unexpected type found")
	}
	return &Target{Target: *t.DeepCopy()}, nil
}

// GetByUpstreamNameAndTarget get a target by upstreamName and target
func (k *TargetsCollection) GetByUpstreamNameAndTarget(upstreamName, target string) (*Target, error) {
	txn := k.memdb.Txn(false)
	defer txn.Abort()

	res, err := txn.First(targetTableName, "upstreamNameTarget", upstreamName, target)
	if err != nil {
		return nil, errors.Wrap(err, "target lookup failed")
	}
	if res == nil {
		return nil, ErrNotFound
	}

	t, ok := res.(*Target)
	if !ok {
		panic("unexpected type found")
	}
	return &Target{Target: *t.DeepCopy()}, nil
}

// GetAllByUpstreamName returns all targets referencing a Upstream
// by its name.
func (k *TargetsCollection) GetAllByUpstreamName(
	name string) ([]*Target, error) {
	txn := k.memdb.Txn(false)
	iter, err := txn.Get(targetTableName, targetsByUpstreamName, name)
	if err != nil {
		return nil, err
	}
	var res []*Target
	for el := iter.Next(); el != nil; el = iter.Next() {
		t, ok := el.(*Target)
		if !ok {
			panic("unexpected type found")
		}
		res = append(res, &Target{Target: *t.DeepCopy()})
	}
	return res, nil
}

// GetAllByUpstreamID returns all targets referencing a Upstream
// by its ID.
func (k *TargetsCollection) GetAllByUpstreamID(id string) ([]*Target,
	error) {
	txn := k.memdb.Txn(false)
	iter, err := txn.Get(targetTableName, targetsByUpstreamID, id)
	if err != nil {
		return nil, err
	}
	var res []*Target
	for el := iter.Next(); el != nil; el = iter.Next() {
		t, ok := el.(*Target)
		if !ok {
			panic("unexpected type found")
		}
		res = append(res, &Target{Target: *t.DeepCopy()})
	}
	return res, nil
}

// Update updates a target
func (k *TargetsCollection) Update(target Target) error {
	if err := k.validateTarget(&target); err != nil {
		return err
	}

	txn := k.memdb.Txn(true)
	defer txn.Abort()
	err := txn.Insert(targetTableName, &target)
	if err != nil {
		return errors.Wrap(err, "update failed")
	}
	txn.Commit()
	return nil
}

// Delete deletes a target by it's ID.
func (k *TargetsCollection) Delete(ID string) error {
	target, err := k.Get(ID)

	if err != nil {
		return errors.Wrap(err, "looking up target")
	}

	txn := k.memdb.Txn(true)
	defer txn.Abort()

	err = txn.Delete(targetTableName, target)
	if err != nil {
		return errors.Wrap(err, "delete failed")
	}
	txn.Commit()
	return nil
}

// GetAll gets a target by Target or ID.
func (k *TargetsCollection) GetAll() ([]*Target, error) {
	txn := k.memdb.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(targetTableName, all, true)
	if err != nil {
		return nil, errors.Wrapf(err, "target lookup failed")
	}

	var res []*Target
	for el := iter.Next(); el != nil; el = iter.Next() {
		t, ok := el.(*Target)
		if !ok {
			panic("unexpected type found")
		}
		res = append(res, &Target{Target: *t.DeepCopy()})
	}
	txn.Commit()
	return res, nil
}

func (k *TargetsCollection) validateTarget(target *Target) error {
	if target.Upstream == nil ||
		target.Upstream.ID == nil || *target.Upstream.ID == "" ||
		target.Upstream.Name == nil || *target.Upstream.Name == "" {
		return errInvalidUpstream
	}
	return nil
}
