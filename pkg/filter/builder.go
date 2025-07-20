package filter

import (
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

	filters := b.parser.Parse()
	b.query = b.applier.ApplyFilters(b.query, filters)
	return b
}

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

func (b *Builder) Query() *bun.SelectQuery {
	return b.query
}

func (b *Builder) parse() []Filter {
	return b.parser.Parse()
}

func (b *Builder) applyFilters(q *bun.SelectQuery, filters []Filter) *bun.SelectQuery {
	if b.applier == nil {
		b.updateValidator()
	}
	return b.applier.ApplyFilters(q, filters)
}

func (b *Builder) isFilterAllowedByConfig(filter Filter) bool {
	if b.validator == nil {
		b.updateValidator()
	}
	return b.validator.isFilterAllowedByConfig(filter)
}
