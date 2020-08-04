package state

import (
	memdb "github.com/hashicorp/go-memdb"
	"github.com/kong/deck/state/indexers"
	"github.com/kong/deck/utils"
	"github.com/pkg/errors"
)

const (
	targetTableName     = "target"
	targetsByUpstreamID = "targetsByUpstreamID"
)

var errInvalidUpstream = errors.New("upstream.ID is required in target")

var targetTableSchema = &memdb.TableSchema{
	Name: targetTableName,
	Indexes: map[string]*memdb.IndexSchema{
		"id": {
			Name:    "id",
			Unique:  true,
			Indexer: &memdb.StringFieldIndex{Field: "ID"},
		},
		"target": {
			Name: "target",
			Indexer: &indexers.SubFieldIndexer{
				Fields: []indexers.Field{
					{
						Struct: "Target",
						Sub:    "Target",
					},
				},
			},
		},
		all: allIndex,
		// foreign
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
	},
}

func validateUpstream(target *Target) error {
	if target.Upstream == nil ||
		utils.Empty(target.Upstream.ID) {
		return errInvalidUpstream
	}
	return nil
}

// TargetsCollection stores and indexes Kong Upstreams.
type TargetsCollection collection

// Add adds a target to TargetsCollection.
// target should have an ID, Target and it's upstream's ID is set.
func (k *TargetsCollection) Add(target Target) error {
	// TODO abstract this check in the go-memdb library itself
	if utils.Empty(target.ID) {
		return errIDRequired
	}

	if err := validateUpstream(&target); err != nil {
		return err
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	var searchBy []string
	searchBy = append(searchBy, *target.ID)
	if !utils.Empty(target.Target.Target) {
		searchBy = append(searchBy, *target.Target.Target)
	}
	_, err := getTarget(txn, *target.Upstream.ID, searchBy...)
	if err == nil {
		return ErrAlreadyExists
	} else if err != ErrNotFound {
		return err
	}

	err = txn.Insert(targetTableName, &target)
	if err != nil {
		return err
	}
	txn.Commit()
	return nil
}

func getTarget(txn *memdb.Txn, upstreamID string, IDs ...string) (*Target, error) {
	targets, err := getAllByUpstreamID(txn, upstreamID)
	if err != nil {
		return nil, err
	}

	for _, id := range IDs {
		for _, target := range targets {
			if id == *target.ID || id == *target.Target.Target {
				return &Target{Target: *target.DeepCopy()}, nil
			}
		}
	}
	return nil, ErrNotFound
}

func getAllByUpstreamID(txn *memdb.Txn, upstreamID string) ([]*Target, error) {
	iter, err := txn.Get(targetTableName, targetsByUpstreamID, upstreamID)
	if err != nil {
		return nil, err
	}

	var targets []*Target
	for el := iter.Next(); el != nil; el = iter.Next() {
		t, ok := el.(*Target)
		if !ok {
			panic(unexpectedType)
		}
		targets = append(targets, &Target{Target: *t.DeepCopy()})
	}
	return targets, nil
}

// Get returns a specific target for upstream with upstreamID.
func (k *TargetsCollection) Get(upstreamID,
	targetOrID string) (*Target, error) {

	txn := k.db.Txn(false)
	defer txn.Abort()

	return getTarget(txn, upstreamID, targetOrID)
}

// Update updates a target
func (k *TargetsCollection) Update(target Target) error {
	// TODO abstract this check in the go-memdb library itself
	if utils.Empty(target.ID) {
		return errIDRequired
	}

	if err := validateUpstream(&target); err != nil {
		return err
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	// This doesn't follow the usual getTarget() because
	// the target.Upstream.ID can be different from the one in the DB.
	res, err := multiIndexLookupUsingTxn(txn, targetTableName,
		[]string{"id"}, *target.ID)
	if err != nil {
		return err
	}

	t, ok := res.(*Target)
	if !ok {
		panic(unexpectedType)
	}
	err = txn.Delete(targetTableName, *t)
	if err != nil {
		return err
	}

	err = txn.Insert(targetTableName, &target)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func deleteTarget(txn *memdb.Txn, upstreamID, targetOrID string) error {
	target, err := getTarget(txn, upstreamID, targetOrID)
	if err != nil {
		return err
	}

	err = txn.Delete(targetTableName, target)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes a target by its ID.
func (k *TargetsCollection) Delete(upstreamID, targetOrID string) error {
	if targetOrID == "" {
		return errIDRequired
	}

	if upstreamID == "" {
		return errInvalidUpstream
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteTarget(txn, upstreamID, targetOrID)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

// GetAll gets a target by Target or ID.
func (k *TargetsCollection) GetAll() ([]*Target, error) {
	txn := k.db.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(targetTableName, all, true)
	if err != nil {
		return nil, err
	}

	var res []*Target
	for el := iter.Next(); el != nil; el = iter.Next() {
		t, ok := el.(*Target)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &Target{Target: *t.DeepCopy()})
	}
	txn.Commit()
	return res, nil
}

// GetAllByUpstreamID returns all targets referencing a Upstream
// by its ID.
func (k *TargetsCollection) GetAllByUpstreamID(id string) ([]*Target,
	error) {
	txn := k.db.Txn(false)
	return getAllByUpstreamID(txn, id)
}
