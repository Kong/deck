package state

import (
	"fmt"

	memdb "github.com/hashicorp/go-memdb"
	"github.com/kong/deck/state/indexers"
	"github.com/kong/deck/utils"
)

var (
	errGroupRequired    = fmt.Errorf("name of ACL group required")
	errConsumerRequired = fmt.Errorf("consumer required")
)

const (
	aclGroupTableName     = "aclGroup"
	aclGroupsByConsumerID = "aclGroupsByConsumerID"
)

var aclGroupTableSchema = &memdb.TableSchema{
	Name: aclGroupTableName,
	Indexes: map[string]*memdb.IndexSchema{
		"id": {
			Name:    "id",
			Unique:  true,
			Indexer: &memdb.StringFieldIndex{Field: "ID"},
		},
		"group": {
			Name:    "group",
			Indexer: &memdb.StringFieldIndex{Field: "Group"},
		},
		all: allIndex,
		// foreign
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
	},
}

// ACLGroupsCollection stores and indexes acl-group credentials.
type ACLGroupsCollection collection

// Add adds aclGroup to ACLGroupsCollection
func (k *ACLGroupsCollection) Add(aclGroup ACLGroup) error {
	// TODO abstract this check in the go-memdb library itself
	if utils.Empty(aclGroup.ID) {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := insertACLGroup(txn, aclGroup)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func insertACLGroup(txn *memdb.Txn, aclGroup ACLGroup) error {
	if utils.Empty(aclGroup.ID) {
		return errIDRequired
	}

	// err out if group with same ID is present
	_, err := getACLGroupByID(txn, *aclGroup.ID)
	if err == nil {
		return fmt.Errorf("inserting acl-group %v: %w", aclGroup.Console(), ErrAlreadyExists)
	} else if err != ErrNotFound {
		return err
	}

	// check if the same combination is present
	if utils.Empty(aclGroup.Group) {
		return errGroupRequired
	}
	if aclGroup.Consumer == nil || utils.Empty(aclGroup.Consumer.ID) {
		return errConsumerRequired
	}
	_, err = getACLGroup(txn, *aclGroup.Consumer.ID, *aclGroup.Group)
	if err == nil {
		return fmt.Errorf("inserting acl-group %v: %w", aclGroup.Console(), ErrAlreadyExists)
	} else if err != ErrNotFound {
		return err
	}

	// all good
	err = txn.Insert(aclGroupTableName, &aclGroup)
	if err != nil {
		return err
	}
	return nil
}

func getACLGroupByID(txn *memdb.Txn, id string) (*ACLGroup, error) {
	res, err := multiIndexLookupUsingTxn(txn, aclGroupTableName,
		[]string{"id"}, id)
	if err != nil {
		return nil, err
	}
	aclGroup, ok := res.(*ACLGroup)
	if !ok {
		panic(unexpectedType)
	}
	return &ACLGroup{ACLGroup: *aclGroup.DeepCopy()}, nil
}

// GetByID gets an acl-group with id.
func (k *ACLGroupsCollection) GetByID(id string) (*ACLGroup, error) {
	if id == "" {
		return nil, errIDRequired
	}
	txn := k.db.Txn(false)
	defer txn.Abort()
	return getACLGroupByID(txn, id)
}

func getACLGroup(txn *memdb.Txn, consumerID, groupOrID string) (*ACLGroup, error) {
	groups, err := getAllACLGroupsByConsumerID(txn, consumerID)
	if err != nil {
		return nil, err
	}
	for _, group := range groups {
		if groupOrID == *group.ID || groupOrID == *group.Group {
			return &ACLGroup{ACLGroup: *group.DeepCopy()}, nil
		}
	}
	return nil, ErrNotFound
}

func getAllACLGroupsByConsumerID(txn *memdb.Txn, consumerID string) ([]*ACLGroup, error) {
	iter, err := txn.Get(aclGroupTableName, aclGroupsByConsumerID, consumerID)
	if err != nil {
		return nil, err
	}
	var res []*ACLGroup
	for el := iter.Next(); el != nil; el = iter.Next() {
		r, ok := el.(*ACLGroup)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &ACLGroup{ACLGroup: *r.DeepCopy()})
	}
	return res, nil
}

// Get gets a acl-group for a consumer by group or ID.
func (k *ACLGroupsCollection) Get(consumerID,
	groupOrID string) (*ACLGroup, error) {
	if groupOrID == "" {
		return nil, errIDRequired
	}

	txn := k.db.Txn(false)
	defer txn.Abort()
	return getACLGroup(txn, consumerID, groupOrID)
}

// GetAllByConsumerID returns all acl-group credentials
// belong to a Consumer with id.
func (k *ACLGroupsCollection) GetAllByConsumerID(id string) ([]*ACLGroup,
	error) {
	if id == "" {
		return nil, errIDRequired
	}

	txn := k.db.Txn(false)
	defer txn.Abort()

	return getAllACLGroupsByConsumerID(txn, id)
}

// Update updates an existing acl-group credential.
func (k *ACLGroupsCollection) Update(aclGroup ACLGroup) error {
	// TODO abstract this check in the go-memdb library itself
	if utils.Empty(aclGroup.ID) {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteACLGroup(txn, *aclGroup.ID)
	if err != nil {
		return err
	}

	err = insertACLGroup(txn, aclGroup)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func deleteACLGroup(txn *memdb.Txn, id string) error {
	group, err := getACLGroupByID(txn, id)
	if err != nil {
		return err
	}

	err = txn.Delete(aclGroupTableName, group)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes an acl-group by id.
func (k *ACLGroupsCollection) Delete(id string) error {
	if id == "" {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteACLGroup(txn, id)
	if err != nil {
		return err
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
		return nil, err
	}

	var res []*ACLGroup
	for el := iter.Next(); el != nil; el = iter.Next() {
		r, ok := el.(*ACLGroup)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &ACLGroup{ACLGroup: *r.DeepCopy()})
	}
	txn.Commit()
	return res, nil
}
