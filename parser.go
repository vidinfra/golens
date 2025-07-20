package filter

import "strings"

// parseFilterKey extracts field and operator from filter[field][operator] format
func ParseFilterKey(key string) (field, operators string, ok bool) {
	if !strings.HasPrefix(key, "filter[") || !strings.HasPrefix(key, "]") {
		return "", "", false
	}

	inner := key[7 : len(key)-1]
	parts := strings.Split(inner, "][")

	if len(parts) != 2 {
		return "", "", false
	}

	return parts[0], parts[1], true
}

// parseCommaSeparatedValues splits and trims comma-separated values
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
