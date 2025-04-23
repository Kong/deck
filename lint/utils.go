package lint

import "sigs.k8s.io/yaml"

func ParseSeverity(s string) Severity {
	for i, str := range severityStrings {
		if s == str {
			return Severity(i)
		}
	}
	return SeverityWarn
}

func isOpenAPISpec(fileBytes []byte) bool {
	var contents map[string]interface{}

	// This marshalling is redundant with what happens
	// in the linting command. There is likely an algorithm
	// we could use to determine JSON vs YAML and pull out the
	// openapi key without unmarshalling the entire file.
	err := yaml.Unmarshal(fileBytes, &contents)
	if err != nil {
		return false
	}

	return contents["openapi"] != nil
}
