package filter

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// --- Test model & fixtures ---

type opUser struct {
	Email *string `gorm:"column:email"`
	Name  string  `gorm:"column:name"`
	ID    int     `gorm:"column:id;primaryKey;autoIncrement"`
	Age   int     `gorm:"column:age"`
}

func mustDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err, "open sqlite")

	require.NoError(t, db.AutoMigrate(&opUser{}), "migrate")

	// Seed data
	emailA := "a@x"
	emailB := "b@x"
	// Name variety for LIKE tests (case-insensitive in SQLite by default)
	rows := []opUser{
		{Name: "alice", Email: &emailA, Age: 20},
		{Name: "alina", Email: &emailB, Age: 22},
		{Name: "bob", Email: nil, Age: 17},
		{Name: "ALF", Email: &emailA, Age: 30},
		{Name: "beta", Email: nil, Age: 10},
	}
	require.NoError(t, db.Create(&rows).Error, "seed")

	return db
}

func run(t *testing.T, db *gorm.DB, f Filter) ([]opUser, *Result) {
	t.Helper()
	v := NewValidator([]string{"name", "email", "age"}, nil) // allow these fields
	a := NewApplier(v)
	res, err := a.Apply(db.Model(&opUser{}), []Filter{f}, "", nil)
	require.NoError(t, err)
	require.NotNil(t, res)
	return fetch(t, res), res
}

func fetch(t *testing.T, res *Result) []opUser {
	t.Helper()
	require.True(t, res != nil)
	var got []opUser
	require.True(t, res.OK(), "unexpected errors: %+v", res.Errors)
	require.NoError(t, res.Query.Find(&got).Error)
	return got
}

func namesOf(us []opUser) []string {
	out := make([]string, len(us))
	for i, u := range us {
		out[i] = u.Name
	}
	sort.Strings(out)
	return out
}

// --- Tests per operator ---

func TestEquals(t *testing.T) {
	db := mustDB(t)
	users, _ := run(t, db, Filter{Field: "name", Operator: Equals, Value: "alice"})
	assert.Equal(t, []string{"alice"}, namesOf(users))
}

func TestNotEquals(t *testing.T) {
	db := mustDB(t)
	users, _ := run(t, db, Filter{Field: "name", Operator: NotEquals, Value: "alice"})
	// everybody except alice
	assert.NotContains(t, namesOf(users), "alice")
}

func TestContains(t *testing.T) {
	db := mustDB(t)
	users, _ := run(t, db, Filter{Field: "name", Operator: Contains, Value: "li"})
	// alice, alina have "li"
	assert.Equal(t, []string{"alice", "alina"}, namesOf(users))
}

func TestNotContains(t *testing.T) {
	db := mustDB(t)
	users, _ := run(t, db, Filter{Field: "name", Operator: NotContains, Value: "li"})
	// should exclude alice, alina
	n := namesOf(users)
	assert.NotContains(t, n, "alice")
	assert.NotContains(t, n, "alina")
}

func TestStartsWith(t *testing.T) {
	db := mustDB(t)
	users, _ := run(t, db, Filter{Field: "name", Operator: StartsWith, Value: "al"})
	// alice, alina, ALF (case-insensitive in SQLite)
	assert.Equal(t, []string{"ALF", "alice", "alina"}, namesOf(users))
}

func TestEndsWith(t *testing.T) {
	db := mustDB(t)
	users, _ := run(t, db, Filter{Field: "name", Operator: EndsWith, Value: "na"})
	assert.Equal(t, []string{"alina"}, namesOf(users))
}

func TestGreaterThan(t *testing.T) {
	db := mustDB(t)
	users, _ := run(t, db, Filter{Field: "age", Operator: GreaterThan, Value: "20"})
	// Age > 20: alina(22), ALF(30)
	assert.Equal(t, []string{"ALF", "alina"}, namesOf(users))
}

func TestGreaterThanOrEq(t *testing.T) {
	db := mustDB(t)
	users, _ := run(t, db, Filter{Field: "age", Operator: GreaterThanOrEq, Value: "20"})
	// Age >= 20: alice(20), alina(22), ALF(30)
	assert.Equal(t, []string{"ALF", "alice", "alina"}, namesOf(users))
}

func TestLessThan(t *testing.T) {
	db := mustDB(t)
	users, _ := run(t, db, Filter{Field: "age", Operator: LessThan, Value: "20"})
	// Age < 20: bob(17), beta(10)
	assert.Equal(t, []string{"beta", "bob"}, namesOf(users))
}

func TestLessThanOrEq(t *testing.T) {
	db := mustDB(t)
	users, _ := run(t, db, Filter{Field: "age", Operator: LessThanOrEq, Value: "20"})
	// Age <= 20: beta(10), bob(17), alice(20)
	assert.Equal(t, []string{"alice", "beta", "bob"}, namesOf(users))
}

func TestIn(t *testing.T) {
	db := mustDB(t)
	users, _ := run(t, db, Filter{Field: "name", Operator: In, Value: "alice, bob"})
	assert.Equal(t, []string{"alice", "bob"}, namesOf(users))
}

func TestNotIn(t *testing.T) {
	db := mustDB(t)
	users, _ := run(t, db, Filter{Field: "name", Operator: NotIn, Value: "alice, bob"})
	// everyone except alice, bob
	n := namesOf(users)
	assert.NotContains(t, n, "alice")
	assert.NotContains(t, n, "bob")
	// sanity: some remain
	assert.Greater(t, len(n), 0)
}

func TestIsNull(t *testing.T) {
	db := mustDB(t)
	users, _ := run(t, db, Filter{Field: "email", Operator: IsNull, Value: ""})
	// Those seeded with nil email: bob, beta
	assert.Equal(t, []string{"beta", "bob"}, namesOf(users))
}

func TestIsNotNull(t *testing.T) {
	db := mustDB(t)
	users, _ := run(t, db, Filter{Field: "email", Operator: IsNotNull, Value: ""})
	// Not null: alice, alina, ALF
	assert.Equal(t, []string{"ALF", "alice", "alina"}, namesOf(users))
}

func TestBetween(t *testing.T) {
	db := mustDB(t)
	// Age between 18 and 25: alice(20), alina(22)
	users, _ := run(t, db, Filter{Field: "age", Operator: Between, Value: "18,25"})
	assert.Equal(t, []string{"alice", "alina"}, namesOf(users))
}

func TestNotBetween(t *testing.T) {
	db := mustDB(t)
	// Age not between 18 and 25: beta(10), bob(17), ALF(30)
	users, _ := run(t, db, Filter{Field: "age", Operator: NotBetween, Value: "18,25"})
	assert.Equal(t, []string{"ALF", "beta", "bob"}, namesOf(users))
}

func TestNotBetween_InvalidValue_ShouldError(t *testing.T) {
	db := mustDB(t)
	v := NewValidator([]string{"age"}, nil)
	a := NewApplier(v)

	// invalid: only one value
	res, _ := a.Apply(db.Model(&opUser{}), []Filter{
		{Field: "age", Operator: NotBetween, Value: "18"},
	}, "", nil)

	require.NotNil(t, res)
	assert.False(t, res.OK(), "expected validation error")
	require.NotNil(t, res.Errors)
	assert.GreaterOrEqual(t, len(res.Errors.Errors), 1)
}
