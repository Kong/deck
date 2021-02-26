package konnect

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"net/url"
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
	resp, err := s.client.Do(ctx, req, &authResponse)
	if err != nil {
		return AuthResponse{}, err
	}
	url, _ := url.Parse(BaseURL())
	jar, err := cookiejar.New(nil)
	if err != nil {
		return AuthResponse{}, err
	}

	jar.SetCookies(url, resp.Cookies())
	s.client.client.Jar = jar
	return authResponse, nil
}
