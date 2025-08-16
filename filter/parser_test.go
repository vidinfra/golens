package filter

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParser_SimpleFormat(t *testing.T) {
	q := url.Values{}
	q.Set("filter[name]", "alice")
	p := NewParser(q)

	res := p.Parse()
	require.NotNil(t, res)
	assert.True(t, res.Errors.OK(), "unexpected parse errors: %+v", res.Errors)
	require.Len(t, res.Filters, 1)

	f := res.Filters[0]
	assert.Equal(t, "name", f.Field)
	assert.Equal(t, Equals, f.Operator)
	assert.Equal(t, "alice", f.Value)
}

func TestParser_JSONAPIFormat(t *testing.T) {
	q := url.Values{}
	q.Set("filter[name][like]", "ali")
	q.Set("filter[age][gte]", "18")
	p := NewParser(q)

	res := p.Parse()
	require.NotNil(t, res)
	assert.True(t, res.Errors.OK(), "unexpected parse errors: %+v", res.Errors)
	assert.Len(t, res.Filters, 2)
}

func TestParser_InvalidOperator(t *testing.T) {
	q := url.Values{}
	q.Set("filter[name][wat]", "x")
	p := NewParser(q)

	res := p.Parse()
	require.NotNil(t, res)
	assert.False(t, res.Errors.OK(), "expected invalid-operator error")
}
