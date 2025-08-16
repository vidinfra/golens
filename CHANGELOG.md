# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

---

## [v0.2.1] - 2025-08-16
### Fixed
- Fixed empty branch handling (`fix: empty branch`)

---

## [v0.2.0] - 2025-08-16
### Added
- Switched ORM integration from Bun to GORM
- Implemented full operator support with validation
- Added structured error handling with `Result` type
- Introduced `OK()` method for cleaner checks
- Added test coverage for all operators using `stretchr/testify`

### Changed
- Refactored `Builder` API (removed `HasErrors()`, replaced with `OK()`)
- Improved error messages and JSON response format
- Updated example app to use SQLite in-memory DB

### Removed
- Removed Bun-specific code
- Removed `HasErrors()` method

---

## [v0.1.0] - 2025-08-10
### Added
- Initial release with Bun ORM integration
- Basic filtering and sorting
- JSON API and simple filter formats
- Error handling with validation
