package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidator_AllowFields(t *testing.T) {
	v := NewValidator([]string{"name", "email"}, nil)

	assert.True(t, v.IsFilterAllowed(Filter{Field: "name", Operator: Equals, Value: "a"}))
	assert.False(t, v.IsFilterAllowed(Filter{Field: "status", Operator: Equals, Value: "active"}))

	err := v.ValidateFilter(Filter{Field: "status", Operator: Equals, Value: "active"})
	require.NotNil(t, err)
	assert.Equal(t, CodeFilterValidation, err.Code)
}

func TestValidator_AllowConfigs(t *testing.T) {
	cfg := []FilterConfig{
		{Field: "name", AllowedOperators: []Clause{Equals, Contains}},
		{Field: "age", AllowedOperators: []Clause{GreaterThan, LessThanOrEq}},
	}
	v := NewValidator(nil, cfg)

	// allowed operator
	assert.Nil(t, v.ValidateFilter(Filter{Field: "name", Operator: Contains, Value: "al"}))

	// operator not allowed for field
	assert.NotNil(t, v.ValidateFilter(Filter{Field: "name", Operator: Between, Value: "1,2"}))

	// field not in configs
	assert.NotNil(t, v.ValidateFilter(Filter{Field: "email", Operator: Equals, Value: "a@b"}))
}

func TestValidator_ValueShapes(t *testing.T) {
	v := NewValidator([]string{"age"}, nil)

	assert.NotNil(t, v.ValidateFilter(Filter{Field: "age", Operator: Between, Value: "1"}))
	assert.Nil(t, v.ValidateFilter(Filter{Field: "age", Operator: Between, Value: "1,2"}))

	assert.Nil(t, v.ValidateFilter(Filter{Field: "age", Operator: IsNull, Value: ""}))
}
