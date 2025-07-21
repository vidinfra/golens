package filter

import (
	"net/url"
	"strings"
)

// Parser handles parsing of URL query parameters into filters
type Parser struct {
	queryValues url.Values
}

// ParseResult represents the result of parsing operations
type ParseResult struct {
	Filters []Filter
	Errors  *FilterErrors
}

func NewParser(queryValues url.Values) *Parser {
	return &Parser{
		queryValues: queryValues,
	}
}

// Parse extracts filters from query parameters with error handling
func (p *Parser) Parse() *ParseResult {
	result := &ParseResult{
		Filters: []Filter{},
		Errors:  &FilterErrors{},
	}

	// Parse JSON API format: filter[field][operator]=value
	jsonAPIFilters, jsonAPIErrors := p.parseJSONAPIFormat()
	result.Filters = append(result.Filters, jsonAPIFilters...)
	if len(jsonAPIErrors) > 0 {
		result.Errors.Errors = append(result.Errors.Errors, jsonAPIErrors...)
	}

	// Fallback to simple format: filter[field]=value (assumes 'eq' operator)
	simpleFilters, simpleErrors := p.parseSimpleFormat()
	result.Filters = append(result.Filters, simpleFilters...)
	if len(simpleErrors) > 0 {
		result.Errors.Errors = append(result.Errors.Errors, simpleErrors...)
	}

	return result
}

func (p *Parser) parseJSONAPIFormat() ([]Filter, []*FilterError) {
	var filters []Filter
	var errors []*FilterError

	for key, values := range p.queryValues {
		if !strings.HasPrefix(key, "filter[") || len(values) == 0 {
			continue
		}

		if field, operator, ok := parseFilterKey(key); ok {
			value := values[0]

			// Validate operator
			clause := Clause(operator)
			if !clause.IsValid() {
				errors = append(errors, NewInvalidOperatorError(operator))
				continue
			}

			filters = append(filters, Filter{
				Field:    field,
				Operator: clause,
				Value:    value,
			})
		} else {
			errors = append(errors, NewInvalidFilterFormatError(key, values[0]))
		}
	}

	return filters, errors
}

func (p *Parser) parseSimpleFormat() ([]Filter, []*FilterError) {
	var filters []Filter
	var errors []*FilterError

	for key, values := range p.queryValues {
		if !strings.HasPrefix(key, "filter[") || !strings.HasSuffix(key, "]") || len(values) == 0 {
			continue
		}

		// Skip if already handled by full format
		if strings.Count(key, "][") > 0 {
			continue
		}

		// Extract field name from filter[field]
		field := key[7 : len(key)-1] // Remove "filter[" and "]"
		if field == "" {
			errors = append(errors, NewValidationError("", "", values[0], "Empty field name in filter"))
			continue
		}

		value := values[0]

		filters = append(filters, Filter{
			Field:    field,
			Operator: Equals,
			Value:    value,
		})
	}

	return filters, errors
}

// parseFilterKey extracts field and operator from filter[field][operator] format
func parseFilterKey(key string) (field, operator string, ok bool) {
	if !strings.HasPrefix(key, "filter[") || !strings.HasSuffix(key, "]") {
		return "", "", false
	}

	inner := key[7 : len(key)-1]
	parts := strings.Split(inner, "][")
	if len(parts) != 2 {
		return "", "", false
	}

	return parts[0], parts[1], true
}
