package filter

import (
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type Filter struct {
	Value    any    `json:"value"`
	Field    string `json:"field"`
	Operator Clause `json:"operator"`
}

type Builder struct {
	ctx           *gin.Context
	query         *bun.SelectQuery
	allowedFields []string
	allowedSorts  []string
	configs       []FilterConfig
	useConfigs    bool
}

func New(c *gin.Context, q *bun.SelectQuery) *Builder {
	return &Builder{
		ctx:   c,
		query: q,
	}
}

func (b *Builder) AllowedFields(fields ...string) *Builder {
	b.allowedFields = fields
	return b
}

func (b *Builder) AllowedSorts(fields ...string) *Builder {
	b.allowedSorts = fields
	return b
}

func (b *Builder) AllowAll(fields ...string) *Builder {
	b.allowedFields = fields
	b.allowedSorts = fields
	return b
}

func (b *Builder) AllowConfigs(configs ...FilterConfig) *Builder {
	b.configs = configs
	b.useConfigs = true
	return b
}

func (b *Builder) Query() *bun.SelectQuery {
	return b.query
}
