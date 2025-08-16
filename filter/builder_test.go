package filter

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newGinCtxWithQuery(q url.Values) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest("GET", "/?"+q.Encode(), nil)
	c.Request = req
	return c, w
}

func TestBuilder_Apply_FullFlow(t *testing.T) {
	// DB + data
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err, "open sqlite")
	require.NoError(t, db.AutoMigrate(&testUser{}), "migrate")

	seed := []testUser{
		{Name: "alice", Email: "a@x", Age: 20},
		{Name: "alina", Email: "b@x", Age: 22},
		{Name: "bob", Email: "c@x", Age: 17},
	}
	require.NoError(t, db.Create(&seed).Error, "seed")

	// filter[name][starts-with]=ali&sort=-age
	q := url.Values{}
	q.Set("filter[name][starts-with]", "ali")
	q.Set("sort", "-age")
	c, _ := newGinCtxWithQuery(q)

	res := New(c, db.Model(&testUser{})).
		AllowAll("name", "age", "email").
		Apply().
		Result()

	require.NotNil(t, res)
	assert.True(t, res.OK(), "unexpected errors: %+v", res.Errors)

	var got []testUser
	require.NoError(t, res.Query.Find(&got).Error)
	require.Len(t, got, 2)
	assert.Equal(t, 22, got[0].Age)
	assert.Equal(t, 20, got[1].Age)
}

func TestBuilder_SortRejected(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&testUser{}))

	q := url.Values{}
	q.Set("sort", "email") // not allowed below
	c, _ := newGinCtxWithQuery(q)

	b := New(c, db.Model(&testUser{})).
		AllowFields("name", "age").
		AllowSorts("age"). // email is not allowed
		Apply()

	assert.False(t, b.OK(), "expected builder not OK due to sort rejection")
}
