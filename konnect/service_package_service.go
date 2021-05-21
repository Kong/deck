package konnect

import (
	"context"
	"encoding/json"
	"fmt"
)

type ServicePackageService service

// Create creates a ServicePackage in Konnect.
func (s *ServicePackageService) Create(ctx context.Context,
	sp *ServicePackage) (*ServicePackage, error) {

	if sp == nil {
		return nil, fmt.Errorf("cannot create a nil service-package")
	}

	endpoint := "/api/service_packages"
	method := "POST"
	req, err := s.client.NewRequest(method, endpoint, nil, sp)
	if err != nil {
		return nil, err
	}

	var createdSP ServicePackage
	_, err = s.client.Do(ctx, req, &createdSP)
	if err != nil {
		return nil, err
	}
	return &createdSP, nil
}

// Delete deletes a ServicePackage in Konnect.
func (s *ServicePackageService) Delete(ctx context.Context, id *string) error {
	if emptyString(id) {
		return fmt.Errorf("id cannot be nil for Delete operation")
	}

	endpoint := fmt.Sprintf("/api/service_packages/%v", *id)
	req, err := s.client.NewRequest("DELETE", endpoint, nil, nil)
	if err != nil {
		return err
	}

	_, err = s.client.Do(ctx, req, nil)
	return err
}

// Update updates a ServicePackage in Konnect.
func (s *ServicePackageService) Update(ctx context.Context,
	sp *ServicePackage) (*ServicePackage, error) {

	if sp == nil {
		return nil, fmt.Errorf("cannot update a nil service-package")
	}

	if emptyString(sp.ID) {
		return nil, fmt.Errorf("ID cannot be nil for Update operation")
	}

	endpoint := fmt.Sprintf("/api/service_packages/%v", *sp.ID)
	req, err := s.client.NewRequest("PATCH", endpoint, nil, sp)
	if err != nil {
		return nil, err
	}

	var updatedSP ServicePackage
	_, err = s.client.Do(ctx, req, &updatedSP)
	if err != nil {
		return nil, err
	}
	return &updatedSP, nil
}

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
