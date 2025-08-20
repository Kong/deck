package sanitize

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
)

var (
	regexPrefix    = "r#\""
	regexSuffix    = "\"#"
	regexPrefixLen = len(regexPrefix)
	regexSuffixLen = len(regexSuffix)
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

	// First, preserving function calls, regex patterns and quoted string by replacing them
	// with placeholders, so that spacing doesn't affect parentheses and logical operators
	// inside the regex, quotes or func calls.

	// preserving function calls
	var functionCalls map[string]string
	var protectionOrder []string
	functionCalls, protectionOrder, expression = preserveFunctionCall(expression)

	protectedValues := make(map[string]string)
	placeholderCounter := 0
	// protecting regex patterns
	regexPattern := regexp.MustCompile(`r#"[^"]*"#`)
	expression = regexPattern.ReplaceAllStringFunc(expression, func(match string) string {
		placeholder := fmt.Sprintf("__PROTECTED_REGEX_%d__", placeholderCounter)
		protectedValues[placeholder] = match
		placeholderCounter++
		return placeholder
	})
	// protecting quoted strings ("..." or '...')
	quotedPattern := regexp.MustCompile(`"[^"]*"|'[^']*'`)
	expression = quotedPattern.ReplaceAllStringFunc(expression, func(match string) string {
		placeholder := fmt.Sprintf("__PROTECTED_STRING_%d__", placeholderCounter)
		protectedValues[placeholder] = match
		placeholderCounter++
		return placeholder
	})

	// Second, handling parentheses and logical operators to maintain expression structure
	parenRegex := regexp.MustCompile(`(\(|\))`)
	logicalOpRegex := regexp.MustCompile(`(&&|\|\|)`)
	// Adding spaces around parentheses and logical operators for easier parsing
	expression = parenRegex.ReplaceAllString(expression, " $1 ")
	expression = logicalOpRegex.ReplaceAllString(expression, " $1 ")

	// Third, breaking into predicate segments (separated by operators)
	segments := strings.Split(expression, " ")
	for i := 0; i < len(segments); i++ {
		segment := segments[i]

		// Skip logical operators and parentheses
		if segment == "&&" || segment == "||" || segment == "(" || segment == ")" || segment == "" {
			continue
		}

		// Check if this segment is part of a predicate (field-operator-value)
		if isField(segment) {
			var op string
			var valueIndex int

			// Handle "not in" operator (special case with space)
			if i+2 < len(segments) && segments[i+1] == "not" && segments[i+2] == "in" {
				op = "not in"
				valueIndex = i + 3
			} else if i+1 < len(segments) {
				op = segments[i+1]
				valueIndex = i + 2
			} else {
				continue
			}

			// Check if we have enough segments for the value and if the operator is recognized
			if valueIndex < len(segments) && isOperator(op) {
				// value = next segment after the operator
				value := segments[valueIndex]

				// Restore protected values before sanitizing
				if restoredValue, exists := protectedValues[value]; exists {
					value = restoredValue
				}

				if !shouldPreserveFieldValue(segment) {
					segments[valueIndex] = s.sanitizeConstantValue(value)
				} else {
					segments[valueIndex] = value
				}

				// Skip the appropriate number of segments based on operator type
				if op == "not in" {
					i += 3
				} else {
					i += 2
				}
			}
		}
	}

	// Reconstruct the expression
	result := strings.Join(segments, " ")
	// Restore original function calls
	for i := len(protectionOrder) - 1; i >= 0; i-- {
		placeholder := protectionOrder[i]
		if original, exists := functionCalls[placeholder]; exists {
			result = strings.ReplaceAll(result, placeholder, original)
		}
	}

	return result
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
		if strings.HasPrefix(strValue, regexPrefix) && strings.HasSuffix(strValue, regexSuffix) {
			return s.sanitizeRegex(strValue)
		}

		// Hashing the string content but preserve the quotes
		hashedValue := s.hashValue(strValue)
		return quote + hashedValue + quote
	}

	// Checking for regex patterns (r#"pattern"#)
	if strings.HasPrefix(value, regexPrefix) && strings.HasSuffix(value, regexSuffix) {
		return s.sanitizeRegex(value)
	}

	if ip, _, err := net.ParseCIDR(value); err == nil && ip != nil {
		// For CIDR notation, sanitize the IP part but keep the prefix length
		parts := strings.Split(value, "/")
		hashedIP := s.sanitizeIP(ip, parts[0])
		return hashedIP + "/" + parts[1]
	}

	if ip := net.ParseIP(value); ip != nil {
		// For single IP addresses, just sanitize the IP
		return s.sanitizeIP(ip, value)
	}

	// Default handling for other types
	// We don't need to sanitize numeric or boolean constants
	return value
}

// sanitizeIP sanitizes an IPv4 or IPv6 address using the provided salt
func (s *Sanitizer) sanitizeIP(ip net.IP, ipString string) string {
	hashedIP := s.hashValue(ipString)

	if ip.To4() != nil {
		return sanitizeIPv4(hashedIP)
	}

	return sanitizeIPv6(hashedIP)
}

// isField checks if a string is a valid field identifier in Kong expressions
func isField(s string) bool {
	// Check for protected function placeholders
	if strings.HasPrefix(s, "__PROTECTED_FUNCTION_") {
		return true
	}

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
		"==", "!=", ">=", "<=", ">", "<", "^=", "=^", "~", "!",
		"in", "not in", "contains",
	}

	for _, op := range operators {
		if s == op {
			return true
		}
	}

	return false
}

func preserveFunctionCall(expression string) (map[string]string, []string, string) {
	functionCalls := make(map[string]string)
	var protectionOrder []string // Track the order of protection
	functionsCounter := 0

	for {
		oldExpression := expression
		functionCallPattern := regexp.MustCompile(`\b(all|any|lower|upper|size|starts_with|ends_with)\s*\([^()]*\)`)

		expression = functionCallPattern.ReplaceAllStringFunc(expression, func(match string) string {
			placeholder := fmt.Sprintf("__PROTECTED_FUNCTION_%d__", functionsCounter)
			functionCalls[placeholder] = match
			protectionOrder = append(protectionOrder, placeholder) // Track order
			functionsCounter++
			return placeholder
		})

		if expression == oldExpression {
			break
		}
	}

	return functionCalls, protectionOrder, expression
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
		// Check if there are enough bytes left to form a 16-bit segment (2 bytes)
		hasEnoughBytesForSegment := i*2+1 < len(hashBytes)
		if hasEnoughBytesForSegment {
			segment := (int(hashBytes[i*2]) << 8) | int(hashBytes[i*2+1])
			ipv6Parts = append(ipv6Parts, strconv.FormatInt(int64(segment), 16))
		} else {
			// If we run out of hash bytes, use a default pattern
			ipv6Parts = append(ipv6Parts, "0")
		}
	}

	return strings.Join(ipv6Parts, ":")
}

// sanitizeRegex sanitizes a regex string by preserving its structure
func (s *Sanitizer) sanitizeRegex(regexString string) string {
	strLen := len(regexString)
	if strLen >= regexPrefixLen+regexSuffixLen {
		pattern := regexString[regexPrefixLen : strLen-regexSuffixLen]
		hashedPattern := s.hashValue(pattern)
		return regexPrefix + hashedPattern + regexSuffix
	}

	// fallback: hash the entire string if the lengths mismatch
	return s.hashValue(regexString)
}
