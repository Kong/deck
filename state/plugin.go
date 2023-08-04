package state

import (
	"errors"
	"fmt"

	memdb "github.com/hashicorp/go-memdb"
	"github.com/kong/deck/state/indexers"
	"github.com/kong/deck/utils"
)

var errPluginNameRequired = fmt.Errorf("name of plugin required")

const (
	pluginTableName          = "plugin"
	pluginsByServiceID       = "pluginsByServiceID"
	pluginsByRouteID         = "pluginsByRouteID"
	pluginsByConsumerID      = "pluginsByConsumerID"
	pluginsByConsumerGroupID = "pluginsByConsumerGroupID"
)

var pluginTableSchema = &memdb.TableSchema{
	Name: pluginTableName,
	Indexes: map[string]*memdb.IndexSchema{
		"id": {
			Name:    "id",
			Unique:  true,
			Indexer: &memdb.StringFieldIndex{Field: "ID"},
		},
		"name": {
			Name:    "name",
			Indexer: &memdb.StringFieldIndex{Field: "Name"},
		},
		all: allIndex,
		// foreign
		pluginsByServiceID: {
			Name: pluginsByServiceID,
			Indexer: &indexers.SubFieldIndexer{
				Fields: []indexers.Field{
					{
						Struct: "Service",
						Sub:    "ID",
					},
				},
			},
			AllowMissing: true,
		},
		pluginsByRouteID: {
			Name: pluginsByRouteID,
			Indexer: &indexers.SubFieldIndexer{
				Fields: []indexers.Field{
					{
						Struct: "Route",
						Sub:    "ID",
					},
				},
			},
			AllowMissing: true,
		},
		pluginsByConsumerID: {
			Name: pluginsByConsumerID,
			Indexer: &indexers.SubFieldIndexer{
				Fields: []indexers.Field{
					{
						Struct: "Consumer",
						Sub:    "ID",
					},
				},
			},
			AllowMissing: true,
		},
		pluginsByConsumerGroupID: {
			Name: pluginsByConsumerGroupID,
			Indexer: &indexers.SubFieldIndexer{
				Fields: []indexers.Field{
					{
						Struct: "ConsumerGroup",
						Sub:    "ID",
					},
				},
			},
			AllowMissing: true,
		},
		// combined foreign fields
		// FIXME bug: collision if svc/route/consumer has the same ID
		// and same type of plugin is created. Consider the case when only
		// of the association is present
		"fields": {
			Name: "fields",
			Indexer: &indexers.SubFieldIndexer{
				Fields: []indexers.Field{
					{
						Struct: "Plugin",
						Sub:    "Name",
					},
					{
						Struct: "Service",
						Sub:    "ID",
					},
					{
						Struct: "Route",
						Sub:    "ID",
					},
					{
						Struct: "Consumer",
						Sub:    "ID",
					},
					{
						Struct: "ConsumerGroup",
						Sub:    "ID",
					},
				},
			},
		},
	},
}

// PluginsCollection stores and indexes Kong Services.
type PluginsCollection collection

// Add adds a plugin to PluginsCollection
func (k *PluginsCollection) Add(plugin Plugin) error {
	txn := k.db.Txn(true)
	defer txn.Abort()

	err := insertPlugin(txn, plugin)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func insertPlugin(txn *memdb.Txn, plugin Plugin) error {
	// TODO abstract this check in the go-memdb library itself
	if utils.Empty(plugin.ID) {
		return errIDRequired
	}
	if utils.Empty(plugin.Name) {
		return errPluginNameRequired
	}

	// err out if plugin with same ID is present
	_, err := getPluginByID(txn, *plugin.ID)
	if err == nil {
		return fmt.Errorf("inserting plugin %v: %w", plugin.Console(), ErrAlreadyExists)
	} else if !errors.Is(err, ErrNotFound) {
		return err
	}

	// err out if another plugin with exact same combination is present
	sID, rID, cID, cgID := "", "", "", ""
	if plugin.Service != nil && !utils.Empty(plugin.Service.ID) {
		sID = *plugin.Service.ID
	}
	if plugin.Route != nil && !utils.Empty(plugin.Route.ID) {
		rID = *plugin.Route.ID
	}
	if plugin.Consumer != nil && !utils.Empty(plugin.Consumer.ID) {
		cID = *plugin.Consumer.ID
	}
	if plugin.ConsumerGroup != nil && !utils.Empty(plugin.ConsumerGroup.ID) {
		cgID = *plugin.ConsumerGroup.ID
	}
	_, err = getPluginBy(txn, *plugin.Name, sID, rID, cID, cgID)
	if err == nil {
		return fmt.Errorf("inserting plugin %v: %w", plugin.Console(), ErrAlreadyExists)
	} else if !errors.Is(err, ErrNotFound) {
		return err
	}

	// all good
	err = txn.Insert(pluginTableName, &plugin)
	if err != nil {
		return err
	}
	return nil
}

func getPluginByID(txn *memdb.Txn, id string) (*Plugin, error) {
	res, err := multiIndexLookupUsingTxn(txn, pluginTableName,
		[]string{"id"}, id)
	if err != nil {
		return nil, err
	}

	plugin, ok := res.(*Plugin)
	if !ok {
		panic(unexpectedType)
	}
	return &Plugin{Plugin: *plugin.DeepCopy()}, nil
}

// Get gets a plugin by id.
func (k *PluginsCollection) Get(id string) (*Plugin, error) {
	if id == "" {
		return nil, errIDRequired
	}

	txn := k.db.Txn(false)
	defer txn.Abort()

	plugin, err := getPluginByID(txn, id)
	if err != nil {
		return nil, err
	}
	return plugin, nil
}

// GetAllByName returns all plugins of a specific type
// (key-auth, ratelimiting, etc).
func (k *PluginsCollection) GetAllByName(name string) ([]*Plugin, error) {
	return k.getAllPluginsBy("name", name)
}

func getPluginBy(txn *memdb.Txn, name, svcID, routeID, consumerID, consumerGroupID string) (
	*Plugin, error,
) {
	if name == "" {
		return nil, errPluginNameRequired
	}

	res, err := txn.First(pluginTableName, "fields",
		name, svcID, routeID, consumerID, consumerGroupID)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, ErrNotFound
	}
	p, ok := res.(*Plugin)
	if !ok {
		panic(unexpectedType)
	}
	return &Plugin{Plugin: *p.DeepCopy()}, nil
}

// GetByProp returns a plugin which matches all the properties passed in
// the arguments. If serviceID, routeID, consumerID and consumerGroupID
// are empty strings, then a global plugin is searched.
// Otherwise, a plugin with name and the supplied foreign references is
// searched.
// name is required.
func (k *PluginsCollection) GetByProp(
	name, serviceID, routeID, consumerID, consumerGroupID string,
) (*Plugin, error) {
	txn := k.db.Txn(false)
	defer txn.Abort()

	return getPluginBy(txn, name, serviceID, routeID, consumerID, consumerGroupID)
}

func (k *PluginsCollection) getAllPluginsBy(index, identifier string) (
	[]*Plugin, error,
) {
	if identifier == "" {
		return nil, errIDRequired
	}

	txn := k.db.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(pluginTableName, index, identifier)
	if err != nil {
		return nil, err
	}
	var res []*Plugin
	for el := iter.Next(); el != nil; el = iter.Next() {
		p, ok := el.(*Plugin)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &Plugin{Plugin: *p.DeepCopy()})
	}
	return res, nil
}

// GetAllByServiceID returns all plugins referencing a service
// by its id.
func (k *PluginsCollection) GetAllByServiceID(id string) ([]*Plugin,
	error,
) {
	return k.getAllPluginsBy(pluginsByServiceID, id)
}

// GetAllByRouteID returns all plugins referencing a route
// by its id.
func (k *PluginsCollection) GetAllByRouteID(id string) ([]*Plugin,
	error,
) {
	return k.getAllPluginsBy(pluginsByRouteID, id)
}

// GetAllByConsumerID returns all plugins referencing a consumer
// by its id.
func (k *PluginsCollection) GetAllByConsumerID(id string) ([]*Plugin,
	error,
) {
	return k.getAllPluginsBy(pluginsByConsumerID, id)
}

// GetAllByConsumerGroupID returns all plugins referencing a consumer-group
// by its id.
func (k *PluginsCollection) GetAllByConsumerGroupID(id string) ([]*Plugin,
	error,
) {
	return k.getAllPluginsBy(pluginsByConsumerGroupID, id)
}

// Update updates a plugin
func (k *PluginsCollection) Update(plugin Plugin) error {
	// TODO abstract this check in the go-memdb library itself
	if utils.Empty(plugin.ID) {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deletePlugin(txn, *plugin.ID)
	if err != nil {
		return err
	}

	err = insertPlugin(txn, plugin)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func deletePlugin(txn *memdb.Txn, id string) error {
	plugin, err := getPluginByID(txn, id)
	if err != nil {
		return err
	}
	return txn.Delete(pluginTableName, plugin)
}

// Delete deletes a plugin by ID.
func (k *PluginsCollection) Delete(id string) error {
	if id == "" {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deletePlugin(txn, id)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

// GetAll gets a plugin by name or ID.
func (k *PluginsCollection) GetAll() ([]*Plugin, error) {
	txn := k.db.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(pluginTableName, all, true)
	if err != nil {
		return nil, err
	}

	var res []*Plugin
	for el := iter.Next(); el != nil; el = iter.Next() {
		p, ok := el.(*Plugin)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &Plugin{Plugin: *p.DeepCopy()})
	}
	return res, nil
}
