// Custom Error Handling Example - Updated for Struct-First Approach
//
// This example shows advanced patterns:
// 1. Error categorization by type and field
// 2. Custom response formatting
// 3. Field-specific logic (e.g., email permissions)
// 4. Translation/internationalization ready patterns
package main

import (
	"fmt"
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

		// Advanced struct-first error handling with custom logic
		if result.HasErrors() {
			errors := result.GetErrors()

			// Example 1: Custom error categorization
			validationErrors := []string{}
			parsingErrors := []string{}

			for _, err := range errors.Errors {
				switch err.Type {
				case "validation_error":
					validationErrors = append(validationErrors, err.Message)
				case "parsing_error":
					parsingErrors = append(parsingErrors, err.Message)
				}
			}

			// Example 2: Field-specific error handling
			for _, err := range errors.Errors {
				if err.Field == "email" {
					log.Printf("Special handling for email field error: %s", err.Message)
					c.JSON(http.StatusForbidden, gin.H{
						"error": "Email filtering requires admin privileges",
						"code":  "ADMIN_REQUIRED",
					})
					return
				}
			}

			// Example 3: Custom error response format
			customResponse := map[string]interface{}{
				"status": "error",
				"errors": map[string]interface{}{
					"validation": validationErrors,
					"parsing":    parsingErrors,
				},
				"suggestions": "Check field names and operator usage",
			}

			// Example 4: Add error details for debugging (if needed)
			if len(errors.Errors) > 0 {
				firstError := errors.Errors[0]
				customResponse["debug"] = map[string]interface{}{
					"first_error_field":    firstError.Field,
					"first_error_operator": firstError.Operator,
					"first_error_code":     firstError.Code,
					"suggestions":          firstError.Suggestions,
				}
			}

			c.JSON(http.StatusBadRequest, customResponse)
			return
		}

		// Success path - use the validated query
		finalQuery := result.Query()
		_ = finalQuery

		c.JSON(http.StatusOK, gin.H{
			"message": "Users filtered successfully with custom error handling",
			"success": true,
		})
	})

	// Example with internationalization/translation ready
	r.GET("/users/i18n", func(c *gin.Context) {
		var query *bun.SelectQuery

		result := filter.New(c, query).
			AllowFields("name", "email").
			Apply()

		if result.HasErrors() {
			translatedErrors := translateErrors(result.GetErrors(), "en")
			c.JSON(http.StatusBadRequest, translatedErrors)
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	r.Run(":8080")
}

// Example translation function for internationalization
func translateErrors(errors *filter.FilterErrors, locale string) map[string]interface{} {
	translated := make([]map[string]interface{}, 0, len(errors.Errors))

	for _, err := range errors.Errors {
		translatedError := map[string]interface{}{
			"field": err.Field,
			"code":  err.Code,
		}

		// Translate based on error code and locale
		switch err.Code {
		case "FILTER_VALIDATION_ERROR":
			if err.Field != "" {
				translatedError["message"] = fmt.Sprintf("Field '%s' is not allowed", err.Field)
			} else if err.Operator != "" {
				translatedError["message"] = fmt.Sprintf("Operator '%s' is not supported", err.Operator)
			}
		default:
			translatedError["message"] = err.Message
		}

		if len(err.Suggestions) > 0 {
			translatedError["suggestions"] = err.Suggestions
		}

		translated = append(translated, translatedError)
	}

	return map[string]interface{}{
		"success": false,
		"errors":  translated,
		"locale":  locale,
	}
}
