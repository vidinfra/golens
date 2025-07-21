package filter_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/vidinfra/golens/filter"
)

func TestResult_ToJSONResponse_ActualJSON(t *testing.T) {
	tests := []struct {
		setup func() *filter.Result // 8 bytes (pointer to function)
		name  string                // 16 bytes (string)
	}{
		{
			name: "successful result JSON",
			setup: func() *filter.Result {
				return filter.NewResult(nil)
			},
		},
		{
			name: "result with single error JSON",
			setup: func() *filter.Result {
				result := filter.NewResult(nil)
				result.AddError(filter.NewFieldNotAllowedError("email", []string{"name", "age"}))
				return result
			},
		},
		{
			name: "result with multiple errors JSON",
			setup: func() *filter.Result {
				result := filter.NewResult(nil)
				result.AddError(filter.NewFieldNotAllowedError("email", []string{"name", "age"}))
				result.AddError(filter.NewInvalidOperatorError("xyz"))
				return result
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.setup()

			// Test 1: Can marshal to JSON successfully
			jsonBytes, err := json.Marshal(result.ToJSONResponse())
			if err != nil {
				t.Fatalf("Failed to marshal to JSON: %v", err)
			}

			// Test 2: JSON is not empty
			if len(jsonBytes) == 0 {
				t.Error("JSON output is empty")
			}

			// Test 3: Can unmarshal back from JSON (validates JSON format)
			var parsed map[string]interface{}
			if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			// Test 4: Has required fields
			if _, exists := parsed["success"]; !exists {
				t.Error("JSON missing required 'success' field")
			}

			// Test 5: Verify structure makes sense
			if hasErrors := result.HasErrors(); hasErrors {
				if _, exists := parsed["errors"]; !exists {
					t.Error("JSON missing 'errors' field when result has errors")
				}
				if success, ok := parsed["success"].(bool); !ok || success {
					t.Error("JSON 'success' should be false when there are errors")
				}
			} else {
				if success, ok := parsed["success"].(bool); !ok || !success {
					t.Error("JSON 'success' should be true when there are no errors")
				}
			}

			// Log the actual JSON for inspection
			t.Logf("Produced valid JSON: %s", string(jsonBytes))
		})
	}
}

func TestResult_ToJSONResponse_ValidJSON(t *testing.T) {
	tests := []struct {
		setup func() *filter.Result // 8 bytes (pointer to function)
		name  string                // 16 bytes (string)
	}{
		{
			name: "successful result produces valid JSON",
			setup: func() *filter.Result {
				return filter.NewResult(nil)
			},
		},
		{
			name: "error result produces valid JSON",
			setup: func() *filter.Result {
				result := filter.NewResult(nil)
				result.AddError(filter.NewFieldNotAllowedError("email", []string{"name", "age"}))
				result.AddError(filter.NewInvalidOperatorError("xyz"))
				return result
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.setup()

			// Test 1: Can marshal to JSON
			jsonBytes, err := json.Marshal(result.ToJSONResponse())
			if err != nil {
				t.Fatalf("Cannot marshal to JSON: %v", err)
			}

			// Test 2: JSON is not empty
			if len(jsonBytes) == 0 {
				t.Error("JSON output is empty")
			}

			// Test 3: Can unmarshal back from JSON
			var parsed map[string]interface{}
			if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
				t.Fatalf("Cannot parse JSON: %v (JSON: %s)", err, string(jsonBytes))
			}

			// Test 4: Parsed data has expected structure
			if _, exists := parsed["success"]; !exists {
				t.Error("Parsed JSON missing 'success' field")
			}

			// Print the actual JSON for verification
			t.Logf("Produced JSON: %s", string(jsonBytes))
		})
	}
}

func TestResult_ToJSONResponse_JSONStructure(t *testing.T) {
	tests := []struct {
		name           string
		setup          func() *filter.Result
		expectedFields []string
		mustHaveErrors bool
	}{
		{
			name: "successful result JSON structure",
			setup: func() *filter.Result {
				return filter.NewResult(nil)
			},
			expectedFields: []string{"success"},
			mustHaveErrors: false,
		},
		{
			name: "error result JSON structure",
			setup: func() *filter.Result {
				result := filter.NewResult(nil)
				result.AddError(filter.NewFieldNotAllowedError("email", []string{"name", "age"}))
				return result
			},
			expectedFields: []string{"success", "errors"},
			mustHaveErrors: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.setup()

			// Convert to JSON
			jsonBytes, err := json.Marshal(result.ToJSONResponse())
			if err != nil {
				t.Fatalf("JSON marshal failed: %v", err)
			}

			// Parse back to verify structure
			var parsed map[string]interface{}
			if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
				t.Fatalf("JSON unmarshal failed: %v", err)
			}

			// Check required fields exist
			for _, field := range tt.expectedFields {
				if _, exists := parsed[field]; !exists {
					t.Errorf("JSON missing required field: %s", field)
				}
			}

			// Check errors array if expected
			if tt.mustHaveErrors {
				if errorsField, exists := parsed["errors"]; !exists {
					t.Error("JSON missing 'errors' field")
				} else if errorsArray, ok := errorsField.([]interface{}); !ok {
					t.Error("'errors' field is not an array")
				} else if len(errorsArray) == 0 {
					t.Error("'errors' array is empty but should have errors")
				} else {
					// Check first error has required fields
					firstError, ok := errorsArray[0].(map[string]interface{})
					if !ok {
						t.Error("First error is not an object")
					} else {
						requiredErrorFields := []string{"type", "message", "code"}
						for _, field := range requiredErrorFields {
							if _, exists := firstError[field]; !exists {
								t.Errorf("Error object missing field: %s", field)
							}
						}
					}
				}
			}
		})
	}
}

func TestResult_ToJSONResponse_PrettyJSON(t *testing.T) {
	// Test that produces nicely formatted JSON for debugging
	result := filter.NewResult(nil)
	result.AddError(filter.NewFieldNotAllowedError("email", []string{"name", "age"}))
	result.AddError(filter.NewMissingValueError("status", "eq"))

	// Get the response
	response := result.ToJSONResponse()

	// Test regular JSON
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	// Test pretty JSON
	prettyBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		t.Fatalf("Pretty JSON marshal failed: %v", err)
	}

	t.Logf("Compact JSON: %s", string(jsonBytes))
	t.Logf("Pretty JSON:\n%s", string(prettyBytes))

	// Verify both are valid
	var compact, pretty map[string]interface{}

	if err := json.Unmarshal(jsonBytes, &compact); err != nil {
		t.Errorf("Compact JSON invalid: %v", err)
	}

	if err := json.Unmarshal(prettyBytes, &pretty); err != nil {
		t.Errorf("Pretty JSON invalid: %v", err)
	}

	// Verify they contain the same data
	if !reflect.DeepEqual(compact, pretty) {
		t.Error("Compact and pretty JSON contain different data")
	}
}

func TestFilterError_ToJSONResponse_ActualJSON(t *testing.T) {
	tests := []struct {
		error *filter.FilterError // 8 bytes (pointer)
		name  string              // 16 bytes (string)
	}{
		{
			name:  "field not allowed error JSON",
			error: filter.NewFieldNotAllowedError("email", []string{"name", "age"}),
		},
		{
			name:  "invalid operator error JSON",
			error: filter.NewInvalidOperatorError("xyz"),
		},
		{
			name:  "missing value error JSON",
			error: filter.NewMissingValueError("status", "eq"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test 1: Can marshal to JSON successfully
			jsonBytes, err := json.Marshal(tt.error.ToJSONResponse())
			if err != nil {
				t.Fatalf("Failed to marshal FilterError to JSON: %v", err)
			}

			// Test 2: JSON is not empty
			if len(jsonBytes) == 0 {
				t.Error("JSON output is empty")
			}

			// Test 3: Can unmarshal back from JSON (validates JSON format)
			var parsed map[string]interface{}
			if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			// Test 4: Has required structure
			if _, exists := parsed["error"]; !exists {
				t.Error("JSON missing required 'error' field")
			}

			// Test 5: Error object has required fields
			if errorObj, ok := parsed["error"].(map[string]interface{}); ok {
				requiredFields := []string{"type", "message", "code"}
				for _, field := range requiredFields {
					if _, exists := errorObj[field]; !exists {
						t.Errorf("Error object missing required field: %s", field)
					}
				}

				// Test specific values
				if errorObj["type"] != "validation_error" {
					t.Errorf("Expected type 'validation_error', got %v", errorObj["type"])
				}
				if errorObj["code"] != "FILTER_VALIDATION_ERROR" {
					t.Errorf("Expected code 'FILTER_VALIDATION_ERROR', got %v", errorObj["code"])
				}
			} else {
				t.Error("'error' field is not an object")
			}

			// Log the actual JSON for inspection
			t.Logf("Produced valid JSON: %s", string(jsonBytes))
		})
	}
}

func TestResult_ToJSONResponse_JSONStringValidation(t *testing.T) {
	tests := []struct {
		name           string
		setup          func() *filter.Result
		mustContain    []string
		mustNotContain []string
	}{
		{
			name: "successful result JSON string",
			setup: func() *filter.Result {
				return filter.NewResult(nil)
			},
			mustContain:    []string{`"success":true`},
			mustNotContain: []string{`"errors"`, `"false"`},
		},
		{
			name: "error result JSON string",
			setup: func() *filter.Result {
				result := filter.NewResult(nil)
				result.AddError(filter.NewFieldNotAllowedError("email", []string{"name", "age"}))
				return result
			},
			mustContain: []string{
				`"success":false`,
				`"errors":[`,
				`"type":"validation_error"`,
				`"field":"email"`,
				`"code":"FILTER_VALIDATION_ERROR"`,
			},
			mustNotContain: []string{`"success":true`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.setup()

			// Convert to JSON string
			jsonBytes, err := json.Marshal(result.ToJSONResponse())
			if err != nil {
				t.Fatalf("Failed to marshal to JSON: %v", err)
			}

			jsonString := string(jsonBytes)
			t.Logf("Produced JSON: %s", jsonString)

			// Check required substrings
			for _, required := range tt.mustContain {
				if !strings.Contains(jsonString, required) {
					t.Errorf("JSON missing required substring: %s\nJSON: %s", required, jsonString)
				}
			}

			// Check forbidden substrings
			for _, forbidden := range tt.mustNotContain {
				if strings.Contains(jsonString, forbidden) {
					t.Errorf("JSON contains forbidden substring: %s\nJSON: %s", forbidden, jsonString)
				}
			}

			// Verify it's valid JSON
			var parsed map[string]interface{}
			if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
				t.Errorf("Produced invalid JSON: %v", err)
			}
		})
	}
}
