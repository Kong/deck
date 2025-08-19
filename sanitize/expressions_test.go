package sanitize

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeExpression(t *testing.T) {
	s := &Sanitizer{
		salt:         "test-salt-123",
		sanitizedMap: make(map[string]interface{}),
	}

	tests := []struct {
		name       string
		expression string
		expected   string
	}{
		{
			name:       "Sanitize simple path in expression",
			expression: `http.path == "/api/users"`,
			expected:   `http.path == "ab1c696d4b37997a1c29a09610f47a7032949190db6c78d06a8d09b050ab91ea"`,
		},
		{
			name:       "Preserve method and protocol in expression",
			expression: `net.protocol == "http" && http.method == "GET"`,
			expected:   `net.protocol == "http"  &&  http.method == "GET"`,
		},
		{
			name:       "Sanitize complex expression",
			expression: `(http.path == "/api/users" || http.path == "/api/products") && http.method == "POST"`,
			expected: ` ( http.path == "ab1c696d4b37997a1c29a09610f47a7032949190db6c78d06a8d09b050ab91ea"  || ` +
				` http.path == "af59ca84d7d64ca44795aa3c0217c5ebe63f2ee593a67b3d77fad31bf3b84372" )   && ` +
				` http.method == "POST"`,
		},
		{
			name:       "Sanitize expression with regex",
			expression: `http.path ~ r#"^/users/\d+$"#`,
			expected:   `http.path ~ r#"dab30f7e826083df3eee6538650e002873dc7eae7cd8d617fc4dd2de2b97000b"#`,
		},
		{
			name:       "Sanitize expression with IPv4 address",
			expression: `http.host == 192.168.1.1`,
			expected:   `http.host == 10.54.56.50`,
		},
		{
			name:       "Sanitize expression with CIDR notation",
			expression: `http.host == 192.168.1.0/24`,
			expected:   `http.host == 10.53.51.54/24`,
		},
		{
			name:       "Sanitize expression with IPv6 address",
			expression: `http.host == 2001:db8::1`,
			expected:   `http.host == fd00:3561:6333:3965:3331:6636:6135:3132`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.sanitizeExpression(tt.expression)
			assert.Equal(t, tt.expected, result)
		})
	}
}
