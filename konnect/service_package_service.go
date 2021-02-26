package konnect

import (
	"context"
	"encoding/json"
)

type ServicePackageService service

// List fetches a list of Service packages.
func (s *ServicePackageService) List(ctx context.Context,
	opt *ListOpt) ([]*ServicePackage, *ListOpt, error) {
	data, next, err := s.client.list(ctx, "/api/service_packages", opt)
	if err != nil {
		return nil, nil, err
	}
	var servicePackages []*ServicePackage

	for _, object := range data {
		b, err := object.MarshalJSON()
		if err != nil {
			return nil, nil, err
		}
		var servicePackage ServicePackage
		err = json.Unmarshal(b, &servicePackage)
		if err != nil {
			return nil, nil, err
		}
		servicePackages = append(servicePackages, &servicePackage)
	}

	return servicePackages, next, nil
}

// ListAll fetches all Service packages.
func (s *ServicePackageService) ListAll(ctx context.Context) ([]*ServicePackage,
	error) {
	var servicePackages, data []*ServicePackage
	var err error
	opt := &ListOpt{Size: pageSize}

	for opt != nil {
		data, opt, err = s.List(ctx, opt)
		if err != nil {
			return nil, err
		}
		servicePackages = append(servicePackages, data...)
	}
	return servicePackages, nil
}
