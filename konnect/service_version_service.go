package konnect

import (
	"context"
	"fmt"
	"net/http"
)

type ServiceVersionService service

// Create creates a ServiceVersion in Konnect.
func (s *ServiceVersionService) Create(ctx context.Context,
	sv *ServiceVersion) (*ServiceVersion, error) {

	if sv == nil {
		return nil, fmt.Errorf("cannot create a nil service-package")
	}

	endpoint := "/api/service_versions"
	method := "POST"

	if !emptyString(sv.ID) {
		method = "PUT"
		endpoint += "/" + *sv.ID

	}

	req, err := s.client.NewRequest(method, endpoint, nil, map[string]string{
		"version":         *sv.Version,
		"service_package": *sv.ServicePackage.ID,
		"control_plane":   s.controlPlaneID,
	})
	if err != nil {
		return nil, err
	}

	var createdSV ServiceVersion
	resp, err := s.client.Do(ctx, req, &createdSV)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return &createdSV, nil
}

// Delete deletes a ServiceVersion in Konnect.
func (s *ServiceVersionService) Delete(ctx context.Context, id *string) error {
	if emptyString(id) {
		return fmt.Errorf("id cannot be nil for Delete operation")
	}

	endpoint := fmt.Sprintf("/api/service_versions/%v", *id)
	req, err := s.client.NewRequest("DELETE", endpoint, nil, nil)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(ctx, req, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return err
}

// Update updates a ServiceVersion in Konnect.
func (s *ServiceVersionService) Update(ctx context.Context,
	sv *ServiceVersion) (*ServiceVersion, error) {

	if sv == nil {
		return nil, fmt.Errorf("cannot update a nil service-package")
	}

	if emptyString(sv.ID) {
		return nil, fmt.Errorf("ID cannot be nil for Update operation")
	}

	endpoint := fmt.Sprintf("/api/service_versions/%v", *sv.ID)
	req, err := s.client.NewRequest("PATCH", endpoint, nil, sv)
	if err != nil {
		return nil, err
	}

	var updatedSV ServiceVersion
	resp, err := s.client.Do(ctx, req, &updatedSV)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return &updatedSV, nil
}

// ListForPackage fetches a list of Service Versions for a given servicePackageID.
func (s *ServiceVersionService) ListForPackage(ctx context.Context,
	servicePackageID *string) ([]ServiceVersion, error) {
	endpoint := "/api/service_packages/" + *servicePackageID + "/service_versions"
	req, err := s.client.NewRequest(http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return nil, err
	}
	// Note: This endpoint doesn't follow the structure of paginated endpoints
	// and instead returns an array with all service versions.
	var response []ServiceVersion
	resp, err := s.client.Do(ctx, req, &response)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return response, nil
}
