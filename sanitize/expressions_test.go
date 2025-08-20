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
			name:       "Sanitize expression with contain",
			expression: `http.path contains "foo"`,
			expected:   `http.path contains "6112060752768f93013e72488fb87d7e08e87753d8bce33951ee628553c49f8a"`,
		},
		{
			name:       "Sanitize expression with complex fields",
			expression: `http.path.segments.0_1 == "foo"`,
			expected:   `http.path.segments.0_1 == "6112060752768f93013e72488fb87d7e08e87753d8bce33951ee628553c49f8a"`,
		},
		{
			name:       "Sanitize expression with regex",
			expression: `http.path ~ r#"^/users/\d+$"#`,
			expected:   `http.path ~ r#"dab30f7e826083df3eee6538650e002873dc7eae7cd8d617fc4dd2de2b97000b"#`,
		},
		{
			name:       "Sanitize complex regex expression",
			expression: `http.headers.Accept ~ r#"^application\/sample\.example\.[vV](1.1.2.3)(.\d+)*?\+json$"#`,
			expected:   `http.headers.Accept ~ r#"c609266d1491d914fb9724f40a232d6830102c2559c7c5514afc2e084639126c"#`,
		},
		{
			name: "Sanitize expression with multiple logical operators and regex",
			expression: `(http.path ^= "/example/{query}" &&` +
				`http.method == "PATCH") && http.headers.Accept ~ r#"^application\[vV](1.1.2.3)(.\d+)*?\+json$"#`,
			expected: ` ( http.path ^= "a64a55a0d8d091518739d983c56c5ee3d2aa41a54c1906cc9bfade3a9462478d"  && ` +
				`http.method == "PATCH" )   &&  ` +
				`http.headers.Accept ~ r#"98699493aea14c6936387ed103ad746e2155b28ff055b3e2c7c0be0fac622b21"#`,
		},
		{
			name:       "Sanitize expression with IPv4 address",
			expression: `net.src.ip == 192.168.1.1`,
			expected:   `net.src.ip == 10.54.56.50`,
		},
		{
			name:       "Sanitize expression with CIDR notation",
			expression: `net.src.ip in 192.168.1.0/24`,
			expected:   `net.src.ip in 10.53.51.54/24`,
		},
		{
			name:       "Sanitize expression with IPv6 address",
			expression: `net.src.ip == 2001:db8::1`,
			expected:   `net.src.ip == fd00:3561:6333:3965:3331:6636:6135:3132`,
		},
		{
			name:       "Sanitize expression with CIDR notation for IPv6",
			expression: `net.src.ip in 2001:db8::/32`,
			expected:   `net.src.ip in fd00:6331:6338:3530:3165:3561:3033:6434/32`,
		},
		{
			name:       "Sanitize expression with CIDR notation for IPv6",
			expression: `net.src.ip not in 2001:db8::/32`,
			expected:   `net.src.ip not in fd00:6331:6338:3530:3165:3561:3033:6434/32`,
		},
		{
			name:       "sanitize any() style of matching",
			expression: `any(http.headers.x_foo) ~ r#"bar\d"#`,
			expected:   `any(http.headers.x_foo) ~ r#"54307518fddf09e0ab66e80cbf576c6a8ddfb0fd72bd0ad4dc6658e323b08b3b"#`,
		},
		{
			name:       "sanitize case-insensitive style of matching",
			expression: `lower(http.path) == "/foo"`,
			expected:   `lower(http.path) == "001e5c9ded211e2e3abe7eb2a7d931cd32a157d643eab2e446a9a7e748fdc5f5"`,
		},
		{
			name:       "sanitize nested-function calls #1",
			expression: `any(lower(http.headers.x_foo)) ~ r#"bar\d"#`,
			expected:   `any(lower(http.headers.x_foo)) ~ r#"54307518fddf09e0ab66e80cbf576c6a8ddfb0fd72bd0ad4dc6658e323b08b3b"#`,
		},
		{
			name:       "sanitize nested-function calls #2",
			expression: `any(lower(upper(http.headers.x_foo))) ~ r#"bar\d"#`,
			expected: `any(lower(upper(http.headers.x_foo))) ~ ` +
				`r#"54307518fddf09e0ab66e80cbf576c6a8ddfb0fd72bd0ad4dc6658e323b08b3b"#`,
		},
		{
			name:       "sanitize nested-function calls #3",
			expression: `any(lower(upper(http.headers.x_foo))) ~ r#"bar\d"# && size(http.path) > 5)`,
			expected: `any(lower(upper(http.headers.x_foo))) ~ ` +
				`r#"54307518fddf09e0ab66e80cbf576c6a8ddfb0fd72bd0ad4dc6658e323b08b3b"#  &&  size(http.path) > 5 ) `,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.sanitizeExpression(tt.expression)
			assert.Equal(t, tt.expected, result)
		})
	}
}
