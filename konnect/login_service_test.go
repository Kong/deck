package konnect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetGlobalAuthEndpoint(t *testing.T) {
	tests := []struct {
		baseURL  string
		expected string
	}{
		{
			baseURL:  "https://us.api.konghq.com",
			expected: "https://global.api.konghq.com/kauth/api/v1/authenticate",
		},
		{
			baseURL:  "https://global.api.konghq.com",
			expected: "https://global.api.konghq.com/kauth/api/v1/authenticate",
		},
		{
			baseURL:  "https://eu.api.konghq.com",
			expected: "https://global.api.konghq.com/kauth/api/v1/authenticate",
		},
		{
			baseURL:  "https://api.konghq.com",
			expected: "https://global.api.konghq.com/kauth/api/v1/authenticate",
		},
		{
			baseURL:  "https://eu.api.konghq.test",
			expected: "https://global.api.konghq.test/kauth/api/v1/authenticate",
		},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, getGlobalAuthEndpoint(tt.baseURL))
	}
}
