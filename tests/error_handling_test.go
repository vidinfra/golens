package filter_test

import (
	"net/url"
	"testing"

	"github.com/vidinfra/golens/filter"
)

func TestErrorHandling_InvalidField(t *testing.T) {
	// Parse a query with an invalid field
	values, _ := url.ParseQuery("filter[invalid_field][eq]=value")
	parser := filter.NewParser(values)
	result := parser.Parse()

	// Create validator with allowed fields
	validator := filter.NewValidator([]string{"name", "age"}, nil)
	applier := filter.NewApplier(validator)

	// This should produce validation errors
	_, err := applier.ApplyFilters(nil, result.Filters)

	if err == nil {
		t.Fatal("Expected error for invalid field, got nil")
	}

	filterErrors, ok := err.(*filter.FilterErrors)
	if !ok {
		t.Fatalf("Expected FilterErrors, got %T", err)
	}

	if !filterErrors.HasErrors() {
		t.Fatal("Expected errors in FilterErrors")
	}

	if len(filterErrors.Errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(filterErrors.Errors))
	}

	firstError := filterErrors.Errors[0]
	if firstError.Type != filter.ErrorTypeValidation {
		t.Errorf("Expected validation error, got %s", firstError.Type)
	}

	if firstError.Field != "invalid_field" {
		t.Errorf("Expected field 'invalid_field', got '%s'", firstError.Field)
	}
}

func TestErrorHandling_InvalidOperator(t *testing.T) {
	// Parse a query with an invalid operator
	values, _ := url.ParseQuery("filter[name][invalid_op]=value")
	parser := filter.NewParser(values)
	result := parser.Parse()

	if !result.Errors.HasErrors() {
		t.Fatal("Expected parsing errors for invalid operator")
	}

	if len(result.Errors.Errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(result.Errors.Errors))
	}

	firstError := result.Errors.Errors[0]
	if firstError.Type != filter.ErrorTypeValidation {
		t.Errorf("Expected validation error, got %s", firstError.Type)
	}
}

func TestErrorHandling_JSONResponse(t *testing.T) {
	// Create an error
	err := filter.NewFieldNotAllowedError("email", []string{"name", "age"})

	// Test JSON response
	jsonResponse := err.ToJSONResponse()

	errorData, ok := jsonResponse["error"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected error object in JSON response")
	}

	if errorData["type"] != string(filter.ErrorTypeValidation) {
		t.Errorf("Expected type '%s', got %v", filter.ErrorTypeValidation, errorData["type"])
	}

	if errorData["field"] != "email" {
		t.Errorf("Expected field 'email', got %v", errorData["field"])
	}

	suggestions, ok := errorData["suggestions"].([]string)
	if !ok {
		t.Fatal("Expected suggestions array")
	}

	if len(suggestions) != 2 {
		t.Errorf("Expected 2 suggestions, got %d", len(suggestions))
	}
}
