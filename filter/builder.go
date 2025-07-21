package filter

import (
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

// Builder holds filter configuration and provides fluent API
type Builder struct {
	ctx           *gin.Context     // 8 bytes (pointer)
	query         *bun.SelectQuery // 8 bytes (pointer)
	parser        *Parser          // 8 bytes (pointer)
	validator     *Validator       // 8 bytes (pointer)
	applier       *Applier         // 8 bytes (pointer)
	result        *Result          // 8 bytes (pointer)
	allowedFields []string         // 24 bytes (slice header: ptr+len+cap)
	allowedSorts  []string         // 24 bytes (slice header: ptr+len+cap)
	configs       []FilterConfig   // 24 bytes (slice header: ptr+len+cap)
	useConfigs    bool             // 1 byte
}

func New(c *gin.Context, q *bun.SelectQuery) *Builder {
	parser := NewParser(c.Request.URL.Query())

	return &Builder{
		ctx:    c,
		query:  q,
		parser: parser,
	}
}

func (b *Builder) AllowFields(fields ...string) *Builder {
	b.allowedFields = fields
	b.updateValidator()
	return b
}

func (b *Builder) AllowSorts(fields ...string) *Builder {
	b.allowedSorts = fields
	return b
}

func (b *Builder) AllowAll(fields ...string) *Builder {
	b.allowedFields = fields
	b.allowedSorts = fields
	b.updateValidator()
	return b
}

func (b *Builder) AllowConfigs(configs ...FilterConfig) *Builder {
	b.configs = configs
	b.useConfigs = true
	b.updateValidator()
	return b
}

func (b *Builder) updateValidator() {
	b.validator = NewValidator(b.allowedFields, b.configs)
	b.applier = NewApplier(b.validator)
}

func (b *Builder) Apply() *Builder {
	if b.applier == nil {
		b.updateValidator()
	}

	parseResult := b.parser.Parse()

	b.result = NewResult(b.query)

	if parseResult.Errors.HasErrors() {
		b.result.AddErrors(parseResult.Errors.Errors...)
	}

	if len(parseResult.Filters) > 0 {
		result, err := b.applier.ApplyFilters(b.query, parseResult.Filters)

		if result.HasErrors() {
			b.result.AddErrors(result.Errors.Errors...)
		}

		if err == nil {
			b.query = result.Query
			b.result.Query = result.Query
		}
	}

	return b
}

func (b *Builder) ApplySort() *Builder {
	if b.applier == nil {
		b.updateValidator()
	}

	if b.result == nil {
		b.result = NewResult(b.query)
	}

	sort := b.ctx.Query("sort")

	allowedFields := b.allowedSorts
	if allowedFields == nil {
		allowedFields = b.allowedFields
	}

	newQuery, sortErrors := b.applier.ApplySort(b.query, sort, allowedFields)
	b.query = newQuery
	b.result.Query = newQuery

	for _, err := range sortErrors {
		b.result.AddError(err)
	}

	return b
}

func (b *Builder) Query() *bun.SelectQuery {
	return b.query
}

func (b *Builder) Result() *Result {
	if b.result == nil {
		return NewResult(b.query)
	}
	return b.result
}

func (b *Builder) HasErrors() bool {
	if b.result == nil {
		return false // No result = no errors
	}
	return b.result.HasErrors()
}

func (b *Builder) GetErrors() *FilterErrors {
	if b.result == nil {
		return &FilterErrors{}
	}
	return b.result.Errors
}
