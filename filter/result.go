package filter

import (
	"github.com/uptrace/bun"
)

// Result represents the outcome of filter operations
type Result struct {
	Query   *bun.SelectQuery `json:"-"`
	Errors  *FilterErrors    `json:"errors,omitempty"`
	Success bool             `json:"success"`
}

func NewResult(query *bun.SelectQuery) *Result {
	return &Result{
		Query:   query,
		Errors:  &FilterErrors{},
		Success: true,
	}
}

func NewErrorResult(errors ...*FilterError) *Result {
	filterErrors := &FilterErrors{}
	for _, err := range errors {
		filterErrors.Add(err)
	}

	return &Result{
		Query:   nil,
		Errors:  filterErrors,
		Success: false,
	}
}

func (r *Result) AddError(err *FilterError) {
	if r.Errors == nil {
		r.Errors = &FilterErrors{}
	}
	r.Errors.Add(err)
	r.Success = false
}

func (r *Result) AddErrors(errors ...*FilterError) {
	for _, err := range errors {
		r.AddError(err)
	}
}

func (r *Result) HasErrors() bool {
	return r.Errors != nil && r.Errors.HasErrors()
}

func (r *Result) GetQuery() *bun.SelectQuery {
	if r.Success {
		return r.Query
	}
	return nil
}

func (r *Result) GetFirstError() *FilterError {
	if r.HasErrors() && len(r.Errors.Errors) > 0 {
		return r.Errors.Errors[0]
	}
	return nil
}

func (r *Result) ToJSONResponse() map[string]interface{} {
	response := map[string]interface{}{
		"success": r.Success,
	}

	if r.HasErrors() {
		response["errors"] = r.Errors.ToJSONResponse()["errors"]
	}

	return response
}
