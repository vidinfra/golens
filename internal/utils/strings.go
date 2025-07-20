package utils

import "strings"

func ParseCommaSeparatedValues(value string) []string {
	if value == "" {
		return []string{}
	}

	parts := strings.Split(value, ",")
	result := make([]string, len(parts))

	for i, part := range parts {
		result[i] = strings.TrimSpace(part)
	}

	return result
}
