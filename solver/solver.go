package solver

import (
	"github.com/hbagdi/go-kong/kong"
	"github.com/kong/deck/crud"
	cruds "github.com/kong/deck/crud/kong"
	drycrud "github.com/kong/deck/crud/kong/dry"
	"github.com/kong/deck/diff"
	"github.com/pkg/errors"
)

// Solve generates a diff and walks the graph.
func Solve(syncer *diff.Syncer, client *kong.Client, dry bool) error {
	var r *crud.Registry
	var err error
	if dry {
		r, err = buildDryRegistry()
	} else {
		r, err = buildRegistry()
	}
	if err != nil {
		return errors.Wrapf(err, "cannot build registry")
	}
	gDelete, gCreateUpdate, err := syncer.Diff()
	if err != nil {
		return err
	}
	err = syncer.Solve(gDelete, client, *r)
	if err != nil {
		return err
	}
	err = syncer.Solve(gCreateUpdate, client, *r)
	if err != nil {
		return err
	}
	return nil
}

func buildDryRegistry() (*crud.Registry, error) {
	var r crud.Registry
	err := r.Register("service", &drycrud.ServiceCRUD{})
	if err != nil {
		return nil, errors.Wrapf(err, "registering 'service' crud")
	}
	err = r.Register("route", &drycrud.RouteCRUD{})
	if err != nil {
		return nil, errors.Wrapf(err, "registering 'route' crud")
	}
	return &r, nil
}

func buildRegistry() (*crud.Registry, error) {
	var r crud.Registry
	err := r.Register("service", &cruds.ServiceCRUD{})
	if err != nil {
		return nil, errors.Wrapf(err, "registering 'service' crud")
	}
	err = r.Register("route", &cruds.RouteCRUD{})
	if err != nil {
		return nil, errors.Wrapf(err, "registering 'route' crud")
	}
	return &r, nil
}
