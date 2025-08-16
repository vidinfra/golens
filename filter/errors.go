// errors.go
package filter

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorType for grouping error families
type ErrorType string

const (
	ErrorTypeValidation    ErrorType = "validation_error"
	ErrorTypeParsing       ErrorType = "parsing_error"
	ErrorTypeConfiguration ErrorType = "configuration_error"
	ErrorTypeDatabase      ErrorType = "database_error"
	ErrorTypeInternal      ErrorType = "internal_error"
)

// ErrorCode enumerates machine-friendly codes for routing/i18n
type ErrorCode string

const (
	CodeFilterValidation ErrorCode = "FILTER_VALIDATION_ERROR"
	CodeFilterParsing    ErrorCode = "FILTER_PARSING_ERROR"
	CodeFilterConfig     ErrorCode = "FILTER_CONFIGURATION_ERROR"
	CodeFilterDatabase   ErrorCode = "FILTER_DATABASE_ERROR"
	CodeFilterInternal   ErrorCode = "FILTER_INTERNAL_ERROR"
)

// FilterError is a structured error suitable for programmatic handling and JSON output.
type FilterError struct {
	// Interface first; allows wrapping without allocation copies
	InternalErr error `json:"-"` // not serialized

	// Stringy, i18n-friendly fields
	Type     ErrorType `json:"type"`
	Message  string    `json:"message"`
	Field    string    `json:"field,omitempty"`
	Operator string    `json:"operator,omitempty"`
	Value    string    `json:"value,omitempty"`
	Code     ErrorCode `json:"code"`

	// UX helpers
	Suggestions []string `json:"suggestions,omitempty"`

	// Transport concern (kept here for convenience)
	HTTPStatus int `json:"-"`
}

// Error implements the standard error interface.
func (e *FilterError) Error() string {
	switch {
	case e.Field != "" && e.Operator != "":
		return fmt.Sprintf("%s: %s (field: %s, operator: %s)", e.Type, e.Message, e.Field, e.Operator)
	case e.Field != "":
		return fmt.Sprintf("%s: %s (field: %s)", e.Type, e.Message, e.Field)
	default:
		return fmt.Sprintf("%s: %s", e.Type, e.Message)
	}
}

// Unwrap exposes the wrapped/internal error.
func (e *FilterError) Unwrap() error { return e.InternalErr }

// ToJSONResponse renders a single-error payload.
// Keep signature for backward compatibility.
func (e *FilterError) ToJSONResponse() map[string]any {
	errObj := map[string]any{
		"type":    string(e.Type),
		"message": e.Message,
		"code":    string(e.Code),
	}
	if e.Field != "" {
		errObj["field"] = e.Field
	}
	if e.Operator != "" {
		errObj["operator"] = e.Operator
	}
	if e.Value != "" {
		errObj["value"] = e.Value
	}
	if len(e.Suggestions) > 0 {
		errObj["suggestions"] = e.Suggestions
	}
	return map[string]any{"error": errObj}
}

// Status returns the HTTP status to use for this error.
// Falls back based on type if HTTPStatus is zero.
func (e *FilterError) Status() int {
	if e.HTTPStatus != 0 {
		return e.HTTPStatus
	}
	switch e.Type {
	case ErrorTypeValidation, ErrorTypeParsing:
		return http.StatusBadRequest
	case ErrorTypeConfiguration, ErrorTypeInternal, ErrorTypeDatabase:
		return http.StatusInternalServerError
	default:
		return http.StatusBadRequest
	}
}

// FilterErrors aggregates multiple FilterError values.
type FilterErrors struct {
	Errors []*FilterError `json:"errors"`
}

func (fe *FilterErrors) OK() bool {
	return fe == nil || len(fe.Errors) == 0
}

// error implements error; summarizes count.
func (fe *FilterErrors) Error() string {
	n := len(fe.Errors)
	switch n {
	case 0:
		return "no errors"
	case 1:
		return fe.Errors[0].Error()
	default:
		return fmt.Sprintf("multiple filter errors (%d errors)", n)
	}
}

func (fe *FilterErrors) Add(err *FilterError) {
	if err == nil {
		return
	}
	fe.Errors = append(fe.Errors, err)
}

// AddAll appends a list of errors.
func (fe *FilterErrors) AddAll(errs ...*FilterError) {
	for _, e := range errs {
		if e != nil {
			fe.Errors = append(fe.Errors, e)
		}
	}
}

// Merge appends errors from another FilterErrors.
func (fe *FilterErrors) Merge(other *FilterErrors) {
	if other == nil || len(other.Errors) == 0 {
		return
	}
	fe.Errors = append(fe.Errors, other.Errors...)
}

// HasErrors indicates any error present.
func (fe *FilterErrors) HasErrors() bool { return len(fe.Errors) > 0 }

// Len returns the number of errors.
func (fe *FilterErrors) Len() int { return len(fe.Errors) }

// First returns the first error or nil.
func (fe *FilterErrors) First() *FilterError {
	if len(fe.Errors) == 0 {
		return nil
	}
	return fe.Errors[0]
}

// Status derives an HTTP status for an error list.
// If mixed types, prefer 400 for client issues else 500.
func (fe *FilterErrors) Status() int {
	if len(fe.Errors) == 0 {
		return http.StatusOK
	}
	// If any server-side error appears, prefer 500.
	hasClient := false
	for _, e := range fe.Errors {
		st := e.Status()
		if st >= 500 {
			return http.StatusInternalServerError
		}
		if st == http.StatusBadRequest {
			hasClient = true
		}
	}
	if hasClient {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

// ToJSONResponse renders a multi-error response.
// Keeps the shape from the existing implementation.
func (fe *FilterErrors) ToJSONResponse() map[string]any {
	arr := make([]map[string]any, 0, len(fe.Errors))
	for _, e := range fe.Errors {
		// Avoid type assertions by building the inner object directly
		item := map[string]any{
			"type":    string(e.Type),
			"message": e.Message,
			"code":    string(e.Code),
		}
		if e.Field != "" {
			item["field"] = e.Field
		}
		if e.Operator != "" {
			item["operator"] = e.Operator
		}
		if e.Value != "" {
			item["value"] = e.Value
		}
		if len(e.Suggestions) > 0 {
			item["suggestions"] = e.Suggestions
		}
		arr = append(arr, item)
	}
	return map[string]any{"errors": arr}
}

// Convenience helpers / constructors (backward-compatible)

func NewValidationError(field, operator, value, message string, suggestions ...string) *FilterError {
	return &FilterError{
		Type:        ErrorTypeValidation,
		Message:     message,
		Field:       field,
		Operator:    operator,
		Value:       value,
		Code:        CodeFilterValidation,
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
		Code:        CodeFilterParsing,
		HTTPStatus:  http.StatusBadRequest,
		InternalErr: internalErr,
	}
}

func NewConfigurationError(message string, suggestions ...string) *FilterError {
	return &FilterError{
		Type:        ErrorTypeConfiguration,
		Message:     message,
		Code:        CodeFilterConfig,
		HTTPStatus:  http.StatusInternalServerError,
		Suggestions: suggestions,
	}
}

func NewDatabaseError(message string, internalErr error) *FilterError {
	return &FilterError{
		Type:        ErrorTypeDatabase,
		Message:     message,
		Code:        CodeFilterDatabase,
		HTTPStatus:  http.StatusInternalServerError,
		InternalErr: internalErr,
	}
}

func NewInternalError(message string, internalErr error) *FilterError {
	return &FilterError{
		Type:        ErrorTypeInternal,
		Message:     message,
		Code:        CodeFilterInternal,
		HTTPStatus:  http.StatusInternalServerError,
		InternalErr: internalErr,
	}
}

func NewFieldNotAllowedError(field string, allowedFields []string) *FilterError {
	// Copy to avoid external slice mutations showing up in error payload
	suggestions := append([]string(nil), allowedFields...)
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
	suggestions := append([]string(nil), allowedFields...)
	return NewValidationError(
		field, "", "",
		fmt.Sprintf("Sort field '%s' is not allowed", field),
		suggestions...,
	)
}

// Helper to combine a generic error into a FilterError when needed.
func WrapAsInternalFilterError(msg string, err error) *FilterError {
	return &FilterError{
		Type:        ErrorTypeInternal,
		Message:     msg,
		Code:        CodeFilterInternal,
		HTTPStatus:  http.StatusInternalServerError,
		InternalErr: err,
	}
}

// Utility to check if any underlying error matches a target (errors.Is)
func (fe *FilterErrors) AnyIs(target error) bool {
	for _, e := range fe.Errors {
		if errors.Is(e, target) || (e.InternalErr != nil && errors.Is(e.InternalErr, target)) {
			return true
		}
	}
	return false
}
