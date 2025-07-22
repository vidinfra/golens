package filter_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/vidinfra/golens/filter"
)

type User struct {
	ID    int    `bun:"id,pk"`
	Name  string `bun:"name"`
	Email string `bun:"email"`
	Age   int    `bun:"age"`
}

// Helper function to create test Gin context
func createTestContext(queryParams string) *gin.Context {
	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest(http.MethodGet, "/?"+queryParams, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	return c
}

// Helper function to create test Bun query
func createTestQuery() *bun.SelectQuery {
	db := bun.NewDB(nil, pgdialect.New())
	return db.NewSelect().Model((*User)(nil))
}

func TestBuilder_New(t *testing.T) {
	tests := []struct {
		name        string
		setupCtx    func() *gin.Context
		setupQuery  func() *bun.SelectQuery
		shouldPanic bool
	}{
		{
			name: "valid context and query",
			setupCtx: func() *gin.Context {
				return createTestContext("")
			},
			setupQuery: func() *bun.SelectQuery {
				return createTestQuery()
			},
			shouldPanic: false,
		},
		{
			name: "valid with query params",
			setupCtx: func() *gin.Context {
				return createTestContext("filter[name][eq]=john")
			},
			setupQuery: func() *bun.SelectQuery {
				return createTestQuery()
			},
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			query := tt.setupQuery()

			builder := filter.New(ctx, query)
			if builder == nil {
				t.Fatal("Builder should not be nil")
			}

			t.Logf("✅ Builder created successfully")
		})
	}
}

func TestBuilder_AllowFields(t *testing.T) {
	tests := []struct {
		name          string
		queryParams   string
		allowedFields []string
		expectError   bool
		errorField    string
	}{
		{
			name:          "allowed field passes",
			queryParams:   "filter[name][eq]=john",
			allowedFields: []string{"name", "email"},
			expectError:   false,
		},
		{
			name:          "disallowed field fails",
			queryParams:   "filter[age][eq]=25",
			allowedFields: []string{"name", "email"},
			expectError:   true,
			errorField:    "age",
		},
		{
			name:          "multiple allowed fields",
			queryParams:   "filter[name][eq]=john&filter[email][like]=test",
			allowedFields: []string{"name", "email", "age"},
			expectError:   false,
		},
		{
			name:          "empty allowed fields - all rejected",
			queryParams:   "filter[name][eq]=john",
			allowedFields: []string{},
			expectError:   true,
			errorField:    "name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext(tt.queryParams)
			query := createTestQuery()

			builder := filter.New(ctx, query).
				AllowFields(tt.allowedFields...).
				Apply()

			if tt.expectError {
				if !builder.HasErrors() {
					t.Errorf("Expected error for field '%s' but got none", tt.errorField)
					return
				}

				errors := builder.GetErrors()
				found := false
				for _, err := range errors.Errors {
					if err.Field == tt.errorField {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error for field '%s' but not found in errors", tt.errorField)
				}
			} else {
				if builder.HasErrors() {
					t.Errorf("Expected no errors but got: %v", builder.GetErrors())
				}
			}

			t.Logf("✅ Field validation test passed")
		})
	}
}

func TestBuilder_AllowSorts(t *testing.T) {
	tests := []struct {
		name         string
		queryParams  string
		allowedSorts []string
		expectError  bool
	}{
		{
			name:         "allowed sort field",
			queryParams:  "sort=name",
			allowedSorts: []string{"name", "created_at"},
			expectError:  false,
		},
		{
			name:         "disallowed sort field",
			queryParams:  "sort=age",
			allowedSorts: []string{"name", "created_at"},
			expectError:  true,
		},
		{
			name:         "multiple sort fields",
			queryParams:  "sort=name,-created_at",
			allowedSorts: []string{"name", "created_at"},
			expectError:  false,
		},
		{
			name:         "mixed allowed and disallowed sorts",
			queryParams:  "sort=name,age",
			allowedSorts: []string{"name", "created_at"},
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext(tt.queryParams)
			query := createTestQuery()

			builder := filter.New(ctx, query).
				AllowSorts(tt.allowedSorts...).
				Apply().
				ApplySort()

			if tt.expectError {
				if !builder.HasErrors() {
					t.Error("Expected error but got none")
				}
			} else {
				if builder.HasErrors() {
					t.Errorf("Expected no errors but got: %v", builder.GetErrors())
				}
			}

			t.Logf("✅ Sort validation test passed")
		})
	}
}

func TestBuilder_AllowAll(t *testing.T) {
	tests := []struct {
		name        string
		queryParams string
		allowedAll  []string
		expectError bool
	}{
		{
			name:        "filter and sort with same allowed fields",
			queryParams: "filter[name][eq]=john&sort=name",
			allowedAll:  []string{"name", "email"},
			expectError: false,
		},
		{
			name:        "disallowed filter field",
			queryParams: "filter[age][eq]=25&sort=name",
			allowedAll:  []string{"name", "email"},
			expectError: true,
		},
		{
			name:        "disallowed sort field",
			queryParams: "filter[name][eq]=john&sort=age",
			allowedAll:  []string{"name", "email"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext(tt.queryParams)
			query := createTestQuery()

			builder := filter.New(ctx, query).
				AllowAll(tt.allowedAll...).
				Apply().
				ApplySort()

			if tt.expectError {
				if !builder.HasErrors() {
					t.Error("Expected error but got none")
				}
			} else {
				if builder.HasErrors() {
					t.Errorf("Expected no errors but got: %v", builder.GetErrors())
				}
			}

			t.Logf("✅ AllowAll test passed")
		})
	}
}

func TestBuilder_AllowConfigs(t *testing.T) {
	tests := []struct {
		name        string
		queryParams string
		configs     []filter.FilterConfig
		expectError bool
		description string
	}{
		{
			name:        "allowed operator for field",
			queryParams: "filter[name][eq]=john",
			configs: []filter.FilterConfig{
				filter.AllowedFilter("name", filter.Equals, filter.Contains),
			},
			expectError: false,
			description: "Should allow equals operator for name field",
		},
		{
			name:        "disallowed operator for field",
			queryParams: "filter[name][gt]=5",
			configs: []filter.FilterConfig{
				filter.AllowedFilter("name", filter.Equals, filter.Contains),
			},
			expectError: true,
			description: "Should reject gt operator for name field",
		},
		{
			name:        "multiple field configs",
			queryParams: "filter[name][like]=john&filter[age][gte]=18",
			configs: []filter.FilterConfig{
				filter.AllowedFilter("name", filter.Equals, filter.Contains),
				filter.AllowedFilter("age", filter.GreaterThanOrEq, filter.LessThanOrEq),
			},
			expectError: false,
			description: "Should allow specific operators for each field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext(tt.queryParams)
			query := createTestQuery()

			builder := filter.New(ctx, query).
				AllowConfigs(tt.configs...).
				Apply()

			if tt.expectError {
				if !builder.HasErrors() {
					t.Errorf("Expected error but got none. %s", tt.description)
				}
			} else {
				if builder.HasErrors() {
					t.Errorf("Expected no errors but got: %v. %s", builder.GetErrors(), tt.description)
				}
			}

			t.Logf("✅ %s", tt.description)
		})
	}
}

func TestBuilder_MethodChaining(t *testing.T) {
	tests := []struct {
		name        string
		queryParams string
		setupChain  func(*filter.Builder) *filter.Builder
		expectError bool
	}{
		{
			name:        "chained AllowFields and Apply",
			queryParams: "filter[name][eq]=john",
			setupChain: func(b *filter.Builder) *filter.Builder {
				return b.AllowFields("name").Apply()
			},
			expectError: false,
		},
		{
			name:        "chained AllowFields, AllowSorts, Apply, ApplySort",
			queryParams: "filter[name][eq]=john&sort=name",
			setupChain: func(b *filter.Builder) *filter.Builder {
				return b.AllowFields("name").AllowSorts("name").Apply().ApplySort()
			},
			expectError: false,
		},
		{
			name:        "chained AllowAll with Apply and ApplySort",
			queryParams: "filter[email][like]=test&sort=-email",
			setupChain: func(b *filter.Builder) *filter.Builder {
				return b.AllowAll("email").Apply().ApplySort()
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext(tt.queryParams)
			query := createTestQuery()

			builder := filter.New(ctx, query)
			result := tt.setupChain(builder)

			if tt.expectError {
				if !result.HasErrors() {
					t.Error("Expected error but got none")
				}
			} else {
				if result.HasErrors() {
					t.Errorf("Expected no errors but got: %v", result.GetErrors())
				}
			}

			// Verify we can still access the final query
			finalQuery := result.Query()
			if finalQuery == nil {
				t.Error("Final query should not be nil")
			}

			t.Logf("✅ Method chaining test passed")
		})
	}
}

func TestBuilder_ComplexQueries(t *testing.T) {
	tests := []struct {
		name            string
		queryParams     string
		setupBuilder    func(*filter.Builder) *filter.Builder
		expectError     bool
		expectedFilters int
		description     string
	}{
		{
			name:        "multiple filters with different operators",
			queryParams: "filter[name][like]=john&filter[age][gte]=18&filter[email][ne]=test",
			setupBuilder: func(b *filter.Builder) *filter.Builder {
				return b.AllowFields("name", "age", "email").Apply()
			},
			expectError:     false,
			expectedFilters: 3,
			description:     "Should handle multiple different filter operators",
		},
		{
			name:        "filters with sorting",
			queryParams: "filter[name][eq]=john&filter[age][lt]=50&sort=name,-age",
			setupBuilder: func(b *filter.Builder) *filter.Builder {
				return b.AllowAll("name", "age").Apply().ApplySort()
			},
			expectError:     false,
			expectedFilters: 2,
			description:     "Should handle both filtering and sorting",
		},
		{
			name:        "between operator",
			queryParams: "filter[age][between]=18,65",
			setupBuilder: func(b *filter.Builder) *filter.Builder {
				return b.AllowFields("age").Apply()
			},
			expectError:     false,
			expectedFilters: 1,
			description:     "Should handle between operator with two values",
		},
		{
			name:        "in operator",
			queryParams: "filter[status][in]=active,pending,draft",
			setupBuilder: func(b *filter.Builder) *filter.Builder {
				return b.AllowFields("status").Apply()
			},
			expectError:     false,
			expectedFilters: 1,
			description:     "Should handle in operator with multiple values",
		},
		{
			name:        "null checks",
			queryParams: "filter[deleted_at][null]=true&filter[email][not-null]=true",
			setupBuilder: func(b *filter.Builder) *filter.Builder {
				return b.AllowFields("deleted_at", "email").Apply()
			},
			expectError:     false,
			expectedFilters: 2,
			description:     "Should handle null and not-null operators",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext(tt.queryParams)
			query := createTestQuery()

			builder := filter.New(ctx, query)
			result := tt.setupBuilder(builder)

			if tt.expectError {
				if !result.HasErrors() {
					t.Errorf("Expected error but got none. %s", tt.description)
				}
			} else {
				if result.HasErrors() {
					t.Errorf("Expected no errors but got: %v. %s", result.GetErrors(), tt.description)
				}

				// Verify final query exists
				if result.Query() == nil {
					t.Error("Final query should not be nil")
				}
			}

			t.Logf("✅ %s", tt.description)
		})
	}
}

func TestBuilder_ErrorHandling(t *testing.T) {
	tests := []struct {
		name         string
		queryParams  string
		setupBuilder func(*filter.Builder) *filter.Builder
		expectError  bool
		errorType    string
	}{
		{
			name:        "invalid operator",
			queryParams: "filter[name][xyz]=john",
			setupBuilder: func(b *filter.Builder) *filter.Builder {
				return b.AllowFields("name").Apply()
			},
			expectError: true,
			errorType:   "invalid operator",
		},
		{
			name:        "invalid between format",
			queryParams: "filter[age][between]=18",
			setupBuilder: func(b *filter.Builder) *filter.Builder {
				return b.AllowFields("age").Apply()
			},
			expectError: true,
			errorType:   "invalid between format",
		},
		{
			name:        "field not allowed",
			queryParams: "filter[restricted][eq]=value",
			setupBuilder: func(b *filter.Builder) *filter.Builder {
				return b.AllowFields("name", "email").Apply()
			},
			expectError: true,
			errorType:   "field not allowed",
		},
		{
			name:        "multiple errors accumulate",
			queryParams: "filter[restricted][eq]=value&filter[name][xyz]=john",
			setupBuilder: func(b *filter.Builder) *filter.Builder {
				return b.AllowFields("name", "email").Apply()
			},
			expectError: true,
			errorType:   "multiple errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext(tt.queryParams)
			query := createTestQuery()

			builder := filter.New(ctx, query)
			result := tt.setupBuilder(builder)

			if tt.expectError {
				if !result.HasErrors() {
					t.Errorf("Expected %s error but got none", tt.errorType)
					return
				}

				// Verify error details are accessible
				errors := result.GetErrors()
				if len(errors.Errors) == 0 {
					t.Error("Expected errors but found none")
				}

				// Test JSON response generation
				jsonResponse := result.Result().ToJSONResponse()
				if jsonResponse == nil {
					t.Error("JSON response should not be nil")
				}
			} else {
				if result.HasErrors() {
					t.Errorf("Expected no errors but got: %v", result.GetErrors())
				}
			}

			t.Logf("✅ Error handling test for %s passed", tt.errorType)
		})
	}
}
