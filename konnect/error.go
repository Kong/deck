package konnect

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

func hasError(res *http.Response) error {
	if res.StatusCode >= 200 && res.StatusCode <= 399 {
		return nil
	}

	body, _ := ioutil.ReadAll(res.Body) // TODO error in error?
	return &APIError{
		httpCode: res.StatusCode,
		message:  messageFromBody(body),
	}
}

func messageFromBody(b []byte) string {
	s := struct {
		Message string
	}{}

	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Sprintf("<failed to parse response body: %v>", err)
	}

	return s.Message
}

// APIError is used for Kong Admin API errors.
type APIError struct {
	httpCode int
	message  string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("HTTP status %d (message: %q)", e.httpCode, e.message)
}

// Code returns the HTTP status code for the error.
func (e *APIError) Code() int {
	return e.httpCode
}

// IsNotFoundErr returns true if the error or it's cause is
// a 404 response from Kong.
func IsNotFoundErr(e error) bool {
	var apiErr *APIError
	return errors.As(e, &apiErr) && apiErr.httpCode == http.StatusNotFound
}

// IsUnauthorizedErr returns true if the error or it's cause is
// a 401 response from Konnect.
func IsUnauthorizedErr(e error) bool {
	var apiErr *APIError
	return errors.As(e, &apiErr) && apiErr.httpCode == http.StatusUnauthorized
}
