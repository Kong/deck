package state

import (
	"fmt"

	memdb "github.com/hashicorp/go-memdb"

	"github.com/kong/deck/state/indexers"
	"github.com/kong/deck/utils"
)

const (
	routeTableName    = "route"
	routesByServiceID = "routesByServiceID"
)

var routeTableSchema = &memdb.TableSchema{
	Name: routeTableName,
	Indexes: map[string]*memdb.IndexSchema{
		"id": {
			Name:    "id",
			Unique:  true,
			Indexer: &memdb.StringFieldIndex{Field: "ID"},
		},
		"name": {
			Name:         "name",
			Unique:       true,
			Indexer:      &memdb.StringFieldIndex{Field: "Name"},
			AllowMissing: true,
		},
		all: allIndex,
		// foreign
		routesByServiceID: {
			Name: routesByServiceID,
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
	},
}

// RoutesCollection stores and indexes Kong Routes.
type RoutesCollection collection

// Add adds a route into RoutesCollection
// route.ID should not be nil else an error is thrown.
func (k *RoutesCollection) Add(route Route) error {
	// TODO abstract this check in the go-memdb library itself
	if utils.Empty(route.ID) {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	var searchBy []string
	searchBy = append(searchBy, *route.ID)
	if !utils.Empty(route.Name) {
		searchBy = append(searchBy, *route.Name)
	}
	_, err := getRoute(txn, searchBy...)
	if err == nil {
		return fmt.Errorf("inserting route %v: %w", route.Console(), ErrAlreadyExists)
	} else if err != ErrNotFound {
		return err
	}

	err = txn.Insert(routeTableName, &route)
	if err != nil {
		return err
	}
	txn.Commit()
	return nil
}

func getRoute(txn *memdb.Txn, IDs ...string) (*Route, error) {
	for _, id := range IDs {
		res, err := multiIndexLookupUsingTxn(txn, routeTableName,
			[]string{"name", "id"}, id)
		if err == ErrNotFound {
			continue
		}
		if err != nil {
			return nil, err
		}

		route, ok := res.(*Route)
		if !ok {
			panic(unexpectedType)
		}
		return &Route{Route: *route.DeepCopy()}, nil
	}
	return nil, ErrNotFound
}

// Get gets a route by name or ID.
func (k *RoutesCollection) Get(nameOrID string) (*Route, error) {
	if nameOrID == "" {
		return nil, errIDRequired
	}

	txn := k.db.Txn(false)
	defer txn.Abort()
	route, err := getRoute(txn, nameOrID)
	if err != nil {
		return nil, err
	}
	return route, nil
}

// Update updates a route
func (k *RoutesCollection) Update(route Route) error {
	// TODO abstract this check in the go-memdb library itself
	if utils.Empty(route.ID) {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteRoute(txn, *route.ID)
	if err != nil {
		return err
	}

	err = txn.Insert(routeTableName, &route)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func deleteRoute(txn *memdb.Txn, nameOrID string) error {
	route, err := getRoute(txn, nameOrID)
	if err != nil {
		return err
	}

	err = txn.Delete(routeTableName, route)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes a route by name or ID.
func (k *RoutesCollection) Delete(nameOrID string) error {
	if nameOrID == "" {
		return errIDRequired
	}

	txn := k.db.Txn(true)
	defer txn.Abort()

	err := deleteRoute(txn, nameOrID)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

// GetAll gets a route by name or ID.
func (k *RoutesCollection) GetAll() ([]*Route, error) {
	txn := k.db.Txn(false)
	defer txn.Abort()

	iter, err := txn.Get(routeTableName, all, true)
	if err != nil {
		return nil, err
	}

	var res []*Route
	for el := iter.Next(); el != nil; el = iter.Next() {
		r, ok := el.(*Route)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &Route{Route: *r.DeepCopy()})
	}
	txn.Commit()
	return res, nil
}

// GetAllByServiceID returns all routes referencing a service
// by its id.
func (k *RoutesCollection) GetAllByServiceID(id string) ([]*Route,
	error,
) {
	txn := k.db.Txn(false)
	iter, err := txn.Get(routeTableName, routesByServiceID, id)
	if err != nil {
		return nil, err
	}
	var res []*Route
	for el := iter.Next(); el != nil; el = iter.Next() {
		r, ok := el.(*Route)
		if !ok {
			panic(unexpectedType)
		}
		res = append(res, &Route{Route: *r.DeepCopy()})
	}
	return res, nil
}
