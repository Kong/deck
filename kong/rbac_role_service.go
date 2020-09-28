package kong

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// RBACRoleService handles Roles in Kong.
type RBACRoleService service

// Create creates a Role in Kong.
func (s *RBACRoleService) Create(ctx context.Context,
	role *RBACRole) (*RBACRole, error) {

	if role == nil {
		return nil, errors.New("cannot create a nil role")
	}

	endpoint := "/rbac/roles"
	method := "POST"
	if role.ID != nil {
		endpoint = endpoint + "/" + *role.ID
		method = "PUT"
	}
	req, err := s.client.NewRequest(method, endpoint, nil, role)

	if err != nil {
		return nil, err
	}

	var createdRole RBACRole
	_, err = s.client.Do(ctx, req, &createdRole)
	if err != nil {
		return nil, err
	}
	return &createdRole, nil
}

// Get fetches a Role in Kong.
func (s *RBACRoleService) Get(ctx context.Context,
	nameOrID *string) (*RBACRole, error) {

	if isEmptyString(nameOrID) {
		return nil, errors.New("nameOrID cannot be nil for Get operation")
	}

	endpoint := fmt.Sprintf("/rbac/roles/%v", *nameOrID)
	req, err := s.client.NewRequest("GET", endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	var Role RBACRole
	_, err = s.client.Do(ctx, req, &Role)
	if err != nil {
		return nil, err
	}
	return &Role, nil
}

// Update updates a Role in Kong.
func (s *RBACRoleService) Update(ctx context.Context,
	role *RBACRole) (*RBACRole, error) {

	if role == nil {
		return nil, errors.New("cannot update a nil Role")
	}

	if isEmptyString(role.ID) {
		return nil, errors.New("ID cannot be nil for Update operation")
	}

	endpoint := fmt.Sprintf("/rbac/roles/%v", *role.ID)
	req, err := s.client.NewRequest("PATCH", endpoint, nil, role)
	if err != nil {
		return nil, err
	}

	var updatedRole RBACRole
	_, err = s.client.Do(ctx, req, &updatedRole)
	if err != nil {
		return nil, err
	}
	return &updatedRole, nil
}

// Delete deletes a Role in Kong
func (s *RBACRoleService) Delete(ctx context.Context,
	RoleOrID *string) error {

	if isEmptyString(RoleOrID) {
		return errors.New("RoleOrID cannot be nil for Delete operation")
	}

	endpoint := fmt.Sprintf("/rbac/roles/%v", *RoleOrID)
	req, err := s.client.NewRequest("DELETE", endpoint, nil, nil)
	if err != nil {
		return err
	}

	_, err = s.client.Do(ctx, req, nil)
	return err
}

// List fetches a list of all Roles in Kong.
func (s *RBACRoleService) List(ctx context.Context) ([]*RBACRole, error) {

	data, _, err := s.client.list(ctx, "/rbac/roles/", nil)
	if err != nil {
		return nil, err
	}
	var roles []*RBACRole
	for _, object := range data {
		b, err := object.MarshalJSON()
		if err != nil {
			return nil, err
		}
		var role RBACRole
		err = json.Unmarshal(b, &role)
		if err != nil {
			return nil, err
		}
		roles = append(roles, &role)
	}

	return roles, nil
}
