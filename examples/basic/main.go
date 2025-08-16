// Basic Example - Struct-First Error Handling with GORM (SQLite in-memory)
package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vidinfra/golens/filter"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	Name   string `json:"name"   gorm:"column:name"`
	Email  string `json:"email"  gorm:"column:email"`
	Status string `json:"status" gorm:"column:status"`
	ID     int    `json:"id"     gorm:"column:id;primaryKey;autoIncrement"`
	Age    int    `json:"age"    gorm:"column:age"`
}

func main() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		log.Fatalf("open sqlite: %v", err)
	}

	// Auto-migrate schema
	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	// Seed a few rows
	seed := []User{
		{Name: "alice", Email: "a@example.com", Status: "active", Age: 20},
		{Name: "alina", Email: "b@example.com", Status: "active", Age: 22},
		{Name: "bob", Email: "c@example.com", Status: "inactive", Age: 17},
	}
	if err := db.Create(&seed).Error; err != nil {
		log.Fatalf("seed: %v", err)
	}

	// --- Gin setup ---
	r := gin.Default()

	// GET /users?filter[name][starts_with]=ali&filter[age][gte]=20&sort=-age
	r.GET("/users", func(c *gin.Context) {
		base := db.Model(&User{})

		builder := filter.New(c, base).
			AllowFields("name", "email", "age", "status").
			AllowSorts("name", "age", "status").AllowSorts("name", "age")
		builder.Apply()

		res := builder.Result()

		// Struct-first error handling
		if !res.OK() {
			// Log details or branch on codes/types if needed
			for _, e := range res.Errors.Errors {
				log.Printf("filter error: type=%s field=%s op=%s code=%s msg=%s",
					e.Type, e.Field, e.Operator, e.Code, e.Message)
			}
			// Quick JSON response
			c.JSON(http.StatusOK, res.ToJSONResponse())
			return
		}

		// Success path: use the validated *gorm.DB to query
		var users []User
		if err := res.Query.Find(&users).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"errors": []map[string]any{
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
			"count":   len(users),
			"data":    users,
		})
	})

	log.Println("listening on :8080")
	_ = r.Run(":8080")
}
