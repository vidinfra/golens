package main

import (
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

		// Create filter with allowed fields and error handling
		builder := filter.New(c, query).
			AllowFields("name", "email", "age", "status").
			AllowSorts("name", "age", "created_at").
			Apply().
			ApplySort()

		// Check for errors
		if builder.HasErrors() {
			c.JSON(http.StatusBadRequest, builder.Result().ToJSONResponse())
			return
		}

		// Get the final query
		finalQuery := builder.Query()

		// Execute your query here
		_ = finalQuery

		c.JSON(http.StatusOK, gin.H{
			"message": "Users filtered successfully",
			"success": true,
		})
	})

	r.Run(":8080")
}
