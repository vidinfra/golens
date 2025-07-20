package filter_test

// import (
// 	"database/sql"
// 	"net/http"
// 	"net/http/httptest"
// 	"strings"
// 	"testing"

// 	"github.com/gin-gonic/gin"
// 	"github.com/uptrace/bun"
// 	"github.com/uptrace/bun/dialect/pgdialect"
// 	"github.com/uptrace/bun/driver/pgdriver"

// 	"github.com/vidinfra/golens/filter"
// )

// type User struct{}

// func createTestContext(queryParams string) *gin.Context {
// 	gin.SetMode(gin.TestMode)
// 	req := httptest.NewRequest(http.MethodGet, "/?"+queryParams, nil)
// 	w := httptest.NewRecorder()
// 	c, _ := gin.CreateTestContext(w)
// 	c.Request = req
// 	return c
// }

// func createBunQuery() *bun.SelectQuery {
// 	dsn := "postgres://user:pass@localhost:5432/testdb" // Dummy, not actually used
// 	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
// 	db := bun.NewDB(sqldb, pgdialect.New())
// 	return db.NewSelect().Model((*User)(nil)).Table("users")
// }

// func TestBuilder_GeneratesSQL(t *testing.T) {
// 	tests := []struct {
// 		name          string
// 		queryParams   string
// 		allowedFields []string
// 		expectedSQL   string
// 	}{
// 		{
// 			name:          "simple equals",
// 			queryParams:   "filter[name][eq]=john",
// 			allowedFields: []string{"name"},
// 			expectedSQL:   `SELECT * FROM "users" WHERE "name" = 'john'`,
// 		},
// 		{
// 			name:          "like operator",
// 			queryParams:   "filter[name][like]=jo",
// 			allowedFields: []string{"name"},
// 			expectedSQL:   `SELECT * FROM "users" WHERE "name" ILIKE '%jo%'`,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			c := createTestContext(tt.queryParams)
// 			query := createBunQuery()

// 			builder := filter.New(c, query).AllowFields(tt.allowedFields...)
// 			builder.Apply()

// 			sql := query.String()
// 			normalized := strings.TrimSpace(sql)
// 			if normalized != tt.expectedSQL {
// 				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expectedSQL, normalized)
// 			}
// 		})
// 	}
// }
