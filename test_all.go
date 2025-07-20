// File: pkg/filter/clause_test.go
package filter_test

import (
	"testing"

	"github.com/yourusername/gofilter/pkg/filter"
)

func TestClause_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		clause   filter.Clause
		expected bool
	}{
		{"valid equals", filter.Equals, true},
		{"valid contains", filter.Contains, true},
		{"valid greater than", filter.GreaterThan, true},
		{"valid in", filter.In, true},
		{"valid is null", filter.IsNull, true},
		{"invalid clause", filter.Clause("invalid"), false},
		{"empty clause", filter.Clause(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.clause.IsValid(); got != tt.expected {
				t.Errorf("Clause.IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestClause_String(t *testing.T) {
	tests := []struct {
		name     string
		clause   filter.Clause
		expected string
	}{
		{"equals", filter.Equals, "eq"},
		{"contains", filter.Contains, "like"},
		{"greater than", filter.GreaterThan, "gt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.clause.String(); got != tt.expected {
				t.Errorf("Clause.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// File: pkg/filter/parser_test.go
package filter_test

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/yourusername/gofilter/pkg/filter"
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

// File: pkg/filter/validator_test.go
package filter_test

import (
	"testing"

	"github.com/yourusername/gofilter/pkg/filter"
)

func TestValidator_IsFilterAllowed(t *testing.T) {
	tests := []struct {
		name          string
		allowedFields []string
		configs       []filter.FilterConfig
		filter        filter.Filter
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

// File: pkg/filter/config_test.go
package filter_test

import (
	"reflect"
	"testing"

	"github.com/yourusername/gofilter/pkg/filter"
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

// File: examples/basic/main.go
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"github.com/yourusername/gofilter/pkg/filter"
)

type User struct {
	ID     int    `json:"id" bun:"id,pk,autoincrement"`
	Name   string `json:"name" bun:"name"`
	Email  string `json:"email" bun:"email"`
	Age    int    `json:"age" bun:"age"`
	Status string `json:"status" bun:"status"`
}

func main() {
	r := gin.Default()

	r.GET("/users", func(c *gin.Context) {
		// Simulate a database query
		var query *bun.SelectQuery // This would be your actual bun query
		
		// Create filter with allowed fields
		result := filter.New(c, query).
			AllowFields("name", "email", "age", "status").
			AllowSorts("name", "age", "created_at").
			Apply().
			ApplySort().
			Query()

		// Execute your query here
		_ = result
		
		c.JSON(http.StatusOK, gin.H{"message": "Users filtered successfully"})
	})

	r.Run(":8080")
}

// File: examples/advanced/main.go
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"github.com/yourusername/gofilter/pkg/filter"
)

func main() {
	r := gin.Default()

	r.GET("/products", func(c *gin.Context) {
		// Simulate a database query
		var query *bun.SelectQuery // This would be your actual bun query
		
		// Create filter with detailed configurations
		configs := []filter.FilterConfig{
			filter.AllowedFilter("name", filter.Equals, filter.Contains, filter.StartsWith),
			filter.AllowedFilter("price", filter.GreaterThan, filter.LessThan, filter.Between, filter.Equals),
			filter.AllowedFilter("category", filter.Equals, filter.In),
			filter.AllowedFilter("status", filter.Equals),
			filter.AllowedFilter("created_at", filter.GreaterThan, filter.LessThan, filter.Between),
		}

		result := filter.New(c, query).
			AllowConfigs(configs...).
			AllowSorts("name", "price", "created_at").
			Apply().
			ApplySort().
			Query()

		// Execute your query here
		_ = result
		
		c.JSON(http.StatusOK, gin.H{"message": "Products filtered successfully"})
	})

	r.Run(":8080")
}

// File: examples/gin-integration/main.go
package main

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/yourusername/gofilter/pkg/filter"
)

type Product struct {
	ID          int     `json:"id" bun:"id,pk,autoincrement"`
	Name        string  `json:"name" bun:"name"`
	Description string  `json:"description" bun:"description"`
	Price       float64 `json:"price" bun:"price"`
	Category    string  `json:"category" bun:"category"`
	Status      string  `json:"status" bun:"status"`
	CreatedAt   string  `json:"created_at" bun:"created_at"`
}

func main() {
	// Setup database connection (example)
	dsn := "postgres://username:password@localhost:5432/database?sslmode=disable"
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())

	r := gin.Default()

	// Example endpoint with comprehensive filtering
	r.GET("/products", func(c *gin.Context) {
		// Start with base query
		query := db.NewSelect().Model((*Product)(nil))

		// Define filter configurations
		configs := []filter.FilterConfig{
			filter.AllowedFilter("name", filter.Equals, filter.Contains, filter.StartsWith, filter.EndsWith),
			filter.AllowedFilter("description", filter.Contains),
			filter.AllowedFilter("price", filter.Equals, filter.GreaterThan, filter.LessThan, filter.GreaterThanOrEq, filter.LessThanOrEq, filter.Between),
			filter.AllowedFilter("category", filter.Equals, filter.In, filter.NotIn),
			filter.AllowedFilter("status", filter.Equals, filter.In),
			filter.AllowedFilter("created_at", filter.GreaterThan, filter.LessThan, filter.Between),
		}

		// Apply filters and sorting
		finalQuery := filter.New(c, query).
			AllowConfigs(configs...).
			AllowSorts("name", "price", "created_at", "category").
			Apply().
			ApplySort().
			Query()

		// Execute the query
		var products []Product
		err := finalQuery.Scan(c.Request.Context(), &products)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  products,
			"count": len(products),
		})
	})

	// Example with simple field allowlist
	r.GET("/users", func(c *gin.Context) {
		query := db.NewSelect().Model((*User)(nil))

		finalQuery := filter.New(c, query).
			AllowAll("name", "email", "status", "created_at").
			Apply().
			ApplySort().
			Query()

		var users []User
		err := finalQuery.Scan(c.Request.Context(), &users)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
			return
		}

		c.JSON(http.StatusOK, users)
	})

	r.Run(":8080")
}

type User struct {
	ID        int    `json:"id" bun:"id,pk,autoincrement"`
	Name      string `json:"name" bun:"name"`
	Email     string `json:"email" bun:"email"`
	Status    string `json:"status" bun:"status"`
	CreatedAt string `json:"created_at" bun:"created_at"`
}

// File: pkg/filter/integration_test.go
package filter_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"github.com/yourusername/gofilter/pkg/filter"
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
		name                string
		queryParams         string
		allowedFields       []string
		configs             []filter.FilterConfig
		expectedWhereCount  int
		expectedOrderCount  int
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


