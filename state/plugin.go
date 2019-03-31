package state

import (
	memdb "github.com/hashicorp/go-memdb"
	"github.com/hbagdi/deck/state/indexers"
	"github.com/pkg/errors"
)

const (
	pluginTableName           = "plugin"
	pluginsByServiceName      = "pluginsByServiceName"
	pluginsByServiceID        = "pluginsByServiceID"
	pluginsByRouteName        = "pluginsByRouteName"
	pluginsByRouteID          = "pluginsByRouteID"
	pluginsByConsumerUsername = "pluginsByConsumerUsername"
	pluginsByConsumerID       = "pluginsByConsumerID"
)

var pluginTableSchema = &memdb.TableSchema{
	Name: pluginTableName,
	Indexes: map[string]*memdb.IndexSchema{
		id: {
			Name:    id,
			Unique:  true,
			Indexer: &memdb.StringFieldIndex{Field: "ID"},
		},
		pluginsByServiceName: {
			Name: pluginsByServiceName,
			Indexer: &indexers.SubFieldIndexer{
				Fields: []indexers.Field{
					{
						Struct: "Service",
						Sub:    "Name",
					},
				},
			},
			AllowMissing: true,
		},
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
		pluginsByRouteName: {
			Name: pluginsByRouteName,
			Indexer: &indexers.SubFieldIndexer{
				Fields: []indexers.Field{
					{
						Struct: "Route",
						Sub:    "Name",
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
		pluginsByConsumerUsername: {
			Name: pluginsByConsumerUsername,
			Indexer: &indexers.SubFieldIndexer{
				Fields: []indexers.Field{
					{
						Struct: "Consumer",
						Sub:    "Username",
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
		"name": {
			Name:    "name",
			Indexer: &memdb.StringFieldIndex{Field: "Name"},
		},
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
						Sub:    "Name",
					},
					{
						Struct: "Route",
						Sub:    "Name",
					},
					{
						Struct: "Consumer",
						Sub:    "Username",
					},
				},
			},
			AllowMissing: true,
		},
		"foreignfields": {
			Name: "foreignfields",
			Indexer: &indexers.SubFieldIndexer{
				Fields: []indexers.Field{
					{
						Struct: "Service",
						Sub:    "Name",
					},
					{
						Struct: "Route",
						Sub:    "Name",
					},
				},
			},
			AllowMissing: true,
		},
		all: allIndex,
	},
}

// PluginsCollection stores and indexes Kong Services.
type PluginsCollection struct {
	memdb *memdb.MemDB
}

// NewPluginsCollection instantiates a PluginsCollection.
func NewPluginsCollection() (*PluginsCollection, error) {
	var schema = &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			pluginTableName: pluginTableSchema,
		},
	}
	m, err := memdb.NewMemDB(schema)
	if err != nil {
		return nil, errors.Wrap(err, "creating new PluginCollection")
	}
	return &PluginsCollection{
		memdb: m,
	}, nil
}

// Add adds a plugin to PluginsCollection
func (k *PluginsCollection) Add(plugin Plugin) error {
	txn := k.memdb.Txn(true)
	defer txn.Abort()
	err := txn.Insert(pluginTableName, &plugin)
	if err != nil {
		return errors.Wrap(err, "insert failed")
	}
	txn.Commit()
	return nil
}

// Get gets a plugin by name or ID.
func (k *PluginsCollection) Get(ID string) (*Plugin, error) {
	res, err := multiIndexLookup(k.memdb, pluginTableName,
		[]string{"name", id}, ID)
	if err == ErrNotFound {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, errors.Wrap(err, "plugin lookup failed")
	}
	if res == nil {
		return nil, ErrNotFound
	}
	p, ok := res.(*Plugin)
	if !ok {
		panic("unexpected type found")
	}
	return &Plugin{Plugin: *p.DeepCopy()}, nil
}

// GetAllByName returns all plugins of a specific type
// (key-auth, ratelimiting, etc).
func (k *PluginsCollection) GetAllByName(name string) ([]*Plugin,
	error) {
	txn := k.memdb.Txn(false)
	iter, err := txn.Get(pluginTableName, "name", name)
	if err != nil {
		return nil, err
	}
	var res []*Plugin
	for el := iter.Next(); el != nil; el = iter.Next() {
		s, ok := el.(*Plugin)
		if !ok {
			panic("unexpected type found")
		}
		res = append(res, s)
	}
	return res, nil
}

// GetByProp returns a plugin which matches all the properties passed in
// the arguments. If serviceName and routeName are empty strings, then
// a global plugin is searched.
// If serviceName is empty, a plugin for the route with routeName is searched.
// If routeName is empty, a plugin for the route with serviceName is searched.
func (k *PluginsCollection) GetByProp(name, serviceName,
	routeName string, consumerUsername string) (*Plugin, error) {
	txn := k.memdb.Txn(false)
	defer txn.Commit()
	res, err := txn.First(pluginTableName, "fields",
		name, serviceName, routeName, consumerUsername)
	if err == ErrNotFound {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "plugin lookup failed")
	}
	if res == nil {
		return nil, ErrNotFound
	}
	p, ok := res.(*Plugin)
	if !ok {
		panic("unexpected type found")
	}
	return &Plugin{Plugin: *p.DeepCopy()}, nil
}

// GetAllByServiceID returns all plugins referencing a service
// by its id.
func (k *PluginsCollection) GetAllByServiceID(id string) ([]*Plugin,
	error) {
	txn := k.memdb.Txn(false)
	iter, err := txn.Get(pluginTableName, pluginsByServiceID, id)
	if err != nil {
		return nil, err
	}
	var res []*Plugin
	for el := iter.Next(); el != nil; el = iter.Next() {
		p, ok := el.(*Plugin)
		if !ok {
			panic("unexpected type found")
		}
		res = append(res, &Plugin{Plugin: *p.DeepCopy()})
	}
	return res, nil
}

// GetAllByRouteID returns all plugins referencing a service
// by its id.
func (k *PluginsCollection) GetAllByRouteID(id string) ([]*Plugin,
	error) {
	txn := k.memdb.Txn(false)
	iter, err := txn.Get(pluginTableName, pluginsByRouteID, id)
	if err != nil {
		return nil, err
	}
	var res []*Plugin
	for el := iter.Next(); el != nil; el = iter.Next() {
		p, ok := el.(*Plugin)
		if !ok {
			panic("unexpected type found")
		}
		res = append(res, &Plugin{Plugin: *p.DeepCopy()})
	}
	return res, nil
}

// GetAllByConsumerID returns all plugins referencing a consumer
// by its id.
func (k *PluginsCollection) GetAllByConsumerID(id string) ([]*Plugin,
	error) {
	txn := k.memdb.Txn(false)
	iter, err := txn.Get(pluginTableName, pluginsByConsumerID, id)
	if err != nil {
		return nil, err
	}
	var res []*Plugin
	for el := iter.Next(); el != nil; el = iter.Next() {
		p, ok := el.(*Plugin)
		if !ok {
			panic("unexpected type found")
		}
		res = append(res, &Plugin{Plugin: *p.DeepCopy()})
	}
	return res, nil
}

// Update updates a plugin
func (k *PluginsCollection) Update(plugin Plugin) error {
	txn := k.memdb.Txn(true)
	defer txn.Abort()
	err := txn.Insert(pluginTableName, &plugin)
	if err != nil {
		return errors.Wrap(err, "update failed")
	}
	txn.Commit()
	return nil
}

// Delete deletes a plugin by name or ID.
func (k *PluginsCollection) Delete(nameOrID string) error {
	plugin, err := k.Get(nameOrID)

	if err != nil {
		return errors.Wrap(err, "looking up plugin")
	}

	txn := k.memdb.Txn(true)
	defer txn.Abort()

	err = txn.Delete(pluginTableName, plugin)
	if err != nil {
		return errors.Wrap(err, "delete failed")
	}
	txn.Commit()
	return nil
}

// GetAll gets a plugin by name or ID.
func (k *PluginsCollection) GetAll() ([]*Plugin, error) {
	txn := k.memdb.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(pluginTableName, all, true)
	if err != nil {
		return nil, errors.Wrapf(err, "plugin lookup failed")
	}

	var res []*Plugin
	for el := iter.Next(); el != nil; el = iter.Next() {
		p, ok := el.(*Plugin)
		if !ok {
			panic("unexpected type found")
		}
		res = append(res, &Plugin{Plugin: *p.DeepCopy()})
	}
	txn.Commit()
	return res, nil
}
