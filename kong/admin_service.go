package kong

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

// TODO:
// Once Kong Roles are implemented, respective CRUD operations on
// those roles wrt Admins should be added here.
