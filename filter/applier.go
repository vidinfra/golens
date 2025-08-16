package filter

import (
	"fmt"
	"slices"
	"strings"

	"gorm.io/gorm"
)

// Applier handles applying filters and sort to database queries
type Applier struct {
	validator *Validator
}

func NewApplier(validator *Validator) *Applier {
	return &Applier{validator: validator}
}

func (a *Applier) validateFilter(filter Filter) *FilterError {
	if a.validator != nil {
		return a.validator.ValidateFilter(filter)
	}
	if !filter.Operator.IsValid() {
		return NewInvalidOperatorError(string(filter.Operator))
	}
	return nil
}

// Apply runs BOTH filters and sort in one shot.
// Public entrypoint: call this from Builder.Apply().
func (a *Applier) Apply(q *gorm.DB, filters []Filter, sortParam string, allowedSorts []string) (*Result, error) {
	// 1) Apply filters
	res, err := a.applyFilters(q, filters)
	if err != nil && res != nil && !res.OK() {
		return res, res.Errors
	}

	// 2) Apply sorting
	db, sortErrs := a.applySort(res.Query, sortParam, allowedSorts)
	if len(sortErrs) > 0 {
		for _, e := range sortErrs {
			res.AddError(e)
		}
	}

	res.Query = db
	if !res.OK() {
		return res, res.Errors
	}
	return res, nil
}

// --- private helpers ---

// applyFilters applies the provided filters in sequence.
func (a *Applier) applyFilters(q *gorm.DB, filters []Filter) (*Result, error) {
	result := NewResult(q)

	for _, f := range filters {
		if err := a.validateFilter(f); err != nil {
			result.AddError(err)
			continue
		}
		if a.validator != nil && !a.validator.IsFilterAllowed(f) {
			result.AddError(NewFieldNotAllowedError(f.Field, a.validator.allowedFields))
			continue
		}

		newQ, ferr := a.applyFilter(result.Query, f)
		if ferr != nil {
			result.AddError(ferr)
			continue
		}
		result.Query = newQ
	}

	if !result.OK() {
		return result, result.Errors
	}
	return result, nil
}

// applySort applies a comma-separated sort spec (e.g., "-created_at,name").
// Pass allowedSorts to restrict which columns can be sorted.
func (a *Applier) applySort(q *gorm.DB, sortParam string, allowedSorts []string) (*gorm.DB, []*FilterError) {
	var errs []*FilterError
	if sortParam == "" {
		return q, nil
	}

	for _, s := range strings.Split(sortParam, ",") {
		sortField := strings.TrimSpace(s)
		if sortField == "" {
			continue
		}

		desc := strings.HasPrefix(sortField, "-")
		if desc {
			sortField = sortField[1:]
		}

		if allowedSorts != nil && !slices.Contains(allowedSorts, sortField) {
			errs = append(errs, NewSortFieldNotAllowedError(sortField, allowedSorts))
			continue
		}

		if desc {
			q = q.Order(sortField + " DESC")
		} else {
			q = q.Order(sortField + " ASC")
		}
	}
	return q, errs
}

// applyFilter applies a single filter condition.
func (a *Applier) applyFilter(q *gorm.DB, filter Filter) (*gorm.DB, *FilterError) {
	field := filter.Field
	value := filter.Value

	switch filter.Operator {
	case Equals:
		q = q.Where(fmt.Sprintf("%s = ?", field), value)

	case NotEquals:
		q = q.Where(fmt.Sprintf("%s <> ?", field), value)

	case Contains:
		q = a.applyCaseInsensitiveLike(q, field, "%"+fmt.Sprintf("%v", value)+"%", false)

	case NotContains:
		q = a.applyCaseInsensitiveLike(q, field, "%"+fmt.Sprintf("%v", value)+"%", true)

	case StartsWith:
		q = a.applyCaseInsensitiveLike(q, field, fmt.Sprintf("%v", value)+"%", false)

	case EndsWith:
		q = a.applyCaseInsensitiveLike(q, field, "%"+fmt.Sprintf("%v", value), false)

	case GreaterThan:
		q = q.Where(fmt.Sprintf("%s > ?", field), value)

	case GreaterThanOrEq:
		q = q.Where(fmt.Sprintf("%s >= ?", field), value)

	case LessThan:
		q = q.Where(fmt.Sprintf("%s < ?", field), value)

	case LessThanOrEq:
		q = q.Where(fmt.Sprintf("%s <= ?", field), value)

	case In:
		values := parseCommaSeparatedValues(fmt.Sprintf("%v", value))
		q = q.Where(fmt.Sprintf("%s IN ?", field), values)

	case NotIn:
		values := parseCommaSeparatedValues(fmt.Sprintf("%v", value))
		q = q.Where(fmt.Sprintf("%s NOT IN ?", field), values)

	case IsNull:
		q = q.Where(fmt.Sprintf("%s IS NULL", field))

	case IsNotNull:
		q = q.Where(fmt.Sprintf("%s IS NOT NULL", field))

	case Between:
		values := parseCommaSeparatedValues(fmt.Sprintf("%v", value))
		if len(values) != 2 {
			return q, NewInvalidBetweenValueError(field, fmt.Sprintf("%v", value))
		}
		q = q.Where(fmt.Sprintf("%s BETWEEN ? AND ?", field), values[0], values[1])

	case NotBetween:
		values := parseCommaSeparatedValues(fmt.Sprintf("%v", value))
		if len(values) != 2 {
			return q, NewValidationError(
				field,
				string(NotBetween),
				fmt.Sprintf("%v", value),
				"Not between operator requires exactly two comma-separated values",
				"Use format: 'value1,value2' (e.g., '10,20')",
			)
		}
		q = q.Where(fmt.Sprintf("%s NOT BETWEEN ? AND ?", field), values[0], values[1])

	default:
		return q, NewInvalidOperatorError(string(filter.Operator))
	}

	return q, nil
}

func parseCommaSeparatedValues(value string) []string {
	if value == "" {
		return []string{}
	}
	parts := strings.Split(value, ",")
	result := make([]string, len(parts))
	for i, part := range parts {
		result[i] = strings.TrimSpace(part)
	}
	return result
}

// DatabaseDriver represents supported database drivers
type DatabaseDriver string

const (
	PostgreSQL DatabaseDriver = "postgresql"
	MySQL      DatabaseDriver = "mysql"
	SQLite     DatabaseDriver = "sqlite"
	Unknown    DatabaseDriver = "unknown"
)

func detectDatabaseDriver(db *gorm.DB) DatabaseDriver {
	if db == nil || db.Dialector == nil {
		return Unknown
	}
	name := strings.ToLower(db.Dialector.Name())
	switch {
	case strings.Contains(name, "postgres"):
		return PostgreSQL
	case strings.Contains(name, "mysql"):
		return MySQL
	case strings.Contains(name, "sqlite"):
		return SQLite
	default:
		return Unknown
	}
}

func (a *Applier) applyCaseInsensitiveLike(q *gorm.DB, field, pattern string, negate bool) *gorm.DB {
	switch detectDatabaseDriver(q) {
	case PostgreSQL:
		if negate {
			return q.Where(fmt.Sprintf("%s NOT ILIKE ?", field), pattern)
		}
		return q.Where(fmt.Sprintf("%s ILIKE ?", field), pattern)

	case MySQL:
		if negate {
			return q.Where(fmt.Sprintf("%s NOT LIKE ? COLLATE utf8mb4_general_ci", field), pattern)
		}
		return q.Where(fmt.Sprintf("%s LIKE ? COLLATE utf8mb4_general_ci", field), pattern)

	case SQLite:
		if negate {
			return q.Where(fmt.Sprintf("%s NOT LIKE ?", field), pattern)
		}
		return q.Where(fmt.Sprintf("%s LIKE ?", field), pattern)

	default:
		if negate {
			return q.Where(fmt.Sprintf("LOWER(%s) NOT LIKE LOWER(?)", field), pattern)
		}
		return q.Where(fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", field), pattern)
	}
}
