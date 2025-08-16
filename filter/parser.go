package filter

import (
	"net/url"
	"strings"
)

// Parser handles parsing of URL query parameters into filters.
type Parser struct {
	queryValues url.Values
	// Optional: key prefix, defaults to "filter"
	prefix string
}

// ParseResult represents the result of parsing operations.
type ParseResult struct {
	Errors  *FilterErrors
	Filters []Filter
}

func NewParser(queryValues url.Values) *Parser {
	return &Parser{
		queryValues: queryValues,
		prefix:      "filter",
	}
}

// WithPrefix allows customizing the key prefix (default "filter").
// Example: p.WithPrefix("q").Parse() will parse q[field] keys.
func (p *Parser) WithPrefix(prefix string) *Parser {
	if prefix != "" {
		p.prefix = prefix
	}
	return p
}

// Parse extracts filters from query parameters with error handling.
// Supports both:
//  1. JSON:   filter[field][operator]=value
//  2. Simple: filter[field]=value    (assumes Equals)
func (p *Parser) Parse() *ParseResult {
	res := &ParseResult{
		Errors:  &FilterErrors{},
		Filters: make([]Filter, 0, len(p.queryValues)),
	}

	prefixOpen := p.prefix + "["
	prefixClose := "]"

	for key, values := range p.queryValues {
		if !strings.HasPrefix(key, prefixOpen) || len(values) == 0 {
			continue
		}
		val := strings.TrimSpace(values[0])
		if val == "" {
			// Leave empty handling to validator (so it can give operator-specific messages)
			// but we still emit a parsing error if the key itself is malformed.
			// For a well-formed key, empty value isn't a parsing error.
		}

		// Fast path: reject keys that don't end with ']'
		if !strings.HasSuffix(key, prefixClose) {
			res.Errors.Add(NewInvalidFilterFormatError(key, val))
			continue
		}

		inner := key[len(prefixOpen) : len(key)-len(prefixClose)] // content inside filter[...]
		parts := strings.Split(inner, "][")

		switch len(parts) {
		case 1:
			// Simple format: filter[field]=value
			field := strings.TrimSpace(parts[0])
			if field == "" {
				res.Errors.Add(NewValidationError("", "", val, "Empty field name in filter"))
				continue
			}
			res.Filters = append(res.Filters, Filter{
				Field:    field,
				Operator: Equals,
				Value:    val,
			})

		case 2:
			// JSON API format: filter[field][operator]=value
			field := strings.TrimSpace(parts[0])
			opStr := strings.TrimSpace(parts[1])
			if field == "" || opStr == "" {
				res.Errors.Add(NewInvalidFilterFormatError(key, val))
				continue
			}
			clause := Clause(opStr)
			if !clause.IsValid() {
				res.Errors.Add(NewInvalidOperatorError(opStr))
				continue
			}
			res.Filters = append(res.Filters, Filter{
				Field:    field,
				Operator: clause,
				Value:    val,
			})

		default:
			// Anything else is malformed: filter[field][op][extra]...
			res.Errors.Add(NewInvalidFilterFormatError(key, val))
		}
	}

	return res
}
