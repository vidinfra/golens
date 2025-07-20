package filter

import (
	"fmt"
	"slices"
	"strings"

	"github.com/uptrace/bun"
)

func (b *Builder) Apply() *Builder {
	filters := b.parse()
	b.query = b.applyFilters(b.query, filters)
	return b
}

func (b *Builder) ApplySort() *Builder {
	sort := b.ctx.Query("sort")
	if sort == "" {
		return b
	}

	allowedFields := b.allowedSorts
	if allowedFields == nil {
		allowedFields = b.allowedFields
	}

	for _, s := range strings.Split(sort, ",") {
		sortField := strings.TrimSpace(s)
		if sortField == "" {
			continue
		}

		desc := strings.HasPrefix(sortField, "-")
		if desc {
			sortField = sortField[1:]
		}

		if allowedFields != nil && !slices.Contains(allowedFields, sortField) {
			continue
		}

		if desc {
			b.query = b.query.Order(sortField + " DESC")
		} else {
			b.query = b.query.Order(sortField + " ASC")
		}
	}

	return b
}

func (b *Builder) parse() []Filter {
	var filters []Filter

	filters = append(filters, b.parseJSONAPIFormat()...)
	filters = append(filters, b.parseSimpleFormat()...)

	return filters
}

func (b *Builder) parseJSONAPIFormat() []Filter {
	var filters []Filter

	for key, values := range b.ctx.Request.URL.Query() {
		if !strings.HasPrefix(key, "filter[") || len(values) == 0 {
			continue
		}

		if field, operator, ok := ParseFilterKey(key); ok {
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

func (b *Builder) parseSimpleFormat() []Filter {
	var filters []Filter

	for key, values := range b.ctx.Request.URL.Query() {
		if !strings.HasPrefix(key, "filter[") || !strings.HasSuffix(key, "]") || len(values) == 0 {
			continue
		}

		if strings.Count(key, "][") > 0 {
			continue
		}

		field := key[7 : len(key)-1]
		value := values[0]

		filters = append(filters, Filter{
			Field:    field,
			Operator: Equals,
			Value:    value,
		})
	}

	return filters
}

func (b *Builder) applyFilters(q *bun.SelectQuery, filters []Filter) *bun.SelectQuery {
	for _, filter := range filters {
		if b.useConfigs {
			if !b.isFilterAllowedByConfig(filter) {
				continue
			}
		} else if b.allowedFields != nil && !slices.Contains(b.allowedFields, filter.Field) {
			continue
		}
		q = applyFilter(q, filter)
	}
	return q
}

func (b *Builder) isFilterAllowedByConfig(filter Filter) bool {
	for _, config := range b.configs {
		if config.Field == filter.Field {
			return slices.Contains(config.AllowedOperators, filter.Operator)
		}
	}
	return false
}

func applyFilter(q *bun.SelectQuery, filter Filter) *bun.SelectQuery {
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
		values := ParseCommaSeparatedValues(fmt.Sprintf("%v", value))
		q = q.Where("? IN (?)", bun.Ident(field), bun.In(values))
	case NotIn:
		values := ParseCommaSeparatedValues(fmt.Sprintf("%v", value))
		q = q.Where("? NOT IN (?)", bun.Ident(field), bun.In(values))
	case IsNull:
		q = q.Where("? IS NULL", bun.Ident(field))
	case IsNotNull:
		q = q.Where("? IS NOT NULL", bun.Ident(field))
	case Between:
		values := ParseCommaSeparatedValues(fmt.Sprintf("%v", value))
		if len(values) == 2 {
			q = q.Where("? BETWEEN ? AND ?", bun.Ident(field), values[0], values[1])
		}
	case NotBetween:
		values := ParseCommaSeparatedValues(fmt.Sprintf("%v", value))
		if len(values) == 2 {
			q = q.Where("? NOT BETWEEN ? AND ?", bun.Ident(field), values[0], values[1])
		}
	}
	return q
}
