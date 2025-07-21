package filter

import (
	"fmt"
	"slices"
	"strings"

	"github.com/uptrace/bun"
)

// Applier handles applying filters to database queries
type Applier struct {
	validator *Validator
}

func NewApplier(validator *Validator) *Applier {
	return &Applier{
		validator: validator,
	}
}

func (a *Applier) validateFilter(filter Filter) *FilterError {
	if a.validator != nil {
		return a.validator.ValidateFilter(filter)
	}

	// Basic validation if no validator is set
	if !filter.Operator.IsValid() {
		return NewInvalidOperatorError(string(filter.Operator))
	}

	return nil
}

func (a *Applier) ApplyFilters(q *bun.SelectQuery, filters []Filter) (*Result, error) {
	result := NewResult(q)

	for _, filter := range filters {
		if err := a.validateFilter(filter); err != nil {
			result.AddError(err)
			continue
		}

		if a.validator != nil && !a.validator.IsFilterAllowed(filter) {
			result.AddError((NewFieldNotAllowedError(filter.Field, a.validator.allowedFields)))
			continue
		}

		newQuery, err := a.applyFilter(result.Query, filter)
		if err != nil {
			result.AddError(err)
			continue
		}
		result.Query = newQuery
	}

	if result.HasErrors() {
		return result, result.Errors
	}
	return result, nil
}

func (a *Applier) ApplySort(q *bun.SelectQuery, sortParam string, allowedSorts []string) (*bun.SelectQuery, []*FilterError) {
	var errors []*FilterError

	if sortParam == "" {
		return q, nil
	}

	for s := range strings.SplitSeq(sortParam, ",") {
		sortField := strings.TrimSpace(s)
		if sortField == "" {
			continue
		}

		// Check for descending order (-)
		desc := strings.HasPrefix(sortField, "-")
		if desc {
			sortField = sortField[1:]
		}

		if allowedSorts != nil && !slices.Contains(allowedSorts, sortField) {
			errors = append(errors, NewSortFieldNotAllowedError(sortField, allowedSorts))
			continue
		}

		if desc {
			q = q.Order(sortField + " DESC")
		} else {
			q = q.Order(sortField + " ASC")
		}
	}

	return q, errors
}

func (a *Applier) applyFilter(q *bun.SelectQuery, filter Filter) (*bun.SelectQuery, *FilterError) {
	field := filter.Field
	value := filter.Value

	switch filter.Operator {
	case Equals:
		q = q.Where("? = ?", bun.Ident(field), value)
	case NotEquals:
		q = q.Where("? != ?", bun.Ident(field), value)
	case Contains:
		q = q.Where("? ILIKE ?", bun.Ident(field), "%"+fmt.Sprintf("%v", value)+"%")
	case NotContains:
		q = q.Where("? NOT ILIKE ?", bun.Ident(field), "%"+fmt.Sprintf("%v", value)+"%")
	case StartsWith:
		q = q.Where("? ILIKE ?", bun.Ident(field), fmt.Sprintf("%v", value)+"%")
	case EndsWith:
		q = q.Where("? ILIKE ?", bun.Ident(field), "%"+fmt.Sprintf("%v", value))
	case GreaterThan:
		q = q.Where("? > ?", bun.Ident(field), value)
	case GreaterThanOrEq:
		q = q.Where("? >= ?", bun.Ident(field), value)
	case LessThan:
		q = q.Where("? < ?", bun.Ident(field), value)
	case LessThanOrEq:
		q = q.Where("? <= ?", bun.Ident(field), value)
	case In:
		values := parseCommaSeparatedValues(fmt.Sprintf("%v", value))
		q = q.Where("? IN (?)", bun.Ident(field), bun.In(values))
	case NotIn:
		values := parseCommaSeparatedValues(fmt.Sprintf("%v", value))
		q = q.Where("? NOT IN (?)", bun.Ident(field), bun.In(values))
	case IsNull:
		q = q.Where("? IS NULL", bun.Ident(field))
	case IsNotNull:
		q = q.Where("? IS NOT NULL", bun.Ident(field))
	case Between:
		values := parseCommaSeparatedValues(fmt.Sprintf("%v", value))
		if len(values) != 2 {
			return q, NewInvalidBetweenValueError(field, fmt.Sprintf("%v", value))
		}
		q = q.Where("? BETWEEN ? AND ?", bun.Ident(field), values[0], values[1])
	case NotBetween:
		values := parseCommaSeparatedValues(fmt.Sprintf("%v", value))
		if len(values) != 2 {
			return q, NewValidationError(field, string(NotBetween), fmt.Sprintf("%v", value), "Not between operator requires exactly two comma-separated values", "Use format: 'value1,value2' (e.g., '10,20')")
		}
		q = q.Where("? NOT BETWEEN ? AND ?", bun.Ident(field), values[0], values[1])
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

// func SplitSeq(s, sep string) func(func(string) bool) {
// 	return func(yield func(string) bool) {
// 		parts := strings.Split(s, sep)
// 		for _, part := range parts {
// 			if !yield(part) {
// 				return
// 			}
// 		}
// 	}
// }
