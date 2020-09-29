package kong

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// RBACEndpointPermissionService handles RBACEndpointPermissions in Kong.
type RBACEndpointPermissionService service

// Create creates a RBACEndpointPermission in Kong.
func (s *RBACEndpointPermissionService) Create(ctx context.Context,
	ep *RBACEndpointPermission) (*RBACEndpointPermission, error) {

	if ep == nil {
		return nil, errors.New("cannot create a nil endpointpermission")
	}
	if ep.Role == nil || ep.Role.ID == nil {
		return nil, errors.New("cannot create endpoint permission with role or role id undefined")
	}

	method := "POST"
	endpoint := fmt.Sprintf("/rbac/roles/%v/endpoints", *ep.Role.ID)
	req, err := s.client.NewRequest(method, endpoint, nil, ep)

	if err != nil {
		return nil, err
	}

	var createdEndpointPermission RBACEndpointPermission

	_, err = s.client.Do(ctx, req, &createdEndpointPermission)
	if err != nil {
		return nil, err
	}
	return &createdEndpointPermission, nil
}

// Get fetches a RBACEndpointPermission in Kong.
func (s *RBACEndpointPermissionService) Get(ctx context.Context,
	roleNameOrID *string, workspaceNameOrID *string, endpointName *string) (*RBACEndpointPermission, error) {

	if isEmptyString(endpointName) {
		return nil, errors.New("endpointName cannot be nil for Get operation")
	}
	if *endpointName == "*" {
		endpointName = String("/" + *endpointName)
	}
	endpoint := fmt.Sprintf("/rbac/roles/%v/endpoints/%v%v", *roleNameOrID, *workspaceNameOrID, *endpointName)
	req, err := s.client.NewRequest("GET", endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	var EndpointPermission RBACEndpointPermission
	_, err = s.client.Do(ctx, req, &EndpointPermission)
	if err != nil {
		return nil, err
	}
	return &EndpointPermission, nil
}

// Update updates a RBACEndpointPermission in Kong.
func (s *RBACEndpointPermissionService) Update(ctx context.Context,
	ep *RBACEndpointPermission) (*RBACEndpointPermission, error) {

	if ep == nil {
		return nil, errors.New("cannot update a nil EndpointPermission")
	}
	if ep.Workspace == nil {
		return nil, errors.New("cannot update an EndpointPermission with workspace as nil")
	}
	if ep.Role == nil || ep.Role.ID == nil {
		return nil, errors.New("cannot create endpoint permission with role or role id undefined")
	}

	if isEmptyString(ep.Endpoint) {
		return nil, errors.New("ID cannot be nil for Update operation")
	}

	endpoint := fmt.Sprintf("/rbac/roles/%v/endpoints/%v/%v",
		*ep.Role.ID, *ep.Workspace, *ep.Endpoint)
	req, err := s.client.NewRequest("PATCH", endpoint, nil, ep)
	if err != nil {
		return nil, err
	}

	var updatedEndpointPermission RBACEndpointPermission
	_, err = s.client.Do(ctx, req, &updatedEndpointPermission)
	if err != nil {
		return nil, err
	}
	return &updatedEndpointPermission, nil
}

// Delete deletes a EndpointPermission in Kong
func (s *RBACEndpointPermissionService) Delete(ctx context.Context,
	roleNameOrID *string, workspaceNameOrID *string, endpoint *string) error {

	if endpoint == nil {
		return errors.New("cannot update a nil EndpointPermission")
	}
	if workspaceNameOrID == nil {
		return errors.New("cannot update an EndpointPermission with workspace as nil")
	}
	if roleNameOrID == nil {
		return errors.New("cannot update an EndpointPermission with role as nil")
	}

	reqEndpoint := fmt.Sprintf("/rbac/roles/%v/endpoints/%v/%v",
		*roleNameOrID, *workspaceNameOrID, *endpoint)
	req, err := s.client.NewRequest("DELETE", reqEndpoint, nil, nil)
	if err != nil {
		return err
	}

	_, err = s.client.Do(ctx, req, nil)
	return err
}

// ListAllForRole fetches a list of all RBACEndpointPermissions in Kong for a given role.
func (s *RBACEndpointPermissionService) ListAllForRole(ctx context.Context,
	roleNameOrID *string) ([]*RBACEndpointPermission, error) {

	data, _, err := s.client.list(ctx, fmt.Sprintf("/rbac/roles/%v/endpoints", *roleNameOrID), nil)
	if err != nil {
		return nil, err
	}
	var eps []*RBACEndpointPermission
	for _, object := range data {
		b, err := object.MarshalJSON()
		if err != nil {
			return nil, err
		}
		var ep RBACEndpointPermission
		err = json.Unmarshal(b, &ep)
		if err != nil {
			return nil, err
		}
		eps = append(eps, &ep)
	}

	return eps, nil
}
