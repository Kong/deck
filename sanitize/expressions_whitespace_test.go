package sanitize

import (
	"strings"
	"testing"
)

func TestSanitizeExpressionWhitespace(t *testing.T) {
	s := &Sanitizer{salt: "test-salt-123"}
	for _, separator := range []string{"", " ", "  ", "\t", "\r\n"} {
		for _, operator := range []string{"==", "!=", "^=", "=^", "~"} {
			t.Run(operator+"/"+separator, func(t *testing.T) {
				input := "http.path" + separator + operator + separator + `"/api/users"`
				want := "http.path " + operator + " " + s.sanitizeConstantValue(`"/api/users"`)
				if got := s.sanitizeExpression(input); got != want {
					t.Fatalf("sanitizeExpression(%q) = %q, want %q", input, got, want)
				}
			})
		}
	}
	if got, want := s.sanitizeExpression(`http.path=="/api/ users"`), "http.path == "+s.sanitizeConstantValue(`"/api/ users"`); got != want {
		t.Fatalf("literal whitespace changed: got %q, want %q", got, want)
	}
	for _, input := range []string{
		`http.method=="GET"`,
		`lower(http.path)=="/api/users"`,
		`http.path~r#"^/api/ users$"#`,
		"net.src.ip\t not\t in\t192.168.1.0/24",
		`(http.path=="/api/users")&&http.method=="GET"`,
	} {
		t.Run(input, func(t *testing.T) {
			got := s.sanitizeExpression(input)
			if strings.Contains(got, "__PROTECTED_") || strings.Contains(got, "/api/") || strings.Contains(got, "192.168.1.0") {
				t.Fatalf("unrestored or unsanitized value in %q", got)
			}
			if strings.Contains(input, "GET") && !strings.Contains(got, `"GET"`) {
				t.Fatalf("method not preserved in %q", got)
			}
		})
	}
}
