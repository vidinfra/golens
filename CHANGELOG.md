# Changelog

## [v0.2.0] - 2025-08-16

### ✨ Features
- **Switched ORM**: Migrated from **Bun ORM** to **GORM** for broader adoption.
- **Struct-first Error Handling**
  - No more panics — errors are returned as structured objects.
  - `Result.OK()` replaces `HasErrors()` for clarity.
  - Built-in JSON response helpers (`ToJSONResponse`) for quick API integration.
- **Comprehensive Operators**
  - Full support for operators: `eq`, `ne`, `like`, `not-like`, `starts-with`, `ends-with`, `gt`, `gte`, `lt`, `lte`, `in`, `not-in`, `null`, `not-null`, `between`, `not-between`.
- **Sorting Improvements**
  - Field validation for `sort` params.
  - Support for ascending/descending syntax (`sort=-created_at,name`).
- **Fluent Builder API**
  - `AllowFields`, `AllowSorts`, `AllowAll`, `AllowConfigs`.
  - Cleaner chaining with `Apply()`.

### 🛠 Improvements
- Optimized parsing logic (`Parser`) for JSON API + simple filter formats.
- Unified **Apply** pipeline: filters and sorting in one pass.
- Case-insensitive `LIKE` support across **Postgres, MySQL, SQLite** with dialect detection.
- Added helper functions for cleaner validation (`Validator`).

### 🧪 Testing
- Migrated to **stretchr/testify** for testing.
- Added full test coverage for every operator.
- Improved test readability and error assertions.

### 📚 Documentation
- Updated `README.md` with:
  - New **Quick Start** using **GORM**.
  - Examples with **SQLite in-memory DB**.
  - Error handling patterns (`struct-first`).
  - Migration guide from `v1.x` → `v2.x`.

---

⚡️ **Upgrade Note**:
`HasErrors()` has been removed. Use:

```go
if !result.OK() {
    c.JSON(http.StatusBadRequest, result.Result().ToJSONResponse())
    return
}
```
