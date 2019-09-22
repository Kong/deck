package state

import (
	memdb "github.com/hashicorp/go-memdb"
	"github.com/hbagdi/deck/state/indexers"
	"github.com/pkg/errors"
)

const (
	aclGroupTableName           = "aclGroup"
	aclGroupsByConsumerUsername = "aclGroupsByConsumerUsername"
	aclGroupsByConsumerID       = "aclGroupsByConsumerID"
)

var aclGroupTableSchema = &memdb.TableSchema{
	Name: aclGroupTableName,
	Indexes: map[string]*memdb.IndexSchema{
		id: {
			Name:    id,
			Unique:  true,
			Indexer: &memdb.StringFieldIndex{Field: "ID"},
		},
		aclGroupsByConsumerUsername: {
			Name: aclGroupsByConsumerUsername,
			Indexer: &indexers.SubFieldIndexer{
				Fields: []indexers.Field{
					{
						Struct: "Consumer",
						Sub:    "Username",
					},
				},
			},
		},
		aclGroupsByConsumerID: {
			Name: aclGroupsByConsumerID,
			Indexer: &indexers.SubFieldIndexer{
				Fields: []indexers.Field{
					{
						Struct: "Consumer",
						Sub:    "ID",
					},
				},
			},
		},
		"group": {
			Name:    "group",
			Indexer: &memdb.StringFieldIndex{Field: "Group"},
		},
		all: allIndex,
	},
}

// ACLGroupsCollection stores and indexes acl-group credentials.
type ACLGroupsCollection collection

// Add adds aclGroup to ACLGroupsCollection
func (k *ACLGroupsCollection) Add(aclGroup ACLGroup) error {
	txn := k.db.Txn(true)
	defer txn.Abort()
	err := txn.Insert(aclGroupTableName, &aclGroup)
	if err != nil {
		return errors.Wrap(err, "insert failed")
	}
	txn.Commit()
	return nil
}

// GetByID gets an acl-group with id.
func (k *ACLGroupsCollection) GetByID(id string) (*ACLGroup, error) {
	res, err := multiIndexLookup(k.db, aclGroupTableName,
		[]string{"id"}, id)
	if err == ErrNotFound {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, errors.Wrap(err, "aclGroup lookup failed")
	}
	if res == nil {
		return nil, ErrNotFound
	}
	group, ok := res.(*ACLGroup)
	if !ok {
		panic("unexpected type found")
	}

	return &ACLGroup{ACLGroup: *group.DeepCopy()}, nil
}

// Get gets a acl-group for a consumer by group or ID.
func (k *ACLGroupsCollection) Get(consumerUsernameOrID,
	groupOrID string) (*ACLGroup, error) {

	txn := k.db.Txn(false)
	defer txn.Abort()

	indices := []string{aclGroupsByConsumerUsername, aclGroupsByConsumerID}
	var groups []*ACLGroup
	// load all groups
	for _, indexName := range indices {
		iter, err := txn.Get(aclGroupTableName, indexName, consumerUsernameOrID)
		if err != nil {
			return nil, errors.Wrapf(err, "aclGroup lookup failed")
		}
		for el := iter.Next(); el != nil; el = iter.Next() {
			r, ok := el.(*ACLGroup)
			if !ok {
				panic("unexpected type found")
			}
			groups = append(groups, &ACLGroup{ACLGroup: *r.DeepCopy()})
		}
	}
	txn.Commit()
	// linear search
	for _, group := range groups {
		if groupOrID == *group.ID || groupOrID == *group.Group {
			return &ACLGroup{ACLGroup: *group.DeepCopy()}, nil
		}
	}
	return nil, ErrNotFound
}

// GetAllByConsumerUsername returns all acl-group credentials
// belong to a Consumer with username.
func (k *ACLGroupsCollection) GetAllByConsumerUsername(username string) ([]*ACLGroup,
	error) {
	txn := k.db.Txn(false)
	iter, err := txn.Get(aclGroupTableName, aclGroupsByConsumerUsername, username)
	if err != nil {
		return nil, err
	}
	var res []*ACLGroup
	for el := iter.Next(); el != nil; el = iter.Next() {
		r, ok := el.(*ACLGroup)
		if !ok {
			panic("unexpected type found")
		}
		res = append(res, &ACLGroup{ACLGroup: *r.DeepCopy()})
	}
	return res, nil
}

// GetAllByConsumerID returns all acl-group credentials
// belong to a Consumer with id.
func (k *ACLGroupsCollection) GetAllByConsumerID(id string) ([]*ACLGroup,
	error) {
	txn := k.db.Txn(false)
	iter, err := txn.Get(aclGroupTableName, aclGroupsByConsumerID, id)
	if err != nil {
		return nil, err
	}
	var res []*ACLGroup
	for el := iter.Next(); el != nil; el = iter.Next() {
		r, ok := el.(*ACLGroup)
		if !ok {
			panic("unexpected type found")
		}
		res = append(res, &ACLGroup{ACLGroup: *r.DeepCopy()})
	}
	return res, nil
}

// Update updates an existing acl-group credential.
func (k *ACLGroupsCollection) Update(aclGroup ACLGroup) error {
	txn := k.db.Txn(true)
	defer txn.Abort()
	err := txn.Insert(aclGroupTableName, &aclGroup)
	if err != nil {
		return errors.Wrap(err, "update failed")
	}
	txn.Commit()
	return nil
}

// DeleteByID deletes an acl-group by ID.
func (k *ACLGroupsCollection) DeleteByID(ID string) error {
	aclGroup, err := k.GetByID(ID)

	if err != nil {
		return errors.Wrap(err, "looking up aclGroup")
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err = txn.Delete(aclGroupTableName, aclGroup)
	if err != nil {
		return errors.Wrap(err, "delete failed")
	}
	txn.Commit()
	return nil
}

// GetAll gets all acl-groups.
func (k *ACLGroupsCollection) GetAll() ([]*ACLGroup, error) {
	txn := k.db.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(aclGroupTableName, all, true)
	if err != nil {
		return nil, errors.Wrapf(err, "aclGroup lookup failed")
	}

	var res []*ACLGroup
	for el := iter.Next(); el != nil; el = iter.Next() {
		r, ok := el.(*ACLGroup)
		if !ok {
			panic("unexpected type found")
		}
		res = append(res, &ACLGroup{ACLGroup: *r.DeepCopy()})
	}
	txn.Commit()
	return res, nil
}
