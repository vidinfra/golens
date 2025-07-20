# Go Filter Library

A flexible and powerful filtering library for Go applications, particularly designed for use with Gin and Bun ORM.

## Features

- **JSON API Compliant**: Supports `filter[field][operator]=value` format
- **Simple Format Support**: Fallback to `filter[field]=value` (assumes equals)
- **Multiple Operators**: Comprehensive set of filter operators (eq, like, gt, in, etc.)
- **Field Validation**: Allow/deny specific fields for filtering
- **Operator Validation**: Configure which operators are allowed per field
- **Sorting Support**: Built-in sorting with field validation
- **Fluent API**: Chain methods for clean, readable code
- **Type Safe**: Full Go type safety with custom types

## Installation

```bash
go get github.com/vidinfra/golens
```

## Quick Start

```go
package main

import (
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
            ApplySort().
            Query()
            
        var users []User
        err := result.Scan(c.Request.Context(), &users)
        // Handle result...
    })
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
- `Apply()`: Parse and apply filters to query
- `ApplySort()`: Apply sorting to query
- `Query()`: Get the final query

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

```go
configs := []filter.FilterConfig{
    filter.AllowedFilter("name", filter.Equals, filter.Contains, filter.StartsWith),
    filter.AllowedFilter("price", filter.GreaterThan, filter.LessThan, filter.Between),
    filter.AllowedFilter("category", filter.Equals, filter.In),
}

result := filter.New(c, query).
    AllowConfigs(configs...).
    Apply().
    Query()
```

## Testing

Run tests:
```bash
make test
```

Run with race detection:
```bash
make test-race  
```

Generate coverage report:
```bash
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

## License

MIT License - see LICENSE file for details.
