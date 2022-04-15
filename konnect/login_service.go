package konnect

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

type user struct {
	FullName string `json:"full_name,omitempty"`
}

type org struct {
	ID   string
	Name string
}

type UserInfo struct {
	Profile user
	ID      string
	Email   string
	Org     org
}

type AuthService service

func (s *AuthService) Login(ctx context.Context, email,
	password string,
) (AuthResponse, error) {
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
	url, _ := url.Parse(s.client.baseURL)
	jar, err := cookiejar.New(nil)
	if err != nil {
		return AuthResponse{}, err
	}

	jar.SetCookies(url, resp.Cookies())
	s.client.client.Jar = jar
	return authResponse, nil
}

func (s *AuthService) LoginV2(ctx context.Context, email,
	password string,
) (AuthResponse, error) {
	body := map[string]string{
		"username": email,
		"password": password,
	}
	req, err := s.client.NewRequest(http.MethodPost, authEndpointV2, nil, body)
	if err != nil {
		return AuthResponse{}, err
	}
	var authResponse AuthResponse
	resp, err := s.client.Do(ctx, req, &authResponse)
	if err != nil {
		return AuthResponse{}, err
	}
	url, _ := url.Parse(s.client.baseURL)
	jar, err := cookiejar.New(nil)
	if err != nil {
		return AuthResponse{}, err
	}

	jar.SetCookies(url, resp.Cookies())
	s.client.client.Jar = jar

	info, err := s.UserInfo(ctx)
	if err != nil {
		return AuthResponse{}, err
	}
	authResponse.FullName = info.Profile.FullName
	authResponse.Organization = info.Org.Name
	authResponse.OrganizationID = info.Org.ID
	return authResponse, nil
}

func (s *AuthService) UserInfo(ctx context.Context) (*UserInfo, error) {
	method := http.MethodGet
	req, err := s.client.NewRequest(method, "/konnect-api/api/userinfo/", nil, nil)
	if err != nil {
		return nil, err
	}

	info := &UserInfo{}
	_, err = s.client.Do(ctx, req, info)
	if err != nil {
		return nil, err
	}
	return info, nil
}
