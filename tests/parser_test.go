package filter_test

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/vidinfra/golens/filter"
)

func TestParser_Parse(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected []filter.Filter
	}{
		{
			name:  "single JSON API filter",
			query: "filter[name][eq]=john",
			expected: []filter.Filter{
				{Field: "name", Operator: filter.Equals, Value: "john"},
			},
		},
		{
			name:  "multiple JSON API filters",
			query: "filter[name][eq]=john&filter[age][gte]=25",
			expected: []filter.Filter{
				{Field: "name", Operator: filter.Equals, Value: "john"},
				{Field: "age", Operator: filter.GreaterThanOrEq, Value: "25"},
			},
		},
		{
			name:  "simple format filter",
			query: "filter[status]=active",
			expected: []filter.Filter{
				{Field: "status", Operator: filter.Equals, Value: "active"},
			},
		},
		{
			name:  "mixed format filters",
			query: "filter[name][eq]=john&filter[status]=active",
			expected: []filter.Filter{
				{Field: "name", Operator: filter.Equals, Value: "john"},
				{Field: "status", Operator: filter.Equals, Value: "active"},
			},
		},
		{
			name:     "no filters",
			query:    "sort=name&limit=10",
			expected: []filter.Filter{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, _ := url.ParseQuery(tt.query)
			parser := filter.NewParser(values)

			result := parser.Parse()
			filters := result.Filters

			if len(filters) != len(tt.expected) {
				t.Fatalf("Expected %d filters, got %d", len(tt.expected), len(filters))
			}

			// Convert to maps for order-independent comparison
			expectedMap := make(map[string]filter.Filter)
			for _, f := range tt.expected {
				key := f.Field + ":" + string(f.Operator)
				expectedMap[key] = f
			}

			actualMap := make(map[string]filter.Filter)
			for _, f := range filters {
				key := f.Field + ":" + string(f.Operator)
				actualMap[key] = f
			}

			if !reflect.DeepEqual(actualMap, expectedMap) {
				t.Errorf("Filters mismatch.\nExpected: %+v\nActual: %+v", tt.expected, filters)
			}
		})
	}
}

func TestParseFilterKey(t *testing.T) {
	// This tests the internal parseFilterKey function
	// You'll need to expose it or make it a method of Parser for testing
}
