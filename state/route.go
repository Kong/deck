package state

import (
	"fmt"

	"github.com/hashicorp/go-memdb"
	"github.com/pkg/errors"
)

var routeTableSchema = &memdb.TableSchema{
	Name: routeTableName,
	Indexes: map[string]*memdb.IndexSchema{
		id: &memdb.IndexSchema{
			Name:    id,
			Unique:  true,
			Indexer: &memdb.StringFieldIndex{Field: "ID"},
		},
		// TODO add ServiceName/ServiceID both fields for indexing
		"routesByServiceName": &memdb.IndexSchema{
			Name: "routesByServiceName",
			Indexer: &SubFieldIndexer{
				StructField: "Service",
				SubField:    "Name",
			},
		},
		"routesByServiceID": &memdb.IndexSchema{
			Name: "routesByServiceID",
			Indexer: &SubFieldIndexer{
				StructField: "Service",
				SubField:    "ID",
			},
		},
		all: &memdb.IndexSchema{
			Name: all,
			Indexer: &memdb.ConditionalIndex{
				Conditional: func(v interface{}) (bool, error) {
					return true, nil
				},
			},
		},
	},
}

// TODO add method to lookup a route based on service association,
// methods, hosts, paths -- Search by service name, then match the fields if equal or not

// AddRoute adds a route to KongState
func (k *KongState) AddRoute(route Route) error {
	txn := k.memdb.Txn(true)
	defer txn.Commit()
	err := txn.Insert(routeTableName, &route)
	if err != nil {
		return errors.Wrap(err, "insert failed")
	}
	return nil
}

// GetRoute gets a route by name or ID.
func (k *KongState) GetRoute(ID string) (*Route, error) {
	res, err := k.multiIndexLookup(routeTableName, []string{id}, ID)
	if err == ErrNotFound {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, errors.Wrap(err, "route lookup failed")
	}
	if res == nil {
		fmt.Println("res is nil")
	}
	route, ok := res.(*Route)
	if !ok {
		panic("unexpected type found")
	}
	return route, nil
}

func (k *KongState) GetAllRoutesByServiceName(name string) ([]*Route, error) {
	txn := k.memdb.Txn(false)
	iter, err := txn.Get(routeTableName, "routesByServiceName", name)
	if err != nil {
		return nil, err
	}
	var res []*Route
	for el := iter.Next(); el != nil; el = iter.Next() {
		s, ok := el.(*Route)
		if !ok {
			panic("unexpected type found")
		}
		res = append(res, s)
	}
	return res, nil
}

func (k *KongState) GetAllRoutesByServiceID(id string) ([]*Route, error) {
	txn := k.memdb.Txn(false)
	iter, err := txn.Get(routeTableName, "routesByServiceID", id)
	if err != nil {
		return nil, err
	}
	var res []*Route
	for el := iter.Next(); el != nil; el = iter.Next() {
		s, ok := el.(*Route)
		if !ok {
			panic("unexpected type found")
		}
		res = append(res, s)
	}
	return res, nil
}

// DeleteRoute deletes a route by name or ID.
func (k *KongState) DeleteRoute(route Route) error {
	txn := k.memdb.Txn(true)
	defer txn.Commit()

	err := txn.Delete(routeTableName, route)
	if err != nil {
		return errors.Wrap(err, "delete failed")
	}
	return nil
}

// GetAllRoutes gets a route by name or ID.
func (k *KongState) GetAllRoutes() ([]*Route, error) {
	txn := k.memdb.Txn(false)
	defer txn.Commit()

	iter, err := txn.Get(routeTableName, all, true)
	if err != nil {
		return nil, errors.Wrapf(err, "route lookup failed")
	}

	var res []*Route
	for el := iter.Next(); el != nil; el = iter.Next() {
		s, ok := el.(*Route)
		if !ok {
			panic("unexpected type found")
		}
		res = append(res, s)
	}
	return res, nil
}
