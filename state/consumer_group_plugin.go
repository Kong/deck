package state

import (
	"errors"
	"fmt"

	memdb "github.com/hashicorp/go-memdb"

	"github.com/kong/deck/state/indexers"
	"github.com/kong/deck/utils"
)

const (
	consumerGroupPluginTableName = "consumerGroupPlugin"
	pluginByGroupID              = "pluginByGroupID"
)

var consumerGroupPluginTableSchema = &memdb.TableSchema{
	Name: consumerGroupPluginTableName,
	Indexes: map[string]*memdb.IndexSchema{
		"id": {
			Name:   "id",
			Unique: true,
			Indexer: &indexers.SubFieldIndexer{
				Fields: []indexers.Field{
					{
						Struct: "ConsumerGroupPlugin",
						Sub:    "ID",
					},
					{
						Struct: "ConsumerGroup",
						Sub:    "ID",
					},
				},
			},
		},
		"name": {
			Name:   "name",
			Unique: true,
			Indexer: &indexers.SubFieldIndexer{
				Fields: []indexers.Field{
					{
						Struct: "ConsumerGroupPlugin",
						Sub:    "Name",
					},
				},
			},
		},
		all: allIndex,
		// foreign
		pluginByGroupID: {
			Name: pluginByGroupID,
			Indexer: &indexers.SubFieldIndexer{
				Fields: []indexers.Field{
					{
						Struct: "ConsumerGroup",
						Sub:    "ID",
					},
				},
			},
		},
	},
}

func validatePluginGroup(plugin *ConsumerGroupPlugin) error {
	if plugin.ConsumerGroup == nil ||
		utils.Empty(plugin.ConsumerGroup.ID) {
		return errInvalidConsumerGroup
	}
	return nil
}

// ConsumerGroupPluginsCollection stores and indexes Kong consumerGroupPlugins.
type ConsumerGroupPluginsCollection collection

// Add adds a consumerGroupPlugin to the collection.
func (k *ConsumerGroupPluginsCollection) Add(plugin ConsumerGroupPlugin) error {
	var nameOrID string
	if plugin.ConsumerGroupPlugin.ID != nil {
		nameOrID = *plugin.ConsumerGroupPlugin.ID
	} else {
		nameOrID = *plugin.ConsumerGroupPlugin.Name
	}

	if err := validatePluginGroup(&plugin); err != nil {
		return err
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	var searchBy []string
	searchBy = append(searchBy, nameOrID, *plugin.ConsumerGroup.ID)
	_, err := getConsumerGroupPlugin(txn, *plugin.ConsumerGroup.ID, searchBy...)
	if err == nil {
		return fmt.Errorf("inserting consumerGroupPlugin %v: %w", plugin.Console(), ErrAlreadyExists)
	} else if !errors.Is(err, ErrNotFound) {
		return err
	}

	err = txn.Insert(consumerGroupPluginTableName, &plugin)
	if err != nil {
		return err
	}
	txn.Commit()
	return nil
}

func getAllPluginsByConsumerGroupID(txn *memdb.Txn, consumerGroupID string) ([]*ConsumerGroupPlugin, error) {
	iter, err := txn.Get(consumerGroupPluginTableName, pluginByGroupID, consumerGroupID)
	if err != nil {
		return nil, err
	}

	var plugins []*ConsumerGroupPlugin
	for el := iter.Next(); el != nil; el = iter.Next() {
		t, ok := el.(*ConsumerGroupPlugin)
		if !ok {
			panic(unexpectedType)
		}
		plugins = append(plugins, &ConsumerGroupPlugin{ConsumerGroupPlugin: *t.DeepCopy()})
	}
	return plugins, nil
}

func getConsumerGroupPlugin(txn *memdb.Txn, consumerGroupID string, IDs ...string) (*ConsumerGroupPlugin, error) {
	plugins, err := getAllPluginsByConsumerGroupID(txn, consumerGroupID)
	if err != nil {
		return nil, err
	}

	for _, id := range IDs {
		for _, plugin := range plugins {
			if id == *plugin.ID || id == *plugin.Name {
				return &ConsumerGroupPlugin{ConsumerGroupPlugin: *plugin.DeepCopy()}, nil
			}
		}
	}
	return nil, ErrNotFound
}

// Get gets a consumerGroupPlugin.
func (k *ConsumerGroupPluginsCollection) Get(
	nameOrID, consumerGroupID string,
) (*ConsumerGroupPlugin, error) {
	txn := k.db.Txn(false)
	defer txn.Abort()

	return getConsumerGroupPlugin(txn, consumerGroupID, nameOrID)
}

// Update udpates an existing consumerGroupPlugin.
func (k *ConsumerGroupPluginsCollection) Update(plugin ConsumerGroupPlugin) error {
	if utils.Empty(plugin.ConsumerGroupPlugin.ID) {
		return errIDRequired
	}

	if err := validatePluginGroup(&plugin); err != nil {
		return err
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	res, err := txn.First(consumerGroupPluginTableName, "id",
		*plugin.ConsumerGroupPlugin.ID, *plugin.ConsumerGroup.ID)
	if err != nil {
		return err
	}

	t, ok := res.(*ConsumerGroupPlugin)
	if !ok {
		panic(unexpectedType)
	}
	err = txn.Delete(consumerGroupPluginTableName, *t)
	if err != nil {
		return err
	}

	err = txn.Insert(consumerGroupPluginTableName, &plugin)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func deleteConsumerGroupPlugin(txn *memdb.Txn, nameOrID, consumerGroupID string) error {
	consumer, err := getConsumerGroupPlugin(txn, consumerGroupID, nameOrID)
	if err != nil {
		return err
	}
	err = txn.Delete(consumerGroupPluginTableName, consumer)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes a consumerGroupPlugin by its username or ID.
func (k *ConsumerGroupPluginsCollection) Delete(nameOrID, consumerGroupID string) error {
	if nameOrID == "" {
		return errIDRequired
	}

	if consumerGroupID == "" {
		return errInvalidConsumerGroup
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteConsumerGroupPlugin(txn, nameOrID, consumerGroupID)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

// GetAll gets all consumerGroupPlugins in the state.
func (k *ConsumerGroupPluginsCollection) GetAll() ([]*ConsumerGroupPlugin, error) {
	txn := k.db.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(consumerGroupPluginTableName, all, true)
	if err != nil {
		return nil, err
	}

	var res []*ConsumerGroupPlugin
	for el := iter.Next(); el != nil; el = iter.Next() {
		u, ok := el.(*ConsumerGroupPlugin)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &ConsumerGroupPlugin{ConsumerGroupPlugin: *u.DeepCopy()})
	}
	txn.Commit()
	return res, nil
}
