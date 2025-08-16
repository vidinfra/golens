package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type testUser struct {
	Name  string `gorm:"column:name"`
	Email string `gorm:"column:email"`
	ID    int    `gorm:"column:id;primaryKey;autoIncrement"`
	Age   int    `gorm:"column:age"`
}

func setupDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err, "open sqlite")

	require.NoError(t, db.AutoMigrate(&testUser{}), "migrate")

	seed := []testUser{
		{Name: "alice", Email: "a@x", Age: 20},
		{Name: "alina", Email: "b@x", Age: 22},
		{Name: "bob", Email: "c@x", Age: 17},
	}
	require.NoError(t, db.Create(&seed).Error, "seed")
	return db
}

func TestApplier_Apply_FiltersAndSort(t *testing.T) {
	db := setupDB(t)

	v := NewValidator([]string{"name", "age"}, nil)
	a := NewApplier(v)

	filters := []Filter{
		{Field: "name", Operator: StartsWith, Value: "ali"},
		{Field: "age", Operator: GreaterThanOrEq, Value: "20"},
	}
	res, err := a.Apply(db.Model(&testUser{}), filters, "-age", []string{"name", "age"})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.True(t, res.OK(), "unexpected errors: %+v", res.Errors)

	var got []testUser
	require.NoError(t, res.Query.Find(&got).Error)
	require.Len(t, got, 2)
	assert.Equal(t, 22, got[0].Age)
	assert.Equal(t, 20, got[1].Age)
}

func TestApplier_SortNotAllowed(t *testing.T) {
	db := setupDB(t)
	v := NewValidator([]string{"name", "age"}, nil)
	a := NewApplier(v)

	res, _ := a.Apply(db.Model(&testUser{}), nil, "email", []string{"name"}) // email not allowed
	require.NotNil(t, res)
	assert.False(t, res.OK(), "expected sort error")
	assert.GreaterOrEqual(t, len(res.Errors.Errors), 1)
}
