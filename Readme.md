# Go Filter Library

A flexible and powerful filtering library for Go applications, designed for
seamless integration with frameworks like **Gin** and ORMs like **GORM**.

## Features

- **JSON API Compliant**: Supports `filter[field][operator]=value`
- **Simple Format Support**: Fallback to `filter[field]=value` (assumes equals)
- **Rich Operators**: `eq`, `like`, `gt`, `in`, `between`, `null`, etc.
- **Field Validation**: Configure which fields can be filtered and sorted
- **Operator Control**: Per-field operator allowlists
- **Sorting Support**: Built-in, with allowlist validation
- **Fluent API**: Chain methods for clean, readable code
- **Struct-First Error Handling**: Typed errors, no panics, JSON helpers for responses

## Installation

```bash
go get github.com/vidinfra/golens
```

## Quick Start

```go
package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"

    "github.com/vidinfra/golens/filter"
)

type User struct {
    ID     int    `json:"id" gorm:"primaryKey;autoIncrement"`
    Name   string `json:"name"`
    Email  string `json:"email"`
    Status string `json:"status"`
    Age    int    `json:"age"`
}

func main() {
    db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    db.AutoMigrate(&User{})

    r := gin.Default()
    r.GET("/users", func(c *gin.Context) {
        base := db.Model(&User{})

        builder := filter.New(c, base).
            AllowFields("name", "email", "age", "status").
            AllowSorts("name", "age").
            Apply()

        res := builder.Result()

        if !res.OK() {
            c.JSON(http.StatusBadRequest, res.ToJSONResponse())
            return
        }

        var users []User
        if err := res.Query.Find(&users).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"success": false,
            "message": "query failed"})
            return
        }

        c.JSON(http.StatusOK, gin.H{"success": true, "data": users})
    })

    r.Run(":8080")
}
```

## Error Handling

Errors are returned as **structs**, with optional JSON helpers.

### Struct-First

```go
res := builder.Result()

if !res.OK() {
    for _, err := range res.Errors.Errors {
        log.Printf("Validation error: %s (Field: %s)", err.Message, err.Field)
    }
    c.JSON(http.StatusBadRequest, res.ToJSONResponse())
    return
}
```

### JSON Error Response Format

```json
{
  "success": false,
  "errors": [
    {
      "type": "validation_error",
      "code": "FIELD_NOT_ALLOWED",
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
GET /users?filter[name][eq]=john&filter[age][gte]=25&sort=-age
```

### Simple Format
```
GET /users?filter[status]=active&filter[name]=john
```

### Advanced
```
GET /products?filter[price][between]=10,100&filter[category][in]=electronics,books
```

## Supported Operators

- `eq`, `ne`
- `like`, `not-like`
- `starts-with`, `ends-with`
- `gt`, `gte`, `lt`, `lte`
- `in`, `not-in`
- `null`, `not-null`
- `between`, `not-between`

## API Reference

### Builder

- `AllowFields(fields ...string)`
- `AllowSorts(fields ...string)`
- `AllowAll(fields ...string)`
- `AllowConfigs(configs ...FilterConfig)`
- `Apply()` → parses + applies filters & sort

### Result

- `OK() bool` → true if no errors
- `Errors *FilterErrors` → structured errors
- `Query *gorm.DB` → final validated query
- `ToJSONResponse()` → standard JSON format

## Migration Guide

### From v1.x (Bun, panic-based) to v2.x (GORM, struct-first)

**v1.x**
```go
query := filter.New(c, query).
    AllowFields("name").
    Apply().
    Query() // could panic
```

**v2.x**
```go
builder := filter.New(c, db.Model(&User{})).
    AllowFields("name").
    Apply()

res := builder.Result()
if !res.OK() {
    c.JSON(http.StatusBadRequest, res.ToJSONResponse())
    return
}
query := res.Query
```

## Examples

See `/examples`:
- `basic/` → simple filtering & sorting
- `gin/` → Gin + GORM integration
- `custom-error/` → advanced error handling

## License

MIT License - see LICENSE file for details.
