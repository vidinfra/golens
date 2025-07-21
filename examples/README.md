# Examples

This directory contains examples demonstrating the **struct-first error handling** approach of the Go Filter Library.

## Updated API Pattern

All examples have been updated to reflect the current API which emphasizes:

1. **Struct-First Error Handling**: Direct access to rich error structs
2. **Optional JSON Conversion**: Use built-in JSON helpers only when needed
3. **Flexible Error Processing**: Custom logic based on error types, fields, and codes

## Key Changes from Previous Versions

### Before (Old Pattern)
```go
builder := filter.New(c, query).
    AllowFields("name").
    Apply().
    Query() // Could panic or return incomplete data
```

### After (Current Pattern)
```go
result := filter.New(c, query).
    AllowFields("name").
    Apply()

// Direct struct access for full control
if result.HasErrors() {
    for _, err := range result.GetErrors().Errors {
        // Custom logic based on error details
        log.Printf("Error: %s (Field: %s, Code: %s)", err.Message, err.Field, err.Code)
    }
    
    // Optional: Use JSON helper for quick responses
    c.JSON(http.StatusBadRequest, result.Result().ToJSONResponse())
    return
}

finalQuery := result.Query()
```

## Examples Overview

### 1. `basic/main.go`
- Simple struct-first error handling
- Demonstrates both direct struct access and JSON helpers
- Basic field and sort validation

### 2. `custom-error-handling/main.go`
- Advanced error categorization by type
- Field-specific error handling logic
- Custom response formats
- Translation/internationalization example

### 3. `gin-integration/main.go`
- Real-world Gin + Bun ORM integration
- Comprehensive filtering with field configurations
- Production-ready error handling patterns
- Database error handling

## Key Benefits

1. **Performance**: No forced JSON serialization
2. **Flexibility**: Build any response format you need
3. **Debugging**: Rich error context for logging and troubleshooting
4. **Internationalization**: Error codes and context ready for translation
5. **Type Safety**: Full Go type safety throughout

## Error Structure

```go
type FilterError struct {
    Type        ErrorType `json:"type"`        // validation_error, parsing_error, etc.
    Message     string    `json:"message"`     // Human-readable message
    Field       string    `json:"field"`       // Field that caused the error
    Operator    string    `json:"operator"`    // Operator that caused the error
    Code        string    `json:"code"`        // Machine-readable code
    Suggestions []string  `json:"suggestions"` // Helpful suggestions
}
```

## Usage Patterns

### Direct Struct Access (Recommended)
```go
if result.HasErrors() {
    for _, err := range result.GetErrors().Errors {
        switch err.Code {
        case "FILTER_VALIDATION_ERROR":
            // Handle validation errors
        case "FILTER_PARSING_ERROR":
            // Handle parsing errors
        }
    }
}
```

### JSON Helper (When Needed)
```go
if result.HasErrors() {
    c.JSON(http.StatusBadRequest, result.Result().ToJSONResponse())
    return
}
```

### Custom Response Formats
```go
if result.HasErrors() {
    customResponse := map[string]interface{}{
        "status": "error",
        "message": result.GetErrors().Errors[0].Message,
        "suggestions": result.GetErrors().Errors[0].Suggestions,
    }
    c.JSON(http.StatusBadRequest, customResponse)
    return
}
```
