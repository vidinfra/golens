package main

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/vidinfra/golens/pkg/filter"
)

type Product struct {
	ID          int     `json:"id" bun:"id,pk,autoincrement"`
	Name        string  `json:"name" bun:"name"`
	Description string  `json:"description" bun:"description"`
	Price       float64 `json:"price" bun:"price"`
	Category    string  `json:"category" bun:"category"`
	Status      string  `json:"status" bun:"status"`
	CreatedAt   string  `json:"created_at" bun:"created_at"`
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

		// Apply filters and sorting
		finalQuery := filter.New(c, query).
			AllowConfigs(configs...).
			AllowSorts("name", "price", "created_at", "category").
			Apply().
			ApplySort().
			Query()

		// Execute the query
		var products []Product
		err := finalQuery.Scan(c.Request.Context(), &products)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  products,
			"count": len(products),
		})
	})

	// Example with simple field allowlist
	r.GET("/users", func(c *gin.Context) {
		query := db.NewSelect().Model((*User)(nil))

		finalQuery := filter.New(c, query).
			AllowAll("name", "email", "status", "created_at").
			Apply().
			ApplySort().
			Query()

		var users []User
		err := finalQuery.Scan(c.Request.Context(), &users)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
			return
		}

		c.JSON(http.StatusOK, users)
	})

	r.Run(":8080")
}

type User struct {
	ID        int    `json:"id" bun:"id,pk,autoincrement"`
	Name      string `json:"name" bun:"name"`
	Email     string `json:"email" bun:"email"`
	Status    string `json:"status" bun:"status"`
	CreatedAt string `json:"created_at" bun:"created_at"`
}
