// Gin Integration Example - Updated for Struct-First Error Handling
//
// Production-ready example showing:
// 1. Real Gin + Bun ORM integration
// 2. Comprehensive filtering with field configurations
// 3. Advanced error handling with logging
// 4. Field-specific permission logic
// 5. Both simple and complex filtering endpoints
package main

import (
	"database/sql"
	"log"
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

		// Apply filters and sorting with struct-first error handling
		result := filter.New(c, query).
			AllowConfigs(configs...).
			AllowSorts("name", "price", "created_at", "category").
			Apply().
			ApplySort()

		// Struct-first error handling with direct access to error details
		if result.HasErrors() {
			errors := result.GetErrors()

			// Log all errors for debugging
			for _, err := range errors.Errors {
				log.Printf("Filter error: %s (Field: %s, Code: %s, Type: %s)",
					err.Message, err.Field, err.Code, err.Type)
			}

			// Option 1: Use built-in JSON conversion for quick response
			c.JSON(http.StatusBadRequest, result.Result().ToJSONResponse())
			return
		}

		finalQuery := result.Query()

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

	// Example with simple field allowlist and struct-first error handling
	r.GET("/users", func(c *gin.Context) {
		query := db.NewSelect().Model((*User)(nil))

		result := filter.New(c, query).
			AllowAll("name", "email", "status", "created_at").
			Apply().
			ApplySort()

		// Direct struct access for custom error handling
		if result.HasErrors() {
			errors := result.GetErrors()

			// Custom logic based on error types
			for _, err := range errors.Errors {
				if err.Field == "email" {
					// Special handling for sensitive email field
					c.JSON(http.StatusForbidden, gin.H{
						"error": "Email filtering requires special permissions",
						"code":  "PERMISSION_DENIED",
					})
					return
				}
			}

			// Use library's JSON helper for other errors
			c.JSON(http.StatusBadRequest, result.Result().ToJSONResponse())
			return
		}

		finalQuery := result.Query()

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
