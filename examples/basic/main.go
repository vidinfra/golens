// Basic Example - Updated for Struct-First Error Handling
//
// This example demonstrates the core struct-first approach:
// 1. Direct access to error structs for full control
// 2. Optional JSON conversion helpers
// 3. Clean separation between validation and response formatting
package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"github.com/vidinfra/golens/filter"
)

type User struct {
	Name   string `json:"name" bun:"name"`
	Email  string `json:"email" bun:"email"`
	Status string `json:"status" bun:"status"`
	ID     int    `json:"id" bun:"id,pk,autoincrement"`
	Age    int    `json:"age" bun:"age"`
}

func main() {
	r := gin.Default()

	r.GET("/users", func(c *gin.Context) {
		// Simulate a database query
		var query *bun.SelectQuery // This would be your actual bun query

		// Create filter with struct-first error handling
		result := filter.New(c, query).
			AllowFields("name", "email", "age", "status").
			AllowSorts("name", "age", "created_at").
			Apply().
			ApplySort()

		// Struct-first error handling - work directly with error structs
		if result.HasErrors() {
			// Option 1: Direct struct access for custom handling
			for _, err := range result.GetErrors().Errors {
				log.Printf("Filter error: %s (Field: %s, Code: %s)",
					err.Message, err.Field, err.Code)
			}

			// Option 2: Use built-in JSON helper for quick responses
			c.JSON(http.StatusBadRequest, result.Result().ToJSONResponse())
			return
		}

		// Get the final validated query
		finalQuery := result.Query()

		// Execute your query here
		_ = finalQuery

		c.JSON(http.StatusOK, gin.H{
			"message": "Users filtered successfully",
			"success": true,
		})
	})

	r.Run(":8080")
}
