package sanitize

import (
	"net"
	"regexp"
	"strconv"
	"strings"
)

// sanitizeExpression sanitizes a route expression by preserving the structure
// This ensures that expressions remain functionally intact but sensitive data is protected.
func (s *Sanitizer) sanitizeExpression(expression string) string {
	// Reference route expressions documentation:
	// https://developer.konghq.com/gateway/routing/expressions/#expressions-router-reference
	// A route has 3 main components:
	// 1. Fields: These are the identifiers like http.path, net.protocol, etc
	// 2. Operators: These are the comparison operators like ==, !=, >, etc.
	// 3. Values/Constant Values: These are the constant values like strings, numbers, IP addresses, etc.
	// These components create predicates that can be combined with logical operators (&&, ||, etc).
	// We need to sanitize an expression while preserving its structure.
	// We will only sanitize values, not fields or operators.

	// First, handling parentheses and logical operators to maintain expression structure
	parenRegex := regexp.MustCompile(`(\(|\))`)
	logicalOpRegex := regexp.MustCompile(`(&&|\|\|)`)
	// Adding spaces around parentheses and logical operators for easier parsing
	expression = parenRegex.ReplaceAllString(expression, " $1 ")
	expression = logicalOpRegex.ReplaceAllString(expression, " $1 ")

	// Second, breaking into predicate segments (separated by operators)
	segments := strings.Split(expression, " ")
	for i := 0; i < len(segments); i++ {
		segment := segments[i]

		// Skip logical operators and parentheses
		if segment == "&&" || segment == "||" || segment == "(" || segment == ")" || segment == "" {
			continue
		}

		// Handle "not in" operator (special case with space)
		if segment == "not" && i+1 < len(segments) && segments[i+1] == "in" {
			continue
		}

		// Check if this segment is part of a predicate (field-operator-value)
		if i+2 < len(segments) && isField(segment) {
			op := segments[i+1]

			// Check if the segment is a recognized operator
			if isOperator(op) {
				// value = next segment after the operator
				value := segments[i+2]

				if !shouldPreserveFieldValue(segment) {
					segments[i+2] = s.sanitizeConstantValue(value)
				}

				i += 2
			}
		}
	}

	// Reconstruct the expression
	return strings.Join(segments, " ")
}

// sanitizeConstantValue sanitizes a 'constant value' in an expression based on its type
func (s *Sanitizer) sanitizeConstantValue(value string) string {
	// Checking for string constants (surrounded by quotes)
	if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
		(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
		quote := string(value[0])
		// Extracting the actual string value without quotes
		strValue := value[1 : len(value)-1]

		// Checking for regex patterns (r#"pattern"#) inside strings
		if strings.HasPrefix(strValue, "r#\"") && strings.HasSuffix(strValue, "\"#") {
			// Preserving the regex structure and hashing the pattern inside
			pattern := strValue[3 : len(strValue)-2]
			hashedPattern := s.hashValue(pattern)
			return "r#\"" + hashedPattern + "\"#"
		}

		// Hashing the string content but preserve the quotes
		hashedValue := s.hashValue(strValue)
		return quote + hashedValue + quote
	}

	// Checking for regex patterns (r#"pattern"#)
	if strings.HasPrefix(value, "r#\"") && strings.HasSuffix(value, "\"#") {
		// Preserving the regex structure and hashing the pattern inside
		pattern := value[3 : len(value)-2]
		hashedPattern := s.hashValue(pattern)
		return "r#\"" + hashedPattern + "\"#"
	}

	// Handling IP addresses and CIDR notation
	ipv4CidrRegex := regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+(/\d+)?$`)
	ipv6CidrRegex := regexp.MustCompile(`^([0-9a-fA-F]{0,4}:){1,7}([0-9a-fA-F]{0,4}|:)(/\d+)?$|` +
		`^::(/\d+)?$|` +
		`^([0-9a-fA-F]{0,4}:){1,6}:([0-9a-fA-F]{0,4})(/\d+)?$`)

	if ipv4CidrRegex.MatchString(value) || ipv6CidrRegex.MatchString(value) {
		if strings.Contains(value, "/") {
			// For CIDR notation
			parts := strings.Split(value, "/")
			hashedIP := s.sanitizeIP(parts[0])
			return hashedIP + "/" + parts[1]
		}
		return s.sanitizeIP(value)
	}

	// Default handling for other types
	// We don't need to sanitize numeric or boolean constants
	return value
}

// sanitizeIP sanitizes an IPv4 or IPv6 address using the provided salt
func (s *Sanitizer) sanitizeIP(ip string) string {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		// If it's not a valid IP, return a hashed version
		return s.hashValue(ip)
	}

	hashedIP := s.hashValue(ip)

	if parsedIP.To4() != nil {
		return sanitizeIPv4(hashedIP)
	}

	return sanitizeIPv6(hashedIP)
}

// isField checks if a string is a valid field identifier in Kong expressions
func isField(s string) bool {
	// Fields are typically in the format: namespace.field or namespace.field.subfield
	fieldPattern := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*(\.[a-zA-Z][a-zA-Z0-9_]*)+$`)

	// Field prefixes in Kong Route expressions
	knownPrefixes := []string{"http.", "net.", "tls."}

	for _, prefix := range knownPrefixes {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}

	return fieldPattern.MatchString(s)
}

// isOperator checks if a string is a valid operator in route expressions
func isOperator(s string) bool {
	operators := []string{
		"==", "!=", ">=", "<=", ">", "<", "^=", "=^", "~",
		"in", "not in", "contains",
	}

	for _, op := range operators {
		if s == op {
			return true
		}
	}

	return false
}

// shouldPreserveFieldValue determines if a field's value should be preserved (not hashed)
// based on the field name. This allows keeping standard values for specific fields.
func shouldPreserveFieldValue(fieldName string) bool {
	// List of field names whose values should not be hashed
	preserveValueFields := []string{
		"http.method",
		"net.protocol",
		"http.protocol",
		"tls.protocol",
	}

	for _, field := range preserveValueFields {
		if field == fieldName {
			return true
		}
	}

	return false
}

// sanitizeIPv4 sanitizes an IPv4 address while preserving the IPv4 format
func sanitizeIPv4(hashedIP string) string {
	hashBytes := []byte(hashedIP)

	// Using the hash bytes as octets, ensuring they're in valid range
	// Using 10 for octet1 to ensure it's in the private IP space
	// to avoid conflicts with public IPs
	octet1 := 10 // Always use 10 for private IP space
	octet2 := int(hashBytes[0])
	octet3 := int(hashBytes[1])
	octet4 := int(hashBytes[2])

	return strconv.Itoa(octet1) + "." + strconv.Itoa(octet2) + "." + strconv.Itoa(octet3) + "." + strconv.Itoa(octet4)
}

// sanitizeIPv6 sanitizes an IPv6 address while preserving the IPv6 format
func sanitizeIPv6(hashedIP string) string {
	hashBytes := []byte(hashedIP)

	// Using fd00::/8 which is reserved for unique local addresses
	var ipv6Parts []string
	ipv6Parts = append(ipv6Parts, "fd00")

	// Generating 7 more 16-bit segments from the hash
	for i := 0; i < 7; i++ {
		if i*2+1 < len(hashBytes) {
			segment := (int(hashBytes[i*2]) << 8) | int(hashBytes[i*2+1])
			ipv6Parts = append(ipv6Parts, strconv.FormatInt(int64(segment), 16))
		} else {
			// If we run out of hash bytes, use a default pattern
			ipv6Parts = append(ipv6Parts, "0")
		}
	}

	return strings.Join(ipv6Parts, ":")
}
