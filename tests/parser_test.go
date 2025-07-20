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

			filters := parser.Parse()

			if len(filters) != len(tt.expected) {
				t.Fatalf("Expected %d filters, got %d", len(tt.expected), len(filters))
			}

			for i, expected := range tt.expected {
				if !reflect.DeepEqual(filters[i], expected) {
					t.Errorf("Filter[%d] = %+v, want %+v", i, filters[i], expected)
				}
			}
		})
	}
}

func TestParseFilterKey(t *testing.T) {
	// This tests the internal parseFilterKey function
	// You'll need to expose it or make it a method of Parser for testing
}
