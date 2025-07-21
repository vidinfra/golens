package filter_test

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/vidinfra/golens/filter"
)

func TestStructResponse_DefaultBehavior(t *testing.T) {
	fmt.Println("🧪 Testing Struct Response vs JSON Response")
	fmt.Println("============================================================")

	// Create some test query parameters with errors
	values := url.Values{}
	values.Set("filter[invalid_field][eq]", "test")
	values.Set("filter[name][invalid_op]", "john")
	values.Set("filter[age][gt]", "25")

	// Parse the filters
	parser := filter.NewParser(values)
	parseResult := parser.Parse()

	// Create validator with limited allowed fields
	validator := filter.NewValidator([]string{"name", "age"}, nil)

	// Create a result manually (simulating what happens in the builder)
	result := filter.NewResult(nil)

	// Manually validate and add errors to see the struct response
	for _, f := range parseResult.Filters {
		if validationErr := validator.ValidateFilter(f); validationErr != nil {
			result.AddError(validationErr)
		}
	}

	// Add parsing errors if any
	if parseResult.Errors.HasErrors() {
		result.AddErrors(parseResult.Errors.Errors...)
	}

	fmt.Println("\n🏗️  STRUCT RESPONSE (Default):")
	fmt.Println("--------------------------------")

	// Test struct access
	fmt.Printf("Success: %t\n", result.Success)
	fmt.Printf("Has Errors: %t\n", result.HasErrors())
	fmt.Printf("Error Count: %d\n", len(result.Errors.Errors))

	// Assertions
	if result.Success {
		t.Error("Expected Success to be false when there are errors")
	}

	if !result.HasErrors() {
		t.Error("Expected HasErrors to be true")
	}

	if len(result.Errors.Errors) == 0 {
		t.Error("Expected at least one error")
	}

	fmt.Println("\n📋 Individual Error Details:")
	for i, err := range result.Errors.Errors {
		fmt.Printf("\n  Error #%d:\n", i+1)
		fmt.Printf("    Type: %s\n", err.Type)
		fmt.Printf("    Message: %s\n", err.Message)
		fmt.Printf("    Field: %s\n", err.Field)
		fmt.Printf("    Operator: %s\n", err.Operator)
		fmt.Printf("    Code: %s\n", err.Code)
		fmt.Printf("    HTTP Status: %d\n", err.HTTPStatus)
		fmt.Printf("    Suggestions: %v\n", err.Suggestions)

		// Test each error has required fields
		if err.Type == "" {
			t.Errorf("Error #%d missing Type", i+1)
		}
		if err.Message == "" {
			t.Errorf("Error #%d missing Message", i+1)
		}
		if err.Code == "" {
			t.Errorf("Error #%d missing Code", i+1)
		}
	}

	fmt.Println("\n🔄 Easy Access Examples:")
	fmt.Println("-------------------------")

	// Test easy access examples
	firstError := result.GetFirstError()
	if firstError != nil {
		fmt.Printf("First Error Message: %s\n", firstError.Message)
		fmt.Printf("First Error Field: %s\n", firstError.Field)

		// Test first error access
		if firstError.Message == "" {
			t.Error("First error should have a message")
		}
	} else {
		t.Error("Expected to get a first error")
	}

	fmt.Println("\n📄 JSON RESPONSE (Optional):")
	fmt.Println("-----------------------------")

	// Test JSON conversion (optional)
	jsonResponse := result.ToJSONResponse()
	fmt.Printf("%+v\n", jsonResponse)

	// Test JSON response structure
	if jsonResponse["success"] != false {
		t.Error("JSON response should have success: false")
	}

	if jsonResponse["errors"] == nil {
		t.Error("JSON response should have errors field")
	}

	fmt.Println("\n✅ CONCLUSION:")
	fmt.Println("- Default: You get full struct access")
	fmt.Println("- Optional: Convert to JSON when needed")
	fmt.Println("- Developers have full control!")
}

func TestStructResponse_CustomErrorFormatting(t *testing.T) {
	fmt.Println("\n\n🎨 Testing Custom Error Formatting")
	fmt.Println("==================================")

	// Create a result with multiple error types
	result := filter.NewResult(nil)

	// Add different types of errors
	result.AddError(filter.NewFieldNotAllowedError("email", []string{"name", "age"}))
	result.AddError(filter.NewInvalidOperatorError("xyz"))
	result.AddError(filter.NewMissingValueError("status", "eq"))

	fmt.Println("\n🎯 Custom Format #1: Simple List")
	fmt.Println("--------------------------------")
	for i, err := range result.Errors.Errors {
		fmt.Printf("%d. %s (Field: %s)\n", i+1, err.Message, err.Field)
	}

	fmt.Println("\n🎯 Custom Format #2: By Error Type")
	fmt.Println("----------------------------------")
	for _, err := range result.Errors.Errors {
		switch err.Type {
		case filter.ErrorTypeValidation:
			fmt.Printf("⚠️  Validation: %s\n", err.Message)
		case filter.ErrorTypeParsing:
			fmt.Printf("🔧 Parsing: %s\n", err.Message)
		default:
			fmt.Printf("❓ Other: %s\n", err.Message)
		}
	}

	fmt.Println("\n🎯 Custom Format #3: With Suggestions")
	fmt.Println("------------------------------------")
	for _, err := range result.Errors.Errors {
		fmt.Printf("Problem: %s\n", err.Message)
		if len(err.Suggestions) > 0 {
			fmt.Printf("  → Try: %v\n", err.Suggestions)
		}
	}

	// Test that we have the expected number of errors
	if len(result.Errors.Errors) != 3 {
		t.Errorf("Expected 3 errors, got %d", len(result.Errors.Errors))
	}
}

func TestStructResponse_TranslationReady(t *testing.T) {
	fmt.Println("\n\n🌍 Testing Translation-Ready Data")
	fmt.Println("=================================")

	result := filter.NewResult(nil)
	result.AddError(filter.NewFieldNotAllowedError("email", []string{"name", "age"}))

	fmt.Println("\n📝 Translation Keys Available:")
	fmt.Println("-----------------------------")

	for _, err := range result.Errors.Errors {
		fmt.Printf("Error Code: %s (for translation key)\n", err.Code)
		fmt.Printf("Field: %s (for context)\n", err.Field)
		fmt.Printf("Type: %s (for categorization)\n", err.Type)
		fmt.Printf("English Message: %s\n", err.Message)

		// Test that translation-ready data is available
		if err.Code == "" {
			t.Error("Error should have a code for translation")
		}
	}

	fmt.Println("\n🔤 Example Translation Usage:")
	fmt.Println("----------------------------")

	// Simulate how a developer would use this for translation
	for _, err := range result.Errors.Errors {
		translationKey := fmt.Sprintf("error.%s", err.Code)
		fmt.Printf("Key: %s\n", translationKey)
		fmt.Printf("Context: {field: '%s', suggestions: %v}\n", err.Field, err.Suggestions)
	}
}

func TestStructResponse_FlexibleResponseFormats(t *testing.T) {
	fmt.Println("\n\n🔄 Testing Flexible Response Formats")
	fmt.Println("====================================")

	result := filter.NewResult(nil)
	result.AddError(filter.NewOperatorNotAllowedError("name", "xyz", []filter.Clause{filter.Equals, filter.Contains}))

	fmt.Println("\n📱 Mobile API Format:")
	fmt.Println("--------------------")
	mobileResponse := map[string]interface{}{
		"status":  "error",
		"code":    result.GetFirstError().Code,
		"message": result.GetFirstError().Message,
	}
	fmt.Printf("%+v\n", mobileResponse)

	fmt.Println("\n🌐 Web API Format:")
	fmt.Println("-----------------")
	webResponse := map[string]interface{}{
		"success": result.Success,
		"data":    nil,
		"errors": func() []map[string]string {
			errors := []map[string]string{}
			for _, err := range result.Errors.Errors {
				errors = append(errors, map[string]string{
					"field":   err.Field,
					"message": err.Message,
					"type":    string(err.Type),
				})
			}
			return errors
		}(),
	}
	fmt.Printf("%+v\n", webResponse)

	fmt.Println("\n📄 Standard JSON (using library method):")
	fmt.Println("---------------------------------------")
	standardJSON := result.ToJSONResponse()
	fmt.Printf("%+v\n", standardJSON)

	// Test that all formats contain the expected data
	if mobileResponse["code"] == "" {
		t.Error("Mobile format should have error code")
	}

	if len(webResponse["errors"].([]map[string]string)) == 0 {
		t.Error("Web format should have errors array")
	}

	if standardJSON["success"] != false {
		t.Error("Standard JSON should show success: false")
	}
}
