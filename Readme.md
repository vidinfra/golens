# Go Filter Library

A flexible and powerful filtering library for Go applications, designed for seamless integration with popular frameworks like Gin and ORMs like Bun.

## Features

- **JSON API Compliant**: Supports `filter[field][operator]=value` format
- **Simple Format Support**: Fallback to `filter[field]=value` (assumes equals)
- **Multiple Operators**: Comprehensive set of filter operators (eq, like, gt, in, between, etc.)
- **Field Validation**: Configure which fields can be filtered and sorted
- **Operator Control**: Specify allowed operators per field
- **Sorting Support**: Built-in sorting with field validation
- **Fluent API**: Chain methods for clean, readable code
- **Type Safe**: Full Go type safety with structured error handling

## Installation

```bash
go get github.com/vidinfra/golens
```

## Quick Start

```go
package main

import (
    "log"
    "net/http"
    
    "github.com/gin-gonic/gin"
    "github.com/uptrace/bun"
    "github.com/vidinfra/golens/filter"
)

func main() {
    r := gin.Default()
    
    r.GET("/users", func(c *gin.Context) {
        query := db.NewSelect().Model((*User)(nil))
        
        // Apply filters and sorting with validation
        result := filter.New(c, query).
            AllowFields("name", "email", "age").
            AllowSorts("name", "created_at").
            Apply().
            ApplySort()
        
        // Handle validation errors
        if result.HasErrors() {
            c.JSON(http.StatusBadRequest, result.Result().ToJSONResponse())
            return
        }
        
        // Execute the validated query
        finalQuery := result.Query()
        var users []User
        err := finalQuery.Scan(c.Request.Context(), &users)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
            return
        }
        
        c.JSON(http.StatusOK, users)
    })
    
    r.Run()
}
```

## Error Handling

The library provides structured error reporting for validation failures.

### Basic Error Handling

```go
result := filter.New(c, query).
    AllowFields("name", "email").
    Apply()

// Check for validation errors
if result.HasErrors() {
    // Use built-in JSON response
    c.JSON(http.StatusBadRequest, result.Result().ToJSONResponse())
    return
}
```

### Custom Error Handling

```go
if result.HasErrors() {
    errors := result.GetErrors()
    
    // Access error details for custom handling
    for _, err := range errors.Errors {
        log.Printf("Validation error: %s (Field: %s)", err.Message, err.Field)
    }
    
    // Build custom response
    c.JSON(http.StatusBadRequest, gin.H{
        "error": "Invalid filters provided",
        "details": errors.Errors[0].Message,
    })
    return
}
```

### Error Response Format

Standard JSON error response format:

```json
{
  "success": false,
  "errors": [
    {
      "type": "validation_error",
      "code": "FILTER_VALIDATION_ERROR", 
      "field": "email",
      "message": "Field 'email' is not allowed for filtering",
      "suggestions": ["name", "age"]
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

## Filter Operators

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

## API Reference

### Builder Methods

- `AllowFields(fields ...string)`: Set allowed fields for filtering
- `AllowSorts(fields ...string)`: Set allowed fields for sorting  
- `AllowAll(fields ...string)`: Set same fields for both filtering and sorting
- `AllowConfigs(configs ...FilterConfig)`: Use detailed field configurations
- `Apply()`: Parse and apply filters (returns Builder for chaining)
- `ApplySort()`: Apply sorting (returns Builder for chaining)

### Result Methods

- `HasErrors() bool`: Check if any validation errors occurred
- `GetErrors() *FilterErrors`: Get detailed error information
- `Query()`: Get the final validated Bun query
- `Result().ToJSONResponse()`: Convert to standard JSON error format

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
    c.JSON(http.StatusBadRequest, result.Result().ToJSONResponse())
    return
}

finalQuery := result.Query()
```

### Production Example

```go
func handleFilters(c *gin.Context, query *bun.SelectQuery) (*bun.SelectQuery, error) {
    result := filter.New(c, query).
        AllowFields("name", "email", "age", "status").
        AllowSorts("name", "created_at", "updated_at").
        Apply().
        ApplySort()
    
    if result.HasErrors() {
        errors := result.GetErrors()
        
        // Log errors for debugging
        for _, err := range errors.Errors {
            log.Printf("Filter validation error: %s (Field: %s)", err.Message, err.Field)
        }
        
        c.JSON(http.StatusBadRequest, result.Result().ToJSONResponse())
        return nil, fmt.Errorf("filter validation failed")
    }
    
    return result.Query(), nil
}
```

## Examples

The `/examples` directory contains comprehensive usage examples:

- **`examples/basic/`**: Basic filtering and sorting
- **`examples/custom-error-handling/`**: Advanced error handling patterns
- **`examples/gin-integration/`**: Production-ready Gin + Bun ORM integration

### Running Examples

```bash
cd examples/basic && go run main.go
cd examples/custom-error-handling && go run main.go  
cd examples/gin-integration && go run main.go
```

## Testing

```bash
# Run all tests
make test

# Run with race detection  
make test-race

# Generate coverage report
make test-cover
```

## Development

```bash
# Install development tools
make install-tools

# Run linting and formatting
make dev

# Run all checks before release
make release
```

## Migration Guide

### From v1.x to v2.x

**v1.x (panic-based):**
```go
query := filter.New(c, query).
    AllowFields("name").
    Apply().
    Query() // Could panic on validation errors
```

**v2.x (error handling):**
```go
result := filter.New(c, query).
    AllowFields("name").
    Apply()

if result.HasErrors() {
    c.JSON(http.StatusBadRequest, result.Result().ToJSONResponse())
    return
}

query := result.Query() // Safe
```

## License

MIT License - see LICENSE file for details.
