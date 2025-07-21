package filter

import (
	"fmt"
	"net/http"
)

type ErrorType string

const (
	ErrorTypeValidation    ErrorType = "validation_error"
	ErrorTypeParsing       ErrorType = "parsing_error"
	ErrorTypeConfiguration ErrorType = "configuration_error"
	ErrorTypeDatabase      ErrorType = "database_error"
	ErrorTypeInternal      ErrorType = "internal_error"
)

type FilterError struct {
	Type        ErrorType `json:"type"`
	Message     string    `json:"message"`
	Field       string    `json:"field,omitempty"`
	Operator    string    `json:"operator,omitempty"`
	Value       string    `json:"value,omitempty"`
	Code        string    `json:"code"`
	HTTPStatus  int       `json:"-"`
	InternalErr error     `json:"-"`
	Suggestions []string  `json:"suggestions,omitempty"`
}

// Converts the error to a readable string for logging/debugging
func (e *FilterError) Error() string {
	if e.Field != "" && e.Operator != "" {
		return fmt.Sprintf("%s: %s (field: %s, operator: %s)", e.Type, e.Message, e.Field, e.Operator)
	}
	if e.Field != "" {
		return fmt.Sprintf("%s: %s (field: %s)", e.Type, e.Message, e.Field)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *FilterError) Unwrap() error {
	return e.InternalErr
}

func (e *FilterError) ToJSONResponse() map[string]interface{} {

	errorData := map[string]interface{}{
		"type":    string(e.Type),
		"message": e.Message,
		"code":    e.Code,
	}

	if e.Field != "" {
		errorData["field"] = e.Field
	}
	if e.Operator != "" {
		errorData["operator"] = e.Operator
	}
	if e.Value != "" {
		errorData["value"] = e.Value
	}
	if len(e.Suggestions) > 0 {
		errorData["suggestions"] = e.Suggestions
	}

	return map[string]interface{}{
		"error": errorData,
	}
}

type FilterErrors struct {
	Errors []*FilterError `json:"errors"`
}

func (fe *FilterErrors) Error() string { // why do we need 2 errors?
	if len(fe.Errors) == 0 {
		return "no errors"
	}
	if len(fe.Errors) == 1 {
		return fe.Errors[0].Error()
	}
	return fmt.Sprintf("multiple filter errors (%d errors)", len(fe.Errors))
}

func (fe *FilterErrors) Add(err *FilterError) {
	fe.Errors = append(fe.Errors, err)
}

func (fe *FilterErrors) HasErrors() bool {
	return len(fe.Errors) > 0
}

func (fe *FilterErrors) ToJSONResponse() map[string]interface{} {
	errors := make([]map[string]interface{}, len(fe.Errors))
	for i, err := range fe.Errors {
		errors[i] = err.ToJSONResponse()["error"].(map[string]interface{})
	}
	return map[string]interface{}{
		"errors": errors,
	}
}

func NewValidationError(field, operator, value, message string, suggestions ...string) *FilterError {
	return &FilterError{
		Type:        ErrorTypeValidation,
		Message:     message,
		Field:       field,
		Operator:    operator,
		Value:       value,
		Code:        "FILTER_VALIDATION_ERROR",
		HTTPStatus:  http.StatusBadRequest,
		Suggestions: suggestions,
	}
}

func NewParsingError(field, value, message string, internalErr error) *FilterError {
	return &FilterError{
		Type:        ErrorTypeParsing,
		Message:     message,
		Field:       field,
		Value:       value,
		Code:        "FILTER_PARSING_ERROR",
		HTTPStatus:  http.StatusBadRequest,
		InternalErr: internalErr,
	}
}

func NewConfigurationError(message string, suggestions ...string) *FilterError {
	return &FilterError{
		Type:        ErrorTypeConfiguration,
		Message:     message,
		Code:        "FILTER_CONFIGURATION_ERROR",
		HTTPStatus:  http.StatusInternalServerError,
		Suggestions: suggestions,
	}
}

func NewDatabaseError(message string, internalErr error) *FilterError {
	return &FilterError{
		Type:        ErrorTypeDatabase,
		Message:     message,
		Code:        "FILTER_DATABASE_ERROR",
		HTTPStatus:  http.StatusInternalServerError,
		InternalErr: internalErr,
	}
}

func NewInternalError(message string, internalErr error) *FilterError {
	return &FilterError{
		Type:        ErrorTypeInternal,
		Message:     message,
		Code:        "FILTER_INTERNAL_ERROR",
		HTTPStatus:  http.StatusInternalServerError,
		InternalErr: internalErr,
	}
}

func NewFieldNotAllowedError(field string, allowedFields []string) *FilterError {
	suggestions := make([]string, len(allowedFields))
	copy(suggestions, allowedFields)

	return NewValidationError(
		field, "", "",
		fmt.Sprintf("Field '%s' is not allowed for filtering", field),
		suggestions...,
	)
}

func NewOperatorNotAllowedError(field, operator string, allowedOperators []Clause) *FilterError {
	suggestions := make([]string, len(allowedOperators))
	for i, op := range allowedOperators {
		suggestions[i] = string(op)
	}

	return NewValidationError(
		field, operator, "",
		fmt.Sprintf("Operator '%s' is not allowed for field '%s'", operator, field),
		suggestions...,
	)
}

func NewInvalidOperatorError(operator string) *FilterError {
	validOperators := []string{
		string(Equals), string(NotEquals), string(Contains), string(NotContains),
		string(StartsWith), string(EndsWith), string(GreaterThan), string(GreaterThanOrEq),
		string(LessThan), string(LessThanOrEq), string(In), string(NotIn),
		string(IsNull), string(IsNotNull), string(Between), string(NotBetween),
	}

	return NewValidationError(
		"", operator, "",
		fmt.Sprintf("Invalid operator '%s'", operator),
		validOperators...,
	)
}

func NewInvalidFilterFormatError(filterKey, value string) *FilterError {
	return NewParsingError(
		"", value,
		fmt.Sprintf("Invalid filter format: '%s'. Expected format: 'filter[field]' or 'filter[field][operator]'", filterKey),
		nil,
	)
}

func NewMissingValueError(field, operator string) *FilterError {
	return NewValidationError(
		field, operator, "",
		"Filter value cannot be empty",
		"Provide a non-empty value for the filter",
	)
}

func NewInvalidBetweenValueError(field, value string) *FilterError {
	return NewValidationError(
		field, string(Between), value,
		"Between operator requires exactly two comma-separated values",
		"Use format: 'value1,value2' (e.g., '10,20')",
	)
}

func NewSortFieldNotAllowedError(field string, allowedFields []string) *FilterError {
	suggestions := make([]string, len(allowedFields))
	copy(suggestions, allowedFields)

	return NewValidationError(
		field, "", "",
		fmt.Sprintf("Sort field '%s' is not allowed", field),
		suggestions...,
	)
}
