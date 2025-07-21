# Go Filter Library

A flexible and powerful filtering library for Go applications, particularly designed for use with Gin and Bun ORM. Features struct-first error handling with optional JSON conversion helpers for maximum flexibility and developer experience.

## Features

- **JSON API Compliant**: Supports `filter[field][operator]=value` format
- **Simple Format Support**: Fallback to `filter[field]=value` (assumes equals)
- **Multiple Operators**: Comprehensive set of filter operators (eq, like, gt, in, etc.)
- **Field Validation**: Allow/deny specific fields for filtering
- **Operator Validation**: Configure which operators are allowed per field
- **Sorting Support**: Built-in sorting with field validation
- **Fluent API**: Chain methods for clean, readable code
- **Type Safe**: Full Go type safety with custom types
- **Structured Error Handling**: Rich error types for validation and processing failures
- **JSON Helpers**: Optional JSON conversion for API responses
- **Comprehensive Testing**: Table-driven tests with JSON validation

## Installation

```bash
go get github.com/vidinfra/golens
```

## Quick Start

```go
package main

import (
    "encoding/json"
    "net/http"
    
    "github.com/gin-gonic/gin"
    "github.com/uptrace/bun"
    "github.com/vidinfra/golens/filter"
)

func main() {
    r := gin.Default()
    
    r.GET("/users", func(c *gin.Context) {
        query := db.NewSelect().Model((*User)(nil))
        
        result := filter.New(c, query).
            AllowFields("name", "email", "age").
            AllowSorts("name", "created_at").
            Apply().
            ApplySort()
            
        // Struct-first approach with error handling
        if result.HasErrors() {
            // Option 1: Use built-in JSON helper
            c.JSON(http.StatusBadRequest, result.ToJSONResponse())
            return
        }
        
        // Apply the validated filters to get the final query
        finalQuery := result.Query()
        var users []User
        err := finalQuery.Scan(c.Request.Context(), &users)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
            return
        }
        
        c.JSON(http.StatusOK, users)
    })
}
```

## Error Handling

The library follows a struct-first approach, returning structured errors that can be easily processed or converted to JSON.

### Error Types

- **`FilterError`**: Single validation or processing error
- **`FilterErrors`**: Collection of multiple errors with aggregation methods

### Basic Error Handling

```go
result := filter.New(c, query).
    AllowFields("name", "email").
    Apply()

// Check for errors using the struct API
if result.HasErrors() {
    errors := result.Errors
    
    // Access individual errors
    for _, err := range errors.Items {
        fmt.Printf("Field: %s, Code: %s, Message: %s\n", 
            err.Field, err.Code, err.Message)
    }
    
    // Use built-in JSON conversion
    c.JSON(http.StatusBadRequest, result.ToJSONResponse())
    return
}
```

### Custom Error Handling

```go
// Custom error processing for internationalization
if result.HasErrors() {
    customErrors := make(map[string]string)
    
    for _, err := range result.Errors.Items {
        // Translate error messages based on error code
        switch err.Code {
        case "INVALID_FIELD":
            customErrors[err.Field] = translator.Translate("invalid_field", err.Field)
        case "INVALID_OPERATOR":
            customErrors[err.Field] = translator.Translate("invalid_operator", err.Details["operator"])
        }
    }
    
    c.JSON(http.StatusBadRequest, gin.H{"errors": customErrors})
    return
}
```

### JSON Response Format

The `ToJSONResponse()` method returns a standardized format:

```json
{
  "success": false,
  "errors": [
    {
      "field": "invalid_field",
      "code": "INVALID_FIELD", 
      "message": "Field 'invalid_field' is not allowed for filtering",
      "details": {}
    },
    {
      "field": "age",
      "code": "INVALID_OPERATOR",
      "message": "Operator 'invalid_op' is not allowed for field 'age'",
      "details": {
        "operator": "invalid_op",
        "allowed_operators": ["eq", "gt", "lt"]
      }
    }
  ]
}
```

## URL Examples

### JSON API Format
```
GET /users?filter[name][eq]=john&filter[age][gte]=25&sort=name,-created_at
```

### Simple Format
```
GET /users?filter[status]=active&filter[name]=john
```

### Advanced Filtering
```
GET /products?filter[price][between]=10,100&filter[category][in]=electronics,books
```

## API Reference

### Builder Methods

- `AllowFields(fields ...string)`: Set allowed fields for filtering
- `AllowSorts(fields ...string)`: Set allowed fields for sorting  
- `AllowAll(fields ...string)`: Set same fields for both filtering and sorting
- `AllowConfigs(configs ...FilterConfig)`: Use detailed field configurations
- `Apply()`: Parse and apply filters to query (returns Result)
- `ApplySort()`: Apply sorting to query (returns Result)
- `Query()`: Get the final query (only if no errors)

### Result Methods

- `HasErrors() bool`: Check if any validation errors occurred
- `ToJSONResponse() map[string]interface{}`: Convert to JSON-ready response
- `Query()`: Get the final Bun query (panics if errors exist - check HasErrors first)

### Filter Operators

- `eq` - Equals
- `ne` - Not equals  
- `like` - Contains (case insensitive)
- `not-like` - Does not contain
- `starts-with` - Starts with
- `ends-with` - Ends with
- `gt` - Greater than
- `gte` - Greater than or equal
- `lt` - Less than
- `lte` - Less than or equal
- `in` - In list (comma-separated)
- `not-in` - Not in list
- `null` - Is null
- `not-null` - Is not null
- `between` - Between two values (comma-separated)
- `not-between` - Not between two values

## Advanced Configuration

### Field-Specific Operator Control

```go
configs := []filter.FilterConfig{
    filter.AllowedFilter("name", filter.Equals, filter.Contains, filter.StartsWith),
    filter.AllowedFilter("price", filter.GreaterThan, filter.LessThan, filter.Between),
    filter.AllowedFilter("category", filter.Equals, filter.In),
}

result := filter.New(c, query).
    AllowConfigs(configs...).
    Apply()

if result.HasErrors() {
    c.JSON(http.StatusBadRequest, result.ToJSONResponse())
    return
}

finalQuery := result.Query()
```

### Complete Error Handling Example

```go
func handleFilters(c *gin.Context, query *bun.SelectQuery) (*bun.SelectQuery, error) {
    result := filter.New(c, query).
        AllowFields("name", "email", "age", "status").
        AllowSorts("name", "created_at", "updated_at").
        Apply().
        ApplySort()
    
    if result.HasErrors() {
        // Log errors for debugging
        for _, err := range result.Errors.Items {
            log.Printf("Filter validation error - Field: %s, Code: %s, Message: %s", 
                err.Field, err.Code, err.Message)
        }
        
        // Return structured JSON response
        c.JSON(http.StatusBadRequest, result.ToJSONResponse())
        return nil, errors.New("validation failed")
    }
    
    return result.Query(), nil
}
```

### Custom Error Translation

```go
type ErrorTranslator struct {
    locale string
}

func (et *ErrorTranslator) TranslateErrors(errors *filter.FilterErrors) map[string]interface{} {
    translated := make(map[string]interface{})
    translatedErrors := make([]map[string]interface{}, 0, len(errors.Items))
    
    for _, err := range errors.Items {
        translatedErr := map[string]interface{}{
            "field": err.Field,
            "code":  err.Code,
        }
        
        switch err.Code {
        case "INVALID_FIELD":
            translatedErr["message"] = et.translate("validation.invalid_field", err.Field)
        case "INVALID_OPERATOR":
            translatedErr["message"] = et.translate("validation.invalid_operator", 
                err.Details["operator"], err.Field)
        default:
            translatedErr["message"] = err.Message
        }
        
        if len(err.Details) > 0 {
            translatedErr["details"] = err.Details
        }
        
        translatedErrors = append(translatedErrors, translatedErr)
    }
    
    return map[string]interface{}{
        "success": false,
        "errors":  translatedErrors,
    }
}
```

## Testing

The library includes comprehensive tests with both struct and JSON validation patterns.

### Running Tests

```bash
# Run all tests
make test

# Run with race detection  
make test-race

# Generate coverage report
make test-cover

# Run specific test files
go test ./tests/result_test.go
go test ./tests/result_json_test.go
```

### Test Structure

The library uses table-driven tests for comprehensive coverage:

#### Unit Tests (`tests/result_test.go`)
Tests the struct API and error handling:
```go
func TestResult_HasErrors(t *testing.T) {
    tests := []struct {
        name     string
        setup    func() *filter.Result
        expected bool
    }{
        {
            name: "no errors",
            setup: func() *filter.Result {
                return &filter.Result{}
            },
            expected: false,
        },
        // ... more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := tt.setup()
            got := result.HasErrors()
            assert.Equal(t, tt.expected, got)
        })
    }
}
```

#### JSON Tests (`tests/result_json_test.go`)
Tests JSON output, structure, and validity:
```go
func TestResult_ToJSONResponse_Structure(t *testing.T) {
    tests := []struct {
        name     string
        setup    func() *filter.Result
        validate func(t *testing.T, response map[string]interface{})
    }{
        {
            name: "success response structure",
            setup: func() *filter.Result {
                return &filter.Result{}
            },
            validate: func(t *testing.T, response map[string]interface{}) {
                assert.Contains(t, response, "success")
                assert.Equal(t, true, response["success"])
                assert.NotContains(t, response, "errors")
            },
        },
        // ... more test cases
    }
}
```

### Writing New Tests

1. **Struct Tests**: Focus on behavior, error conditions, and edge cases
2. **JSON Tests**: Validate output format, structure, and marshaling
3. **Use anonymous functions**: For complex test setup that needs isolation
4. **Table-driven approach**: For comprehensive coverage of multiple scenarios

Example test addition:
```go
func TestNewFeature(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        setup    func() *filter.Result
        expected bool
        wantErr  bool
    }{
        // Test cases here
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := tt.setup()
            // Test logic here
        })
    }
}
```

## Development

```bash
# Install development tools
make install-tools

# Run linting and formatting
make dev

# Run all checks before release
make release

# Run tests with verbose output
go test -v ./tests/

# Run tests for specific functionality
go test -v ./tests/ -run TestResult_HasErrors
go test -v ./tests/ -run TestResult_ToJSONResponse
```

### Contributing

1. **Error Handling**: All new features should use the struct-first error pattern
2. **Testing**: Add both struct and JSON tests for new functionality
3. **Documentation**: Update README examples when adding new features
4. **Validation**: Ensure comprehensive validation with meaningful error messages

### Architecture

- **`filter/`**: Core library code
  - `errors.go`: Structured error types and JSON helpers
  - `result.go`: Result struct and aggregation methods
  - `builder.go`: Fluent API builder pattern
  - `validator.go`: Field and operator validation
  - `applier.go`: Query application logic
  - Other supporting files...

- **`tests/`**: Test files
  - `result_test.go`: Table-driven unit tests for struct API
  - `result_json_test.go`: JSON output and structure validation tests

## Migration Guide

### From v1.x to v2.x

The library now returns a `Result` struct instead of panicking on errors:

**Old (v1.x):**
```go
query := filter.New(c, query).
    AllowFields("name").
    Apply().
    Query() // Could panic on validation errors
```

**New (v2.x):**
```go
result := filter.New(c, query).
    AllowFields("name").
    Apply()

if result.HasErrors() {
    c.JSON(http.StatusBadRequest, result.ToJSONResponse())
    return
}

query := result.Query() // Safe - errors already checked
```

## License

MIT License - see LICENSE file for details.
