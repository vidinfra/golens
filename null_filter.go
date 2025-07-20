// File: pkg/filter/clause.go
package filter

// Clause represents a filter operation type
type Clause string

const (
	Equals          Clause = "eq"
	NotEquals       Clause = "ne"
	Contains        Clause = "like"
	NotContains     Clause = "not-like"
	StartsWith      Clause = "starts-with"
	EndsWith        Clause = "ends-with"
	GreaterThan     Clause = "gt"
	GreaterThanOrEq Clause = "gte"
	LessThan        Clause = "lt"
	LessThanOrEq    Clause = "lte"
	In              Clause = "in"
	NotIn           Clause = "not-in"
	IsNull          Clause = "null"
	IsNotNull       Clause = "not-null"
	Between         Clause = "between"
	NotBetween      Clause = "not-between"
)

// IsValid checks if the clause is a valid operator
func (c Clause) IsValid() bool {
	switch c {
	case Equals, NotEquals, Contains, NotContains, StartsWith, EndsWith,
		GreaterThan, GreaterThanOrEq, LessThan, LessThanOrEq,
		In, NotIn, IsNull, IsNotNull, Between, NotBetween:
		return true
	default:
		return false
	}
}

// String returns the string representation
func (c Clause) String() string {
	return string(c)
}
// done 1
// File: pkg/filter/config.go
package filter

import "fmt"

// FilterConfig is optional configuration for a specific field
type FilterConfig struct {
	Field            string
	DefaultOperator  Clause
	Description      string
	AllowedOperators []Clause
}

// AllowedFilter creates a FilterConfig with specified operators
func AllowedFilter(field string, operators ...Clause) FilterConfig {
	if len(operators) == 0 {
		operators = []Clause{Equals}
	}

	return FilterConfig{
		Field:            field,
		AllowedOperators: operators,
		DefaultOperator:  operators[0],
		Description:      fmt.Sprintf("Filter by %s", field),
	}
}

//done 2
// File: pkg/filter/filter.go
package filter

// Filter represents a single filter condition
type Filter struct {
	Value    any    `json:"value"`
	Field    string `json:"field"`
	Operator Clause `json:"operator"`
}
// done 3
// File: pkg/filter/parser.go
package filter

import (
	"net/url"
	"strings"
)

// Parser handles parsing of URL query parameters into filters
type Parser struct {
	queryValues url.Values
}

// NewParser creates a new parser with URL query values
func NewParser(queryValues url.Values) *Parser {
	return &Parser{
		queryValues: queryValues,
	}
}

// Parse extracts filters from JSON API compliant query parameters
func (p *Parser) Parse() []Filter {
	var filters []Filter

	// Parse JSON API format: filter[field][operator]=value
	filters = append(filters, p.parseJSONAPIFormat()...)

	// Fallback to simple format: filter[field]=value (assumes 'eq' operator)
	filters = append(filters, p.parseSimpleFormat()...)

	return filters
}

// parseJSONAPIFormat handles: filter[name][eq]=AWS&filter[status][in]=active,pending
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

// parseSimpleFormat handles: filter[name]=AWS (assumes eq operator)
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
			Operator: Equals, // Default to equals
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
//done 4
// File: pkg/filter/validator.go
package filter

import "slices"

// Validator handles validation of filters against configurations
type Validator struct {
	allowedFields []string
	configs       []FilterConfig
	useConfigs    bool
}

// NewValidator creates a new validator
func NewValidator(allowedFields []string, configs []FilterConfig) *Validator {
	return &Validator{
		allowedFields: allowedFields,
		configs:       configs,
		useConfigs:    len(configs) > 0,
	}
}

// IsFilterAllowed checks if a filter is allowed based on configuration
func (v *Validator) IsFilterAllowed(filter Filter) bool {
	// Apply validation based on configuration
	if v.useConfigs {
		return v.isFilterAllowedByConfig(filter)
	} else if v.allowedFields != nil {
		return slices.Contains(v.allowedFields, filter.Field)
	}
	return true
}

// isFilterAllowedByConfig checks if filter is allowed with detailed configs
func (v *Validator) isFilterAllowedByConfig(filter Filter) bool {
	if v.configs == nil {
		return true
	}

	for _, config := range v.configs {
		if config.Field == filter.Field {
			return slices.Contains(config.AllowedOperators, filter.Operator)
		}
	}

	return false
}
// done 5
// File: pkg/filter/applier.go
package filter

import (
	"fmt"
	"slices"
	"strings"

	"github.com/uptrace/bun"
)

// SplitSeq is a utility function that mimics strings.SplitSeq for compatibility
func SplitSeq(s, sep string) func(func(string) bool) {
	return func(yield func(string) bool) {
		parts := strings.Split(s, sep)
		for _, part := range parts {
			if !yield(part) {
				return
			}
		}
	}
}

// Applier handles applying filters to database queries
type Applier struct {
	validator *Validator
}

// NewApplier creates a new applier with a validator
func NewApplier(validator *Validator) *Applier {
	return &Applier{
		validator: validator,
	}
}

// ApplyFilters applies filters to a bun query with validation
func (a *Applier) ApplyFilters(q *bun.SelectQuery, filters []Filter) *bun.SelectQuery {
	for _, filter := range filters {
		if a.validator != nil && !a.validator.IsFilterAllowed(filter) {
			continue
		}
		q = a.applyFilter(q, filter)
	}
	return q
}

// ApplySort adds sorting to the query
func (a *Applier) ApplySort(q *bun.SelectQuery, sortParam string, allowedSorts []string) *bun.SelectQuery {
	if sortParam == "" {
		return q
	}

	for s := range SplitSeq(sortParam, ",") {
		sortField := strings.TrimSpace(s)
		if sortField == "" {
			continue
		}

		// Check for descending order (-)
		desc := strings.HasPrefix(sortField, "-")
		if desc {
			sortField = sortField[1:]
		}

		// Check if field is allowed
		if allowedSorts != nil && !slices.Contains(allowedSorts, sortField) {
			continue
		}

		if desc {
			q = q.Order(sortField + " DESC")
		} else {
			q = q.Order(sortField + " ASC")
		}
	}
	return q
}

// applyFilter applies a single filter condition
func (a *Applier) applyFilter(q *bun.SelectQuery, filter Filter) *bun.SelectQuery {
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
		if len(values) == 2 {
			q = q.Where("? BETWEEN ? AND ?", bun.Ident(field), values[0], values[1])
		}
	case NotBetween:
		values := parseCommaSeparatedValues(fmt.Sprintf("%v", value))
		if len(values) == 2 {
			q = q.Where("? NOT BETWEEN ? AND ?", bun.Ident(field), values[0], values[1])
		}
	}

	return q
}

// parseCommaSeparatedValues splits and trims comma-separated values
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
// done 6
// File: pkg/filter/builder.go
package filter

import (
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// Builder holds filter configuration and provides fluent API
type Builder struct {
	ctx           *gin.Context
	query         *bun.SelectQuery
	allowedFields []string
	allowedSorts  []string
	configs       []FilterConfig
	useConfigs    bool
	parser        *Parser
	validator     *Validator
	applier       *Applier
}

// New creates a new filter builder with context and query
func New(c *gin.Context, q *bun.SelectQuery) *Builder {
	parser := NewParser(c.Request.URL.Query())
	
	return &Builder{
		ctx:    c,
		query:  q,
		parser: parser,
	}
}

// AllowFields sets allowed field names for filtering
func (b *Builder) AllowFields(fields ...string) *Builder {
	b.allowedFields = fields
	b.updateValidator()
	return b
}

// AllowSorts sets allowed field names for sorting
func (b *Builder) AllowSorts(fields ...string) *Builder {
	b.allowedSorts = fields
	return b
}

// AllowAll sets the same fields for both filtering and sorting
func (b *Builder) AllowAll(fields ...string) *Builder {
	b.allowedFields = fields
	b.allowedSorts = fields
	b.updateValidator()
	return b
}

// AllowConfigs sets detailed filter configurations with operator validation
func (b *Builder) AllowConfigs(configs ...FilterConfig) *Builder {
	b.configs = configs
	b.useConfigs = true
	b.updateValidator()
	return b
}

// updateValidator creates or updates the validator based on current configuration
func (b *Builder) updateValidator() {
	b.validator = NewValidator(b.allowedFields, b.configs)
	b.applier = NewApplier(b.validator)
}

// Apply parses filters from context and applies them to query
func (b *Builder) Apply() *Builder {
	if b.applier == nil {
		b.updateValidator()
	}
	
	filters := b.parser.Parse()
	b.query = b.applier.ApplyFilters(b.query, filters)
	return b
}

// ApplySort adds sorting to the query using builder configuration
func (b *Builder) ApplySort() *Builder {
	if b.applier == nil {
		b.updateValidator()
	}
	
	sort := b.ctx.Query("sort")
	
	// Use builder's allowed sorts if set, otherwise fall back to allowed fields
	allowedFields := b.allowedSorts
	if allowedFields == nil {
		allowedFields = b.allowedFields
	}
	
	b.query = b.applier.ApplySort(b.query, sort, allowedFields)
	return b
}

// Query returns the modified query
func (b *Builder) Query() *bun.SelectQuery {
	return b.query
}

// parse extracts filters from JSON API compliant query parameters (kept for backward compatibility)
func (b *Builder) parse() []Filter {
	return b.parser.Parse()
}

// applyFilters applies filters with appropriate validation (kept for backward compatibility)
func (b *Builder) applyFilters(q *bun.SelectQuery, filters []Filter) *bun.SelectQuery {
	if b.applier == nil {
		b.updateValidator()
	}
	return b.applier.ApplyFilters(q, filters)
}

// isFilterAllowedByConfig checks if filter is allowed with detailed configs (kept for backward compatibility)
func (b *Builder) isFilterAllowedByConfig(filter Filter) bool {
	if b.validator == nil {
		b.updateValidator()
	}
	return b.validator.isFilterAllowedByConfig(filter)
}

// File: internal/utils/strings.go
package utils

import "strings"

// ParseCommaSeparatedValues splits and trims comma-separated values  
func ParseCommaSeparatedValues(value string) []string {
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