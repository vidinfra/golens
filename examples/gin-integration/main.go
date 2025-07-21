package main

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/vidinfra/golens/filter"
)

type Product struct {
	Name        string  `json:"name" bun:"name"`
	Description string  `json:"description" bun:"description"`
	Category    string  `json:"category" bun:"category"`
	Status      string  `json:"status" bun:"status"`
	CreatedAt   string  `json:"created_at" bun:"created_at"`
	ID          int     `json:"id" bun:"id,pk,autoincrement"`
	Price       float64 `json:"price" bun:"price"`
}

func main() {
	// Setup database connection (example)
	dsn := "postgres://username:password@localhost:5432/database?sslmode=disable"
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())

	r := gin.Default()

	// Example endpoint with comprehensive filtering
	r.GET("/products", func(c *gin.Context) {
		// Start with base query
		query := db.NewSelect().Model((*Product)(nil))

		// Define filter configurations
		configs := []filter.FilterConfig{
			filter.AllowedFilter("name", filter.Equals, filter.Contains, filter.StartsWith, filter.EndsWith),
			filter.AllowedFilter("description", filter.Contains),
			filter.AllowedFilter("price", filter.Equals, filter.GreaterThan, filter.LessThan, filter.GreaterThanOrEq, filter.LessThanOrEq, filter.Between),
			filter.AllowedFilter("category", filter.Equals, filter.In, filter.NotIn),
			filter.AllowedFilter("status", filter.Equals, filter.In),
			filter.AllowedFilter("created_at", filter.GreaterThan, filter.LessThan, filter.Between),
		}

		// Apply filters and sorting with error handling
		builder := filter.New(c, query).
			AllowConfigs(configs...).
			AllowSorts("name", "price", "created_at", "category").
			Apply().
			ApplySort()

		// Check for filter/sort errors
		if builder.HasErrors() {
			c.JSON(http.StatusBadRequest, builder.Result().ToJSONResponse())
			return
		}

		finalQuery := builder.Query()

		// Execute the query
		var products []Product
		err := finalQuery.Scan(c.Request.Context(), &products)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"errors": []map[string]interface{}{
					{
						"type":    "database_error",
						"message": "Failed to fetch products",
						"code":    "DATABASE_ERROR",
					},
				},
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    products,
			"count":   len(products),
		})
	})

	// Example with simple field allowlist
	r.GET("/users", func(c *gin.Context) {
		query := db.NewSelect().Model((*User)(nil))

		builder := filter.New(c, query).
			AllowAll("name", "email", "status", "created_at").
			Apply().
			ApplySort()

		// Check for errors
		if builder.HasErrors() {
			c.JSON(http.StatusBadRequest, builder.Result().ToJSONResponse())
			return
		}

		finalQuery := builder.Query()

		var users []User
		err := finalQuery.Scan(c.Request.Context(), &users)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"errors": []map[string]interface{}{
					{
						"type":    "database_error",
						"message": "Failed to fetch users",
						"code":    "DATABASE_ERROR",
					},
				},
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    users,
		})
	})

	r.Run(":8080")
}

type User struct {
	Name      string `json:"name" bun:"name"`
	Email     string `json:"email" bun:"email"`
	Status    string `json:"status" bun:"status"`
	CreatedAt string `json:"created_at" bun:"created_at"`
	ID        int    `json:"id" bun:"id,pk,autoincrement"`
}
