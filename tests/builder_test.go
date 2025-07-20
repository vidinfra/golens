package filter_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/vidinfra/golens/filter"
)

// Simple mock for bun.SelectQuery to capture SQL generation
type MockBunQuery struct {
	conditions []string
	orders     []string
}

// Add interface compatibility for bun.SelectQuery
func (m *MockBunQuery) Where(condition string, args ...interface{}) *MockBunQuery {
	// Simple string replacement for testing
	formatted := condition
	for _, arg := range args {
		formatted = strings.Replace(formatted, "?", "'"+arg.(string)+"'", 1)
	}
	m.conditions = append(m.conditions, formatted)
	return m
}

// Add compatibility: return interface{} so it matches bun.SelectQuery's method signatures
func (m *MockBunQuery) WhereInterface(condition string, args ...interface{}) interface{} {
	return m.Where(condition, args...)
}

func (m *MockBunQuery) Order(order string) *MockBunQuery {
	m.orders = append(m.orders, order)
	return m
}

// Add compatibility: return interface{} so it matches bun.SelectQuery's method signatures
func (m *MockBunQuery) OrderInterface(order string) interface{} {
	return m.Order(order)
}

// Generate SQL for comparison
func (m *MockBunQuery) ToSQL() string {
	sql := "SELECT * FROM users"

	if len(m.conditions) > 0 {
		sql += " WHERE " + strings.Join(m.conditions, " AND ")
	}

	if len(m.orders) > 0 {
		sql += " ORDER BY " + strings.Join(m.orders, ", ")
	}

	return sql
}

// Helper to create Gin context with query params
func createTestContext(queryParams string) *gin.Context {
	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest(http.MethodGet, "/?"+queryParams, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	return c
}

func TestBuilder_FilterGeneration(t *testing.T) {
	tests := []struct {
		name          string
		queryParams   string
		allowedFields []string
		configs       []filter.FilterConfig
		expectedSQL   string
		description   string
	}{
		{
			name:          "single equals filter",
			queryParams:   "filter[name][eq]=john",
			allowedFields: []string{"name"},
			expectedSQL:   "SELECT * FROM users WHERE \"name\" = 'john'",
			description:   "Basic equals filter",
		},
		{
			name:          "multiple filters",
			queryParams:   "filter[name][eq]=john&filter[age][gte]=25",
			allowedFields: []string{"name", "age"},
			expectedSQL:   "SELECT * FROM users WHERE \"name\" = 'john' AND \"age\" >= '25'",
			description:   "Multiple filters with AND",
		},
		{
			name:          "like operator",
			queryParams:   "filter[name][like]=jo",
			allowedFields: []string{"name"},
			expectedSQL:   "SELECT * FROM users WHERE \"name\" ILIKE '%jo%'",
			description:   "LIKE operator with wildcards",
		},
		{
			name:          "starts-with operator",
			queryParams:   "filter[name][starts-with]=john",
			allowedFields: []string{"name"},
			expectedSQL:   "SELECT * FROM users WHERE \"name\" ILIKE 'john%'",
			description:   "Starts-with operator",
		},
		{
			name:          "ends-with operator",
			queryParams:   "filter[email][ends-with]=.com",
			allowedFields: []string{"email"},
			expectedSQL:   "SELECT * FROM users WHERE \"email\" ILIKE '%.com'",
			description:   "Ends-with operator",
		},
		{
			name:          "in operator",
			queryParams:   "filter[status][in]=active,pending",
			allowedFields: []string{"status"},
			expectedSQL:   "SELECT * FROM users WHERE \"status\" IN ('active','pending')",
			description:   "IN operator with multiple values",
		},
		{
			name:          "between operator",
			queryParams:   "filter[age][between]=25,65",
			allowedFields: []string{"age"},
			expectedSQL:   "SELECT * FROM users WHERE \"age\" BETWEEN '25' AND '65'",
			description:   "BETWEEN operator",
		},
		{
			name:          "null check",
			queryParams:   "filter[deleted_at][null]=",
			allowedFields: []string{"deleted_at"},
			expectedSQL:   "SELECT * FROM users WHERE \"deleted_at\" IS NULL",
			description:   "NULL check",
		},
		{
			name:          "not null check",
			queryParams:   "filter[email][not-null]=",
			allowedFields: []string{"email"},
			expectedSQL:   "SELECT * FROM users WHERE \"email\" IS NOT NULL",
			description:   "NOT NULL check",
		},
		{
			name:          "not equals",
			queryParams:   "filter[status][ne]=inactive",
			allowedFields: []string{"status"},
			expectedSQL:   "SELECT * FROM users WHERE \"status\" != 'inactive'",
			description:   "Not equals operator",
		},
		{
			name:          "not like",
			queryParams:   "filter[name][not-like]=test",
			allowedFields: []string{"name"},
			expectedSQL:   "SELECT * FROM users WHERE \"name\" NOT ILIKE '%test%'",
			description:   "NOT LIKE operator",
		},
		{
			name:          "not in",
			queryParams:   "filter[role][not-in]=admin,super",
			allowedFields: []string{"role"},
			expectedSQL:   "SELECT * FROM users WHERE \"role\" NOT IN ('admin','super')",
			description:   "NOT IN operator",
		},
		{
			name:          "not between",
			queryParams:   "filter[age][not-between]=10,20",
			allowedFields: []string{"age"},
			expectedSQL:   "SELECT * FROM users WHERE \"age\" NOT BETWEEN '10' AND '20'",
			description:   "NOT BETWEEN operator",
		},
		{
			name:          "simple format",
			queryParams:   "filter[name]=john",
			allowedFields: []string{"name"},
			expectedSQL:   "SELECT * FROM users WHERE \"name\" = 'john'",
			description:   "Simple format defaults to equals",
		},
		{
			name:          "mixed formats",
			queryParams:   "filter[name][eq]=john&filter[status]=active",
			allowedFields: []string{"name", "status"},
			expectedSQL:   "SELECT * FROM users WHERE \"name\" = 'john' AND \"status\" = 'active'",
			description:   "Mixed JSON API and simple format",
		},
		{
			name:          "disallowed field filtered",
			queryParams:   "filter[name][eq]=john&filter[password][eq]=secret",
			allowedFields: []string{"name"}, // password not allowed
			expectedSQL:   "SELECT * FROM users WHERE \"name\" = 'john'",
			description:   "Disallowed fields should be ignored",
		},
		{
			name:        "config validation",
			queryParams: "filter[name][eq]=john&filter[name][gt]=test", // gt not allowed for name
			configs: []filter.FilterConfig{
				filter.AllowedFilter("name", filter.Equals, filter.Contains),
			},
			expectedSQL: "SELECT * FROM users WHERE \"name\" = 'john'",
			description: "Invalid operators should be filtered out",
		},
		{
			name:          "no filters",
			queryParams:   "sort=name&limit=10",
			allowedFields: []string{"name"},
			expectedSQL:   "SELECT * FROM users",
			description:   "No filter params should return base query",
		},
		{
			name:          "special characters",
			queryParams:   "filter[email][eq]=test@example.com",
			allowedFields: []string{"email"},
			expectedSQL:   "SELECT * FROM users WHERE \"email\" = 'test@example.com'",
			description:   "Handle special characters in values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test context
			c := createTestContext(tt.queryParams)
			fmt.Println(c)

			// Create mock query
			mockQuery := &MockBunQuery{}

			// Test your actual Builder
			var builder *filter.Builder
			if len(tt.configs) > 0 {
				// builder = filter.New(c, mockQuery).AllowConfigs(tt.configs...)
				fmt.Println("if on tt.configs")
			} else {
				// builder = filter.New(c, mockQuery).AllowFields(tt.allowedFields...)
				fmt.Println("else on tt.configs")
			}

			// Apply filters
			builder.Apply()

			// Check generated SQL
			actualSQL := mockQuery.ToSQL()
			if actualSQL != tt.expectedSQL {
				t.Errorf("\n%s\nExpected: %s\nActual:   %s",
					tt.description, tt.expectedSQL, actualSQL)
			}
		})
	}
}

func TestBuilder_SortGeneration(t *testing.T) {
	tests := []struct {
		name         string
		queryParams  string
		allowedSorts []string
		expectedSQL  string
		description  string
	}{
		{
			name:         "single ascending sort",
			queryParams:  "sort=name",
			allowedSorts: []string{"name"},
			expectedSQL:  "SELECT * FROM users ORDER BY name ASC",
			description:  "Single field ascending",
		},
		{
			name:         "single descending sort",
			queryParams:  "sort=-name",
			allowedSorts: []string{"name"},
			expectedSQL:  "SELECT * FROM users ORDER BY name DESC",
			description:  "Single field descending",
		},
		{
			name:         "multiple sorts",
			queryParams:  "sort=name,-age",
			allowedSorts: []string{"name", "age"},
			expectedSQL:  "SELECT * FROM users ORDER BY name ASC, age DESC",
			description:  "Multiple fields mixed order",
		},
		{
			name:         "disallowed sort field",
			queryParams:  "sort=name,password", // password not allowed
			allowedSorts: []string{"name"},
			expectedSQL:  "SELECT * FROM users ORDER BY name ASC",
			description:  "Disallowed sort fields filtered out",
		},
		{
			name:         "empty sort",
			queryParams:  "sort=",
			allowedSorts: []string{"name"},
			expectedSQL:  "SELECT * FROM users",
			description:  "Empty sort param",
		},
		{
			name:         "whitespace in sort",
			queryParams:  "sort= name , -age ",
			allowedSorts: []string{"name", "age"},
			expectedSQL:  "SELECT * FROM users ORDER BY name ASC, age DESC",
			description:  "Handle whitespace",
		},
		{
			name:         "no sort param",
			queryParams:  "filter[name][eq]=john",
			allowedSorts: []string{"name"},
			expectedSQL:  "SELECT * FROM users",
			description:  "No sort parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := createTestContext(tt.queryParams)
			mockQuery := &MockBunQuery{}

			fmt.Println(c)

			// builder := filter.New(c, mockQuery).AllowSorts(tt.allowedSorts...)
			// builder.ApplySort()

			actualSQL := mockQuery.ToSQL()
			if actualSQL != tt.expectedSQL {
				t.Errorf("\n%s\nExpected: %s\nActual:   %s",
					tt.description, tt.expectedSQL, actualSQL)
			}
		})
	}
}

func TestBuilder_CombinedFilterAndSort(t *testing.T) {
	tests := []struct {
		name          string
		queryParams   string
		allowedFields []string
		allowedSorts  []string
		expectedSQL   string
		description   string
	}{
		{
			name:          "filter and sort",
			queryParams:   "filter[name][eq]=john&sort=age",
			allowedFields: []string{"name"},
			allowedSorts:  []string{"age"},
			expectedSQL:   "SELECT * FROM users WHERE \"name\" = 'john' ORDER BY age ASC",
			description:   "Combine filtering and sorting",
		},
		{
			name:          "multiple filters and sorts",
			queryParams:   "filter[name][like]=john&filter[status]=active&sort=name,-created_at",
			allowedFields: []string{"name", "status"},
			allowedSorts:  []string{"name", "created_at"},
			expectedSQL:   "SELECT * FROM users WHERE \"name\" ILIKE '%john%' AND \"status\" = 'active' ORDER BY name ASC, created_at DESC",
			description:   "Complex filtering with sorting",
		},
		{
			name:          "only filter",
			queryParams:   "filter[name][eq]=john",
			allowedFields: []string{"name"},
			allowedSorts:  []string{"age"},
			expectedSQL:   "SELECT * FROM users WHERE \"name\" = 'john'",
			description:   "Only filtering, no sorting",
		},
		{
			name:          "only sort",
			queryParams:   "sort=name",
			allowedFields: []string{"name"},
			allowedSorts:  []string{"name"},
			expectedSQL:   "SELECT * FROM users ORDER BY name ASC",
			description:   "Only sorting, no filtering",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := createTestContext(tt.queryParams)
			mockQuery := &MockBunQuery{}
			fmt.Println(c)

			// builder := filter.New(c, mockQuery).
			// 	AllowFields(tt.allowedFields...).
			// 	AllowSorts(tt.allowedSorts...)

			// builder.Apply().ApplySort()

			actualSQL := mockQuery.ToSQL()
			if actualSQL != tt.expectedSQL {
				t.Errorf("\n%s\nExpected: %s\nActual:   %s",
					tt.description, tt.expectedSQL, actualSQL)
			}
		})
	}
}

func TestBuilder_AllowAll(t *testing.T) {
	c := createTestContext("filter[name][eq]=john&sort=-name")
	mockQuery := &MockBunQuery{}
	fmt.Println(c)

	// Test AllowAll method
	// builder := filter.New(c, mockQuery).AllowAll("name", "age")
	// builder.Apply().ApplySort()

	expectedSQL := "SELECT * FROM users WHERE \"name\" = 'john' ORDER BY name DESC"
	actualSQL := mockQuery.ToSQL()

	if actualSQL != expectedSQL {
		t.Errorf("AllowAll test failed\nExpected: %s\nActual:   %s", expectedSQL, actualSQL)
	}
}

// Test edge cases
func TestBuilder_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		queryParams string
		setupFunc   func(*filter.Builder) *filter.Builder
		expectedSQL string
	}{
		{
			name:        "malformed filter param",
			queryParams: "filter[name=john", // Missing closing bracket
			setupFunc: func(b *filter.Builder) *filter.Builder {
				return b.AllowFields("name")
			},
			expectedSQL: "SELECT * FROM users",
		},
		{
			name:        "empty filter value",
			queryParams: "filter[name][eq]=",
			setupFunc: func(b *filter.Builder) *filter.Builder {
				return b.AllowFields("name")
			},
			expectedSQL: "SELECT * FROM users WHERE \"name\" = ''",
		},
		{
			name:        "invalid operator",
			queryParams: "filter[name][invalid]=john",
			setupFunc: func(b *filter.Builder) *filter.Builder {
				return b.AllowFields("name")
			},
			expectedSQL: "SELECT * FROM users", // Invalid operator should be ignored
		},
		{
			name:        "between with single value",
			queryParams: "filter[age][between]=25", // Missing second value
			setupFunc: func(b *filter.Builder) *filter.Builder {
				return b.AllowFields("age")
			},
			expectedSQL: "SELECT * FROM users", // Invalid between should be ignored
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := createTestContext(tt.queryParams)
			mockQuery := &MockBunQuery{}
			fmt.Println(c)

			// builder := filter.New(c, mockQuery)
			// builder = tt.setupFunc(builder)
			// builder.Apply()

			actualSQL := mockQuery.ToSQL()
			if actualSQL != tt.expectedSQL {
				t.Errorf("%s\nExpected: %s\nActual:   %s",
					tt.name, tt.expectedSQL, actualSQL)
			}
		})
	}
}

// Benchmark tests
func BenchmarkBuilder_SimpleFilter(b *testing.B) {
	// c := createTestContext("filter[name][eq]=john")

	// b.ResetTimer()
	// for i := 0; i < b.N; i++ {
	// 	mockQuery := &MockBunQuery{}
	// 	filter.New(c, mockQuery).AllowFields("name").Apply()
	// }
}

func BenchmarkBuilder_ComplexFilter(b *testing.B) {
	// 	c := createTestContext("filter[name][like]=john&filter[age][gte]=25&filter[status][in]=active,pending&sort=name,-age")

	// b.ResetTimer()
	//
	//	for i := 0; i < b.N; i++ {
	//		mockQuery := &MockBunQuery{}
	//		filter.New(c, mockQuery).
	//			AllowFields("name", "age", "status").
	//			AllowSorts("name", "age").
	//			Apply().
	//			ApplySort()
	//	}
}
