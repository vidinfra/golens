package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"github.com/vidinfra/golens/pkg/filter"
)

type User struct {
	ID     int    `json:"id" bun:"id,pk,autoincrement"`
	Name   string `json:"name" bun:"name"`
	Email  string `json:"email" bun:"email"`
	Age    int    `json:"age" bun:"age"`
	Status string `json:"status" bun:"status"`
}

func main() {
	r := gin.Default()

	r.GET("/users", func(c *gin.Context) {
		// Simulate a database query
		var query *bun.SelectQuery // This would be your actual bun query

		// Create filter with allowed fields
		result := filter.New(c, query).
			AllowFields("name", "email", "age", "status").
			AllowSorts("name", "age", "created_at").
			Apply().
			ApplySort().
			Query()

		// Execute your query here
		_ = result

		c.JSON(http.StatusOK, gin.H{"message": "Users filtered successfully"})
	})

	r.Run(":8080")
}
