package konnect

import (
	"context"
	"net/http"
)

type AuthResponse struct {
	Organization   string `json:"org_name"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	OrganizationID string `json:"org_id"`
}

type AuthService service

func (s *AuthService) Login(ctx context.Context, email,
	password string) (AuthResponse, error) {

	body := map[string]string{
		"username": email,
		"password": password,
	}
	req, err := s.client.NewRequest(http.MethodPost, authEndpoint, nil, body)
	if err != nil {
		return AuthResponse{}, err
	}
	var authResponse AuthResponse
	_, err = s.client.Do(ctx, req, &authResponse)
	if err != nil {
		return AuthResponse{}, err
	}

	if err != nil {
		return AuthResponse{}, nil
	}
	return authResponse, nil
}
