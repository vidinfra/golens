package filter

import "fmt"

// FilterConfig is optional configuration for a specific field
type FilterConfig struct {
	Field            string
	DefaultOperator  Clause
	Description      string
	AllowedOperators []Clause
}

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
