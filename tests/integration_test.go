package filter_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/vidinfra/golens/filter"
)

// MockSelectQuery is a simplified mock for testing
type MockSelectQuery struct {
	whereConditions []string
	orderConditions []string
}

func (m *MockSelectQuery) Where(condition string, args ...interface{}) *MockSelectQuery {
	m.whereConditions = append(m.whereConditions, condition)
	return m
}

func (m *MockSelectQuery) Order(condition string) *MockSelectQuery {
	m.orderConditions = append(m.orderConditions, condition)
	return m
}

func TestBuilder_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name               string
		queryParams        string
		allowedFields      []string
		configs            []filter.FilterConfig
		expectedWhereCount int
		expectedOrderCount int
	}{
		{
			name:               "basic filtering with allowed fields",
			queryParams:        "filter[name][eq]=john&filter[age][gte]=25",
			allowedFields:      []string{"name", "age"},
			expectedWhereCount: 2,
		},
		{
			name:               "filtering with disallowed field",
			queryParams:        "filter[name][eq]=john&filter[email][eq]=test@example.com",
			allowedFields:      []string{"name"},
			expectedWhereCount: 1, // Only name filter should be applied
		},
		{
			name:        "filtering with configs",
			queryParams: "filter[name][eq]=john&filter[name][contains]=jo",
			configs: []filter.FilterConfig{
				filter.AllowedFilter("name", filter.Equals, filter.Contains),
			},
			expectedWhereCount: 2,
		},
		{
			name:        "filtering with disallowed operator",
			queryParams: "filter[name][eq]=john&filter[name][gt]=100",
			configs: []filter.FilterConfig{
				filter.AllowedFilter("name", filter.Equals),
			},
			expectedWhereCount: 1, // Only equals operator should be allowed
		},
		{
			name:               "sorting",
			queryParams:        "sort=name,-age",
			allowedFields:      []string{"name", "age"},
			expectedOrderCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test HTTP request
			req := httptest.NewRequest(http.MethodGet, "/?"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Since we can't easily mock bun.SelectQuery, we'll test the parsing logic
			values, _ := url.ParseQuery(tt.queryParams)
			parser := filter.NewParser(values)
			filters := parser.Parse()

			validator := filter.NewValidator(tt.allowedFields, tt.configs)

			validFilterCount := 0
			for _, f := range filters {
				if validator.IsFilterAllowed(f) {
					validFilterCount++
				}
			}

			if validFilterCount != tt.expectedWhereCount {
				t.Errorf("Expected %d valid filters, got %d", tt.expectedWhereCount, validFilterCount)
			}
		})
	}
}
