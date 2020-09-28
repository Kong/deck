package kong

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// AdminService handles Admins in Kong.
type AdminService service

// Invite creates an Admin in Kong.
func (s *AdminService) Invite(ctx context.Context,
	admin *Admin) (*Admin, error) {

	if admin == nil {
		return nil, errors.New("cannot create a nil admin")
	}

	endpoint := "/admins"
	method := "POST"
	req, err := s.client.NewRequest(method, endpoint, nil, admin)
	if err != nil {
		return nil, err
	}

	var createdAdmin struct {
		Admin Admin `json:"admin,omitempty" yaml:"admin,omitempty"`
	}
	_, err = s.client.Do(ctx, req, &createdAdmin)
	if err != nil {
		return nil, err
	}
	return &createdAdmin.Admin, nil
}

// Create aliases the Invite function as it performs
// essentially the same operation.
func (s *AdminService) Create(ctx context.Context,
	admin *Admin) (*Admin, error) {
	return s.Invite(ctx, admin)
}

// Get fetches a Admin in Kong.
func (s *AdminService) Get(ctx context.Context,
	nameOrID *string) (*Admin, error) {

	if isEmptyString(nameOrID) {
		return nil, errors.New("nameOrID cannot be nil for Get operation")
	}

	endpoint := fmt.Sprintf("/admins/%v", *nameOrID)

	req, err := s.client.NewRequest("GET", endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	var Admin Admin
	_, err = s.client.Do(ctx, req, &Admin)
	if err != nil {
		return nil, err
	}
	return &Admin, nil
}

// GenerateRegisterURL fetches an Admin in Kong
// and returns a unique registration URL for the Admin
func (s *AdminService) GenerateRegisterURL(ctx context.Context,
	nameOrID *string) (*Admin, error) {

	if isEmptyString(nameOrID) {
		return nil, errors.New("nameOrID cannot be nil for Get operation")
	}

	endpoint := fmt.Sprintf("/admins/%v?generate_register_url=true", *nameOrID)

	req, err := s.client.NewRequest("GET", endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	var Admin Admin
	_, err = s.client.Do(ctx, req, &Admin)
	if err != nil {
		return nil, err
	}
	return &Admin, nil
}

// Update updates an Admin in Kong.
func (s *AdminService) Update(ctx context.Context,
	admin *Admin) (*Admin, error) {

	if admin == nil {
		return nil, errors.New("cannot update a nil Admin")
	}

	if isEmptyString(admin.ID) {
		return nil, errors.New("ID cannot be nil for Update operation")
	}

	endpoint := fmt.Sprintf("/admins/%v", *admin.ID)
	req, err := s.client.NewRequest("PATCH", endpoint, nil, admin)
	if err != nil {
		return nil, err
	}

	var updatedAdmin Admin
	_, err = s.client.Do(ctx, req, &updatedAdmin)
	if err != nil {
		return nil, err
	}
	return &updatedAdmin, nil
}

// Delete deletes an Admin in Kong
func (s *AdminService) Delete(ctx context.Context,
	AdminOrID *string) error {

	if isEmptyString(AdminOrID) {
		return errors.New("AdminOrID cannot be nil for Delete operation")
	}

	endpoint := fmt.Sprintf("/admins/%v", *AdminOrID)
	req, err := s.client.NewRequest("DELETE", endpoint, nil, nil)
	if err != nil {
		return err
	}

	_, err = s.client.Do(ctx, req, nil)
	return err
}

// List fetches a list of all Admins in Kong.
func (s *AdminService) List(ctx context.Context,
	opt *ListOpt) ([]*Admin, *ListOpt, error) {

	data, next, err := s.client.list(ctx, "/admins/", opt)
	if err != nil {
		return nil, nil, err
	}
	var admins []*Admin
	for _, object := range data {
		b, err := object.MarshalJSON()
		if err != nil {
			return nil, nil, err
		}
		var admin Admin
		err = json.Unmarshal(b, &admin)
		if err != nil {
			return nil, nil, err
		}
		admins = append(admins, &admin)
	}

	return admins, next, nil
}

// RegisterCredentials registers credentials for existing Kong Admins
func (s *AdminService) RegisterCredentials(ctx context.Context,
	admin *Admin) error {

	if admin == nil {
		return errors.New("cannot register credentials for a nil Admin")
	}

	if isEmptyString(admin.Username) {
		return errors.New("Username cannot be nil for a registration operation")
	}
	if isEmptyString(admin.Email) {
		return errors.New("Email cannot be nil for a registration operation")
	}
	if isEmptyString(admin.Password) {
		return errors.New("Password cannot be nil for a registration operation")
	}

	req, err := s.client.NewRequest("POST", "/admins/register", nil, admin)
	if err != nil {
		return err
	}

	_, err = s.client.Do(ctx, req, nil)
	if err != nil {
		return err
	}
	return nil
}

// ListWorkspaces lists the workspaces associated with an admin
func (s *AdminService) ListWorkspaces(ctx context.Context,
	emailOrID *string) ([]*Workspace, error) {
	endpoint := fmt.Sprintf("/admins/%v/workspaces", *emailOrID)
	req, err := s.client.NewRequest("GET", endpoint, nil, nil)
	if err != nil {
		return nil, err
	}
	var workspaces []*Workspace
	_, err = s.client.Do(ctx, req, &workspaces)
	if err != nil {
		return nil, fmt.Errorf("error updating admin workspaces: %v", err)
	}
	return workspaces, nil
}

// ListRoles returns a slice of Kong RBAC roles associated with an Admin.
func (s *AdminService) ListRoles(ctx context.Context,
	emailOrID *string, opt *ListOpt) ([]*RBACRole, error) {

	endpoint := fmt.Sprintf("/admins/%v/roles", *emailOrID)
	req, err := s.client.NewRequest("GET", endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	var listRoles struct {
		Roles []*RBACRole `json:"roles,omitempty" yaml:"roles,omitempty"`
	}
	_, err = s.client.Do(ctx, req, &listRoles)
	if err != nil {
		return nil, fmt.Errorf("error listing admin roles: %v", err)
	}

	return listRoles.Roles, nil
}

// UpdateRoles creates or updates roles associated with an Admin
func (s *AdminService) UpdateRoles(ctx context.Context,
	emailOrID *string, roles []*RBACRole) ([]*RBACRole, error) {

	var updateRoles struct {
		NameOrID *string `json:"name_or_id,omitempty" yaml:"name_or_id,omitempty"`
		Roles    *string `json:"roles,omitempty" yaml:"roles,omitempty"`
	}

	// Flatten roles
	var r []string
	for _, role := range roles {
		r = append(r, *role.Name)
	}
	updateRoles.NameOrID = emailOrID
	updateRoles.Roles = String(strings.Join(r, ","))

	endpoint := fmt.Sprintf("/admins/%v/roles", *emailOrID)
	req, err := s.client.NewRequest("POST", endpoint, nil, updateRoles)
	if err != nil {
		return nil, err
	}
	var listRoles struct {
		Roles []*RBACRole `json:"roles,omitempty" yaml:"roles,omitempty"`
	}
	_, err = s.client.Do(ctx, req, &listRoles)
	if err != nil {
		return nil, fmt.Errorf("error updating admin roles: %v", err)
	}
	return listRoles.Roles, nil
}

// DeleteRoles deletes roles associated with an Admin
func (s *AdminService) DeleteRoles(ctx context.Context,
	emailOrID *string, roles []*RBACRole) error {

	var updateRoles struct {
		NameOrID *string `json:"name_or_id,omitempty" yaml:"name_or_id,omitempty"`
		Roles    *string `json:"roles,omitempty" yaml:"roles,omitempty"`
	}

	// Flatten roles
	var r []string
	for _, role := range roles {
		r = append(r, *role.Name)
	}
	updateRoles.NameOrID = emailOrID
	updateRoles.Roles = String(strings.Join(r, ","))

	endpoint := fmt.Sprintf("/admins/%v/roles", *emailOrID)
	req, err := s.client.NewRequest("DELETE", endpoint, nil, updateRoles)
	if err != nil {
		return err
	}

	_, err = s.client.Do(ctx, req, nil)
	if err != nil {
		return fmt.Errorf("error deleting admin roles: %v", err)
	}

	return nil
}

// GetConsumer fetches the Consumer that gets generated for an Admin when
// the Admin is created.
func (s *AdminService) GetConsumer(ctx context.Context,
	emailOrID *string) (*Consumer, error) {

	if isEmptyString(emailOrID) {
		return nil, errors.New("emailOrID cannot be nil for GetConsumer operation")
	}

	endpoint := fmt.Sprintf("/admins/%v/consumer", *emailOrID)

	req, err := s.client.NewRequest("GET", endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	var consumer Consumer
	_, err = s.client.Do(ctx, req, &consumer)
	if err != nil {
		return nil, err
	}
	return &consumer, nil
}
