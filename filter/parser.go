package filter

import (
	"net/url"
	"strings"
)

// Parser handles parsing of URL query parameters into filters
type Parser struct {
	queryValues url.Values
}

func NewParser(queryValues url.Values) *Parser {
	return &Parser{
		queryValues: queryValues,
	}
}

func (p *Parser) Parse() []Filter {
	var filters []Filter

	// Parse JSON API format: filter[field][operator]=value
	filters = append(filters, p.parseJSONAPIFormat()...)

	// Fallback to simple format: filter[field]=value (assumes 'eq' operator)
	filters = append(filters, p.parseSimpleFormat()...)

	return filters
}

func (p *Parser) parseJSONAPIFormat() []Filter {
	var filters []Filter

	for key, values := range p.queryValues {
		if !strings.HasPrefix(key, "filter[") || len(values) == 0 {
			continue
		}

		if field, operator, ok := parseFilterKey(key); ok {
			value := values[0]

			filters = append(filters, Filter{
				Field:    field,
				Operator: Clause(operator),
				Value:    value,
			})
		}
	}

	return filters
}

func (p *Parser) parseSimpleFormat() []Filter {
	var filters []Filter

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
			continue
		}

		value := values[0]

		filters = append(filters, Filter{
			Field:    field,
			Operator: Equals,
			Value:    value,
		})
	}

	return filters
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
