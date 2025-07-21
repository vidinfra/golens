package filter_test

import (
	"fmt"
	"net/url"
	"reflect"
	"testing"

	"github.com/vidinfra/golens/filter"
)

func TestResult_HasErrors(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *filter.Result
		expected bool
	}{
		{
			name: "no errors",
			setup: func() *filter.Result {
				return filter.NewResult(nil)
			},
			expected: false,
		},
		{
			name: "single error",
			setup: func() *filter.Result {
				result := filter.NewResult(nil)
				result.AddError(filter.NewValidationError("field", "op", "val", "test message"))
				return result
			},
			expected: true,
		},
		{
			name: "multiple error",
			setup: func() *filter.Result {
				result := filter.NewResult(nil)
				result.AddError(filter.NewFieldNotAllowedError("invalid_field", []string{"name", "age"}))
				result.AddError(filter.NewInvalidOperatorError("invalid_op"))
				return result
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.setup()
			if got := result.HasErrors(); got != tt.expected {
				t.Errorf("HasError() = %v, want %v", got, tt.expected)
			}
		})
	}

}

func TestResult_GetFirstError(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *filter.Result
		expected *filter.FilterError
		isNil    bool
	}{
		{
			name: "no errors",
			setup: func() *filter.Result {
				return filter.NewResult(nil)
			},
			expected: nil,
			isNil:    true,
		},
		{
			name: "single error",
			setup: func() *filter.Result {
				result := filter.NewResult(nil)
				err := filter.NewValidationError("test_field", "eq", "value", "test message")
				result.AddError(err)
				return result
			},
			expected: filter.NewValidationError("test_field", "eq", "value", "test message"),
			isNil:    false,
		},
		{
			name: "multiple errors returns first",
			setup: func() *filter.Result {
				result := filter.NewResult(nil)
				first := filter.NewFieldNotAllowedError("email", []string{"name", "age"})
				second := filter.NewInvalidOperatorError("xyz")
				result.AddError(first)
				result.AddError(second)
				return result
			},
			expected: filter.NewFieldNotAllowedError("email", []string{"name", "age"}),
			isNil:    false,
		},
	}

	for _, tt := range tests {

		fmt.Println(tt)
		t.Run(tt.name, func(t *testing.T) {
			result := tt.setup()
			got := result.GetFirstError()

			if tt.isNil {
				if got != nil {
					t.Error("GetFirstError() = nil, want error")
					return
				}
			} else {
				if got.Field != tt.expected.Field {
					t.Errorf("GetFirstError().Field = %v, want %v", got.Field, tt.expected.Field)
				}
				if got.Message != tt.expected.Message {
					t.Errorf("GetFirstError().Message = %v, want %v", got.Message, tt.expected.Message)
				}
				if got.Type != tt.expected.Type {
					t.Errorf("GetFirstError().Type = %v, want %v", got.Type, tt.expected.Type)
				}
			}
		})
	}
}

func TestResult_SuccessFlag(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *filter.Result
		expected bool
	}{
		{
			name: "new result is successful",
			setup: func() *filter.Result {
				return filter.NewResult(nil)
			},
			expected: true,
		},
		{
			name: "result with error is not successful",
			setup: func() *filter.Result {
				result := filter.NewResult(nil)
				result.AddError(filter.NewValidationError("field", "op", "val", "message"))
				return result
			},
			expected: false,
		},
		{
			name: "error result is not successful",
			setup: func() *filter.Result {
				return filter.NewErrorResult(filter.NewValidationError("field", "op", "val", "message"))
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.setup()
			if got := result.Success; got != tt.expected {
				t.Errorf("Success = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestResult_ErrorCount(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *filter.Result
		expected int
	}{
		{
			name: "no errors",
			setup: func() *filter.Result {
				return filter.NewResult(nil)
			},
			expected: 0,
		},
		{
			name: "single error",
			setup: func() *filter.Result {
				result := filter.NewResult(nil)
				result.AddError(filter.NewValidationError("field", "op", "val", "message"))
				return result
			},
			expected: 1,
		},
		{
			name: "multiple errors",
			setup: func() *filter.Result {
				result := filter.NewResult(nil)
				result.AddError(filter.NewFieldNotAllowedError("email", []string{"name", "age"}))
				result.AddError(filter.NewInvalidOperatorError("xyz"))
				result.AddError(filter.NewMissingValueError("status", "eq"))
				return result
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.setup()
			if got := len(result.Errors.Errors); got != tt.expected {
				t.Errorf("Error count = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFilterError_Fields(t *testing.T) {
	tests := []struct {
		name     string
		error    *filter.FilterError
		expected struct {
			Type        filter.ErrorType
			Message     string
			Field       string
			Operator    string
			Code        string
			HTTPStatus  int
			Suggestions []string
		}
	}{
		{
			name:  "field not allowed error",
			error: filter.NewFieldNotAllowedError("email", []string{"name", "age"}),
			expected: struct {
				Type        filter.ErrorType
				Message     string
				Field       string
				Operator    string
				Code        string
				HTTPStatus  int
				Suggestions []string
			}{
				Type:        filter.ErrorTypeValidation,
				Message:     "Field 'email' is not allowed for filtering",
				Field:       "email",
				Operator:    "",
				Code:        "FILTER_VALIDATION_ERROR",
				HTTPStatus:  400,
				Suggestions: []string{"name", "age"},
			},
		},
		{
			name:  "invalid operator error",
			error: filter.NewInvalidOperatorError("xyz"),
			expected: struct {
				Type        filter.ErrorType
				Message     string
				Field       string
				Operator    string
				Code        string
				HTTPStatus  int
				Suggestions []string
			}{
				Type:       filter.ErrorTypeValidation,
				Message:    "Invalid operator 'xyz'",
				Field:      "",
				Operator:   "xyz",
				Code:       "FILTER_VALIDATION_ERROR",
				HTTPStatus: 400,
				Suggestions: []string{
					"eq", "ne", "like", "not-like", "starts-with", "ends-with",
					"gt", "gte", "lt", "lte", "in", "not-in", "null", "not-null",
					"between", "not-between",
				},
			},
		},
		{
			name:  "missing value error",
			error: filter.NewMissingValueError("status", "eq"),
			expected: struct {
				Type        filter.ErrorType
				Message     string
				Field       string
				Operator    string
				Code        string
				HTTPStatus  int
				Suggestions []string
			}{
				Type:        filter.ErrorTypeValidation,
				Message:     "Filter value cannot be empty",
				Field:       "status",
				Operator:    "eq",
				Code:        "FILTER_VALIDATION_ERROR",
				HTTPStatus:  400,
				Suggestions: []string{"Provide a non-empty value for the filter"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.error

			if err.Type != tt.expected.Type {
				t.Errorf("Type = %v, want %v", err.Type, tt.expected.Type)
			}
			if err.Message != tt.expected.Message {
				t.Errorf("Message = %v, want %v", err.Message, tt.expected.Message)
			}
			if err.Field != tt.expected.Field {
				t.Errorf("Field = %v, want %v", err.Field, tt.expected.Field)
			}
			if err.Operator != tt.expected.Operator {
				t.Errorf("Operator = %v, want %v", err.Operator, tt.expected.Operator)
			}
			if err.Code != tt.expected.Code {
				t.Errorf("Code = %v, want %v", err.Code, tt.expected.Code)
			}
			if err.HTTPStatus != tt.expected.HTTPStatus {
				t.Errorf("HTTPStatus = %v, want %v", err.HTTPStatus, tt.expected.HTTPStatus)
			}
			if !reflect.DeepEqual(err.Suggestions, tt.expected.Suggestions) {
				t.Errorf("Suggestions = %v, want %v", err.Suggestions, tt.expected.Suggestions)
			}
		})
	}
}

func TestResult_ToJSONResponse(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *filter.Result
		expected map[string]interface{}
	}{
		{
			name: "successful result",
			setup: func() *filter.Result {
				return filter.NewResult(nil)
			},
			expected: map[string]interface{}{
				"success": true,
			},
		},
		{
			name: "result with errors",
			setup: func() *filter.Result {
				result := filter.NewResult(nil)
				result.AddError(filter.NewFieldNotAllowedError("email", []string{"name", "age"}))
				return result
			},
			expected: map[string]interface{}{
				"success": false,
				"errors": []map[string]interface{}{
					{
						"type":        "validation_error",
						"message":     "Field 'email' is not allowed for filtering",
						"field":       "email",
						"code":        "FILTER_VALIDATION_ERROR",
						"suggestions": []string{"name", "age"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.setup()
			got := result.ToJSONResponse()

			if got["success"] != tt.expected["success"] {
				t.Errorf("ToJSONResponse()['success'] = %v, want %v", got["success"], tt.expected["success"])
			}

			if tt.expected["errors"] != nil {
				if got["errors"] == nil {
					t.Error("ToJSONResponse()['errors'] = nil, want errors array")
				}
			}
		})
	}
}

func TestResult_IntegrationWithParser(t *testing.T) {
	values := url.Values{}
	values.Set("filter[invalid_field][eq]", "test")
	values.Set("filter[name][invalid_op]", "john")
	values.Set("filter[age][gt]", "25")

	parser := filter.NewParser(values)
	parseResult := parser.Parse()

	validator := filter.NewValidator([]string{"name", "age"}, nil)
	result := filter.NewResult(nil)

	for _, f := range parseResult.Filters {
		if validationErr := validator.ValidateFilter(f); validationErr != nil {
			result.AddError(validationErr)
		}
	}

	if parseResult.Errors.HasErrors() {
		result.AddErrors(parseResult.Errors.Errors...)
	}

	// Test the complete integration
	if result.Success {
		t.Error("Expected Success to be false when there are errors")
	}

	if !result.HasErrors() {
		t.Error("Expected HasErrors to be true")
	}

	if len(result.Errors.Errors) == 0 {
		t.Error("Expected at least one error")
	}

	// Test error details
	for i, err := range result.Errors.Errors {
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

	// Test first error access
	firstError := result.GetFirstError()
	if firstError == nil {
		t.Error("Expected to get a first error")
	} else if firstError.Message == "" {
		t.Error("First error should have a message")
	}

	// Test JSON response
	jsonResponse := result.ToJSONResponse()
	if jsonResponse["success"] != false {
		t.Error("JSON response should have success: false")
	}
	if jsonResponse["errors"] == nil {
		t.Error("JSON response should have errors field")
	}
}
