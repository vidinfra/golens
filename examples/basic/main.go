// Basic Example - Updated for Struct-First Error Handling (GORM)
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
	"gorm.io/gorm"

	// Choose your driver and init the DB accordingly:
	// _ "gorm.io/driver/postgres"
	// _ "gorm.io/driver/mysql"
	// _ "gorm.io/driver/sqlite"

	"github.com/vidinfra/golens/filter"
)

// User is your example model. Converted Bun tags -> GORM tags.
type User struct {
	Name   string `json:"name"   gorm:"column:name"`
	Email  string `json:"email"  gorm:"column:email"`
	Status string `json:"status" gorm:"column:status"`
	ID     int    `json:"id"     gorm:"column:id;primaryKey;autoIncrement"`
	Age    int    `json:"age"    gorm:"column:age"`
}

// Optional: if your table name is not the pluralized default
// func (User) TableName() string { return "users" }

func main() {
	r := gin.Default()

	// TODO: Initialize your *gorm.DB here (examples):
	//
	// Postgres:
	// dsn := "host=... user=... password=... dbname=... port=5432 sslmode=disable TimeZone=UTC"
	// db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	//
	// MySQL:
	// dsn := "user:pass@tcp(127.0.0.1:3306)/dbname?parseTime=true&charset=utf8mb4&loc=Local"
	// db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	//
	// SQLite:
	// db, err := gorm.Open(sqlite.Open("app.db"), &gorm.Config{})
	//
	// if err != nil { log.Fatal(err) }

	var db *gorm.DB // <- replace with your initialized DB

	r.GET("/users", func(c *gin.Context) {
		// Start from a model-backed query in GORM
		query := db.Model(&User{}) // This would be your actual GORM query

		// Create filter with struct-first error handling
		result := filter.New(c, query).
			AllowFields("name", "email", "age", "status").
			AllowSorts("name", "age", "created_at").
			Apply()

		// Struct-first error handling - work directly with error structs
		if result.OK() {
			// Option 1: Direct struct access for custom handling
			for _, err := range result.GetErrors().Errors {
				log.Printf("Filter error: %s (Field: %s, Code: %s)", err.Message, err.Field, err.Code)
			}

			// Option 2: Use built-in JSON helper for quick responses
			c.JSON(http.StatusBadRequest, result.Result().ToJSONResponse())
			return
		}

		// Get the final validated query (*gorm.DB)
		finalQuery := result.Query()

		_ = finalQuery

		// Execute your query here, e.g.:
		// var users []User
		// if err := finalQuery.Find(&users).Error; err != nil {
		//     c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "query failed"})
		//     return
		// }

		c.JSON(http.StatusOK, gin.H{
			"message": "Users filtered successfully",
			"success": true,
			// "data":    users,
		})
	})

	r.Run(":8080")
}
