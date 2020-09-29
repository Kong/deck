package kong

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// RBACUserService handles Users in Kong.
type RBACUserService service

// Create creates an RBAC User in Kong.
func (s *RBACUserService) Create(ctx context.Context,
	user *RBACUser) (*RBACUser, error) {

	if user == nil {
		return nil, errors.New("cannot create a nil user")
	}

	endpoint := "/rbac/users"
	method := "POST"
	if user.ID != nil {
		endpoint = endpoint + "/" + *user.ID
		method = "PUT"
	}
	req, err := s.client.NewRequest(method, endpoint, nil, user)

	if err != nil {
		return nil, err
	}

	var createdUser RBACUser
	_, err = s.client.Do(ctx, req, &createdUser)
	if err != nil {
		return nil, err
	}
	return &createdUser, nil
}

// Get fetches a User in Kong.
func (s *RBACUserService) Get(ctx context.Context,
	nameOrID *string) (*RBACUser, error) {

	if isEmptyString(nameOrID) {
		return nil, errors.New("nameOrID cannot be nil for Get operation")
	}

	endpoint := fmt.Sprintf("/rbac/users/%v", *nameOrID)
	req, err := s.client.NewRequest("GET", endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	var RBACUser RBACUser
	_, err = s.client.Do(ctx, req, &RBACUser)
	if err != nil {
		return nil, err
	}
	return &RBACUser, nil
}

// Update updates a User in Kong.
func (s *RBACUserService) Update(ctx context.Context,
	user *RBACUser) (*RBACUser, error) {

	if user == nil {
		return nil, errors.New("cannot update a nil User")
	}

	if isEmptyString(user.ID) && isEmptyString(user.Name) {
		return nil, errors.New("ID and Name cannot both be nil for Update operation")
	}

	endpoint := fmt.Sprintf("/rbac/users/%v", *user.ID)
	req, err := s.client.NewRequest("PATCH", endpoint, nil, user)
	if err != nil {
		return nil, err
	}

	var updatedUser RBACUser
	_, err = s.client.Do(ctx, req, &updatedUser)
	if err != nil {
		return nil, err
	}
	return &updatedUser, nil
}

// Delete deletes a User in Kong
func (s *RBACUserService) Delete(ctx context.Context,
	userOrID *string) error {

	if isEmptyString(userOrID) {
		return errors.New("UserOrID cannot be nil for Delete operation")
	}

	endpoint := fmt.Sprintf("/rbac/users/%v", *userOrID)
	req, err := s.client.NewRequest("DELETE", endpoint, nil, nil)
	if err != nil {
		return err
	}

	_, err = s.client.Do(ctx, req, nil)
	return err
}

// List fetches a list of Users in Kong.
// opt can be used to control pagination.
func (s *RBACUserService) List(ctx context.Context,
	opt *ListOpt) ([]*RBACUser, *ListOpt, error) {

	data, next, err := s.client.list(ctx, "/rbac/users/", opt)
	if err != nil {
		return nil, nil, err
	}
	var users []*RBACUser
	for _, object := range data {
		b, err := object.MarshalJSON()
		if err != nil {
			return nil, nil, err
		}
		var user RBACUser
		err = json.Unmarshal(b, &user)
		if err != nil {
			return nil, nil, err
		}
		users = append(users, &user)
	}

	return users, next, nil
}

// ListAll fetches all users in Kong.
func (s *RBACUserService) ListAll(ctx context.Context) ([]*RBACUser, error) {

	var users, data []*RBACUser
	var err error
	opt := &ListOpt{Size: pageSize}

	for opt != nil {
		data, opt, err = s.List(ctx, opt)
		if err != nil {
			return nil, err
		}
		users = append(users, data...)
	}

	return users, nil
}

// AddRoles adds a comma separated list of roles to a User.
func (s *RBACUserService) AddRoles(ctx context.Context,
	nameOrID *string, roles []*RBACRole) ([]*RBACRole, error) {

	var updateRoles struct {
		NameOrID *string `json:"name_or_id,omitempty" yaml:"name_or_id,omitempty"`
		Roles    *string `json:"roles,omitempty" yaml:"roles,omitempty"`
	}

	// Flatten roles
	var r []string
	for _, role := range roles {
		r = append(r, *role.Name)
	}
	updateRoles.NameOrID = nameOrID
	updateRoles.Roles = String(strings.Join(r, ","))

	endpoint := fmt.Sprintf("/rbac/users/%v/roles", *nameOrID)
	req, err := s.client.NewRequest("POST", endpoint, nil, updateRoles)
	if err != nil {
		return nil, err
	}
	var listRoles struct {
		Roles []*RBACRole `json:"roles,omitempty" yaml:"roles,omitempty"`
		User  *RBACUser   `json:"user,omitempty" yaml:"user,omitempty"`
	}
	_, err = s.client.Do(ctx, req, &listRoles)
	if err != nil {
		return nil, fmt.Errorf("error updating roles: %v", err)
	}
	return listRoles.Roles, nil
}

// DeleteRoles deletes roles associated with a User
func (s *RBACUserService) DeleteRoles(ctx context.Context,
	nameOrID *string, roles []*RBACRole) error {

	var updateRoles struct {
		NameOrID *string `json:"name_or_id,omitempty" yaml:"name_or_id,omitempty"`
		Roles    *string `json:"roles,omitempty" yaml:"roles,omitempty"`
	}

	// Flatten roles
	var r []string
	for _, role := range roles {
		r = append(r, *role.Name)
	}
	updateRoles.NameOrID = nameOrID
	updateRoles.Roles = String(strings.Join(r, ","))

	endpoint := fmt.Sprintf("/rbac/users/%v/roles", *nameOrID)
	req, err := s.client.NewRequest("DELETE", endpoint, nil, updateRoles)
	if err != nil {
		return err
	}

	_, err = s.client.Do(ctx, req, nil)
	if err != nil {
		return fmt.Errorf("error deleting roles: %v", err)
	}

	return nil
}

// ListRoles returns a slice of Kong RBAC roles associated with a User.
func (s *RBACUserService) ListRoles(ctx context.Context,
	nameOrID *string) ([]*RBACRole, error) {

	endpoint := fmt.Sprintf("/rbac/users/%v/roles", *nameOrID)
	req, err := s.client.NewRequest("GET", endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	var listRoles struct {
		Roles []*RBACRole `json:"roles,omitempty" yaml:"roles,omitempty"`
	}
	_, err = s.client.Do(ctx, req, &listRoles)
	if err != nil {
		return nil, fmt.Errorf("error retrieving list of roles: %v", err)
	}
	return listRoles.Roles, nil
}

// ListPermissions returns the entity and endpoint permissions associated with a user.
func (s *RBACUserService) ListPermissions(ctx context.Context,
	nameOrID *string) (*RBACPermissionsList, error) {

	endpoint := fmt.Sprintf("/rbac/users/%v/permissions", *nameOrID)
	req, err := s.client.NewRequest("GET", endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	var permissionsList RBACPermissionsList
	_, err = s.client.Do(ctx, req, &permissionsList)
	if err != nil {
		return nil, fmt.Errorf("error retrieving list of permissions for role: %v", err)
	}

	return &permissionsList, nil
}
