package konnect

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

type OrgUserInfo struct {
	Name  string `json:"name,omitempty"`
	OrgID string `json:"id,omitempty"`
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

// getGlobalAuthEndpoint returns the global auth endpoint
// given a base Konnect URL.
func getGlobalAuthEndpoint(baseURL string) string {
	parts := strings.Split(baseURL, "api.konghq")
	return baseEndpointUS + parts[len(parts)-1] + authEndpointV2
}

func createAuthRequest(baseURL, email, password string) (*http.Request, error) {
	var (
		buf []byte
		err error

		body = map[string]string{
			"username": email,
			"password": password,
		}
	)

	buf, err = json.Marshal(body)
	if err != nil {
		return nil, err
	}

	endpoint := getGlobalAuthEndpoint(baseURL)
	return http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(buf))
}

func (s *AuthService) sessionAuth(ctx context.Context, email,
	password string,
) (AuthResponse, error) {
	req, err := createAuthRequest(s.client.baseURL, email, password)
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
	password, token string,
) (AuthResponse, error) {
	var (
		err          error
		authResponse AuthResponse
	)

	if token != "" {
		s.client.token = token
	} else if email != "" && password != "" {
		authResponse, err = s.sessionAuth(ctx, email, password)
		if err != nil {
			return AuthResponse{}, err
		}
	} else {
		return AuthResponse{}, errors.New(
			"at least one of email/password or personal access token must be provided",
		)
	}

	info, err := s.OrgUserInfo(ctx)
	if err != nil {
		return AuthResponse{}, err
	}
	authResponse.Name = info.Name
	authResponse.OrganizationID = info.OrgID
	return authResponse, nil
}

func (s *AuthService) OrgUserInfo(ctx context.Context) (*OrgUserInfo, error) {
	method := http.MethodGet
	req, err := s.client.NewRequest(method, "/v2/organizations/me", nil, nil)
	if err != nil {
		return nil, err
	}

	info := &OrgUserInfo{}
	_, err = s.client.Do(ctx, req, info)
	if err != nil {
		return nil, err
	}
	return info, nil
}
