package konnect

import (
	"context"
	"net/http"
)

type ServiceVersionService service

// List fetches a list of Service Versions for a given servicePackageID.
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
	_, err = s.client.Do(ctx, req, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}
