package filter_test

import (
	"testing"

	"github.com/vidinfra/golens/filter"
)

func TestValidator_IsFilterAllowed(t *testing.T) {
	tests := []struct {
		filter        filter.Filter
		name          string
		allowedFields []string
		configs       []filter.FilterConfig
		expected      bool
	}{
		{
			name:          "allowed field without configs",
			allowedFields: []string{"name", "age"},
			configs:       nil,
			filter:        filter.Filter{Field: "name", Operator: filter.Equals, Value: "john"},
			expected:      true,
		},
		{
			name:          "disallowed field without configs",
			allowedFields: []string{"name", "age"},
			configs:       nil,
			filter:        filter.Filter{Field: "email", Operator: filter.Equals, Value: "test@example.com"},
			expected:      false,
		},
		{
			name:          "allowed field and operator with configs",
			allowedFields: nil,
			configs: []filter.FilterConfig{
				filter.AllowedFilter("name", filter.Equals, filter.Contains),
			},
			filter:   filter.Filter{Field: "name", Operator: filter.Equals, Value: "john"},
			expected: true,
		},
		{
			name:          "allowed field but disallowed operator with configs",
			allowedFields: nil,
			configs: []filter.FilterConfig{
				filter.AllowedFilter("name", filter.Equals),
			},
			filter:   filter.Filter{Field: "name", Operator: filter.Contains, Value: "john"},
			expected: false,
		},
		{
			name:          "no restrictions",
			allowedFields: nil,
			configs:       nil,
			filter:        filter.Filter{Field: "anything", Operator: filter.Equals, Value: "value"},
			expected:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := filter.NewValidator(tt.allowedFields, tt.configs)

			if got := validator.IsFilterAllowed(tt.filter); got != tt.expected {
				t.Errorf("Validator.IsFilterAllowed() = %v, want %v", got, tt.expected)
			}
		})
	}
}
