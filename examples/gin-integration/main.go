package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/vidinfra/golens/filter"
)

type Product struct {
	Name        string  `json:"name"        gorm:"column:name"`
	Description string  `json:"description" gorm:"column:description"`
	Category    string  `json:"category"    gorm:"column:category"`
	Status      string  `json:"status"      gorm:"column:status"`
	CreatedAt   string  `json:"created_at"  gorm:"column:created_at"`
	ID          int     `json:"id"          gorm:"column:id;primaryKey;autoIncrement"`
	Price       float64 `json:"price"       gorm:"column:price"`
}

// If your table name isn't the default pluralization, uncomment:
// func (Product) TableName() string { return "products" }

type User struct {
	Name      string `json:"name"       gorm:"column:name"`
	Email     string `json:"email"      gorm:"column:email"`
	Status    string `json:"status"     gorm:"column:status"`
	CreatedAt string `json:"created_at" gorm:"column:created_at"`
	ID        int    `json:"id"         gorm:"column:id;primaryKey;autoIncrement"`
}

// func (User) TableName() string { return "users" }

// -------------------------------------

func main() {
	r := gin.Default()

	// ----- GORM DB Init (Postgres example) -----
	// Replace with your own DSN/driver as needed
	dsn := "host=localhost user=username password=password dbname=database port=5432 sslmode=disable TimeZone=UTC"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	// -------------------------------------------

	// Example endpoint with comprehensive filtering
	r.GET("/products", func(c *gin.Context) {
		// Start with base GORM query
		query := db.Model(&Product{})

		// Define filter configurations (same semantics as your Bun example)
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
			Apply()

		// Struct-first error handling with direct access to error details
		if result.OK() {
			errors := result.GetErrors()

			// Log all errors for debugging
			for _, err := range errors.Errors {
				log.Printf("Filter error: %s (Field: %s, Code: %s, Type: %s)",
					err.Message, err.Field, err.Code, err.Type)
			}

			// Quick response via helper
			c.JSON(http.StatusBadRequest, result.Result().ToJSONResponse())
			return
		}

		finalQuery := result.Query() // *gorm.DB

		// Execute the query
		var products []Product
		if err := finalQuery.Find(&products).Error; err != nil {
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
		query := db.Model(&User{})

		// If your library has AllowAll, use it; otherwise do AllowFields+AllowSorts explicitly:
		result := filter.New(c, query).
			AllowFields("name", "email", "status", "created_at").
			AllowSorts("name", "created_at").
			Apply()

		// Direct struct access for custom error handling
		if result.OK() {
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

		finalQuery := result.Query() // *gorm.DB

		var users []User
		if err := finalQuery.Find(&users).Error; err != nil {
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
