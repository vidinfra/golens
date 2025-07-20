package filter_test

import (
	"reflect"
	"testing"

	"github.com/vidinfra/golens/filter"
)

func TestAllowedFilter(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		operators []filter.Clause
		expected  filter.FilterConfig
	}{
		{
			name:      "single operator",
			field:     "name",
			operators: []filter.Clause{filter.Equals},
			expected: filter.FilterConfig{
				Field:            "name",
				AllowedOperators: []filter.Clause{filter.Equals},
				DefaultOperator:  filter.Equals,
				Description:      "Filter by name",
			},
		},
		{
			name:      "multiple operators",
			field:     "age",
			operators: []filter.Clause{filter.GreaterThan, filter.LessThan, filter.Equals},
			expected: filter.FilterConfig{
				Field:            "age",
				AllowedOperators: []filter.Clause{filter.GreaterThan, filter.LessThan, filter.Equals},
				DefaultOperator:  filter.GreaterThan,
				Description:      "Filter by age",
			},
		},
		{
			name:      "no operators (should default to equals)",
			field:     "status",
			operators: []filter.Clause{},
			expected: filter.FilterConfig{
				Field:            "status",
				AllowedOperators: []filter.Clause{filter.Equals},
				DefaultOperator:  filter.Equals,
				Description:      "Filter by status",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.AllowedFilter(tt.field, tt.operators...)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("AllowedFilter() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}
