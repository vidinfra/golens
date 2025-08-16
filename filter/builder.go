package filter

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Builder holds filter configuration and provides a fluent API.
type Builder struct {
	ctx           *gin.Context
	query         *gorm.DB
	parser        *Parser
	validator     *Validator
	applier       *Applier
	result        *Result
	allowedFields []string
	allowedSorts  []string
	configs       []FilterConfig
	useConfigs    bool
}

// New creates a new Builder bound to a Gin context and a base *gorm.DB query.
func New(c *gin.Context, q *gorm.DB) *Builder {
	parser := NewParser(c.Request.URL.Query())
	return &Builder{
		ctx:    c,
		query:  q,
		parser: parser,
	}
}

// AllowFields sets the field allowlist for filtering.
func (b *Builder) AllowFields(fields ...string) *Builder {
	b.allowedFields = fields
	b.updateValidator()
	return b
}

// AllowSorts sets the field allowlist for sorting.
// If not provided, sorting falls back to allowedFields.
func (b *Builder) AllowSorts(fields ...string) *Builder {
	b.allowedSorts = fields
	return b
}

// AllowAll sets the same allowlist for both filtering and sorting.
func (b *Builder) AllowAll(fields ...string) *Builder {
	b.allowedFields = fields
	b.allowedSorts = fields
	b.updateValidator()
	return b
}

// AllowConfigs sets per-field operator allowlists (most precise control).
func (b *Builder) AllowConfigs(configs ...FilterConfig) *Builder {
	b.configs = configs
	b.useConfigs = true
	b.updateValidator()
	return b
}

// updateValidator rebuilds the validator+applier when allowlists/configs change.
func (b *Builder) updateValidator() {
	b.validator = NewValidator(b.allowedFields, b.configs)
	b.applier = NewApplier(b.validator)
}

// Apply parses query params and applies BOTH filters and sort in one pass.
// This is the only method callers need to invoke.
func (b *Builder) Apply() *Builder {
	if b.applier == nil {
		b.updateValidator()
	}

	// Parse incoming filters from the query string.
	parseResult := b.parser.Parse()

	// Seed a fresh result around the current query.
	b.result = NewResult(b.query)

	// Carry parser (format) errors forward.
	if !parseResult.Errors.OK() {
		b.result.AddErrors(parseResult.Errors.Errors...)
	}

	// Determine allowed sort fields: explicit list or fallback to allowedFields.
	allowedSorts := b.allowedSorts
	if allowedSorts == nil {
		allowedSorts = b.allowedFields
	}

	// Pull sort param, e.g. ?sort=-created_at,name
	sortParam := b.ctx.Query("sort")

	// Single entrypoint: run filters + sort
	res, _ := b.applier.Apply(b.query, parseResult.Filters, sortParam, allowedSorts)

	// Merge any applier errors into the builder result
	if res != nil && !res.OK() {
		b.result.AddErrors(res.Errors.Errors...)
	}

	// Update final query
	if res != nil && res.Query != nil {
		b.query = res.Query
		b.result.Query = res.Query
	}

	return b
}

// Query returns the final *gorm.DB after Apply().
func (b *Builder) Query() *gorm.DB {
	return b.query
}

// Result returns the accumulated result (errors + success flag).
func (b *Builder) Result() *Result {
	if b.result == nil {
		return NewResult(b.query)
	}
	return b.result
}

// OK reports whether the builder completed without errors.
func (b *Builder) OK() bool {
	return b.result != nil && b.result.OK()
}

// GetErrors returns the aggregated FilterErrors (never nil).
func (b *Builder) GetErrors() *FilterErrors {
	if b.result == nil {
		return &FilterErrors{}
	}
	return b.result.Errors
}
