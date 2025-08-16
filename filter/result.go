package filter

import (
	"gorm.io/gorm"
)

// Result represents the outcome of filter operations
type Result struct {
	Query   *gorm.DB      `json:"-"`
	Errors  *FilterErrors `json:"errors,omitempty"`
	Success bool          `json:"success"`
}

func NewResult(query *gorm.DB) *Result {
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

// ✅ Unified: just OK()
func (r *Result) OK() bool {
	return r.Errors == nil || !r.Errors.HasErrors()
}

func (r *Result) GetQuery() *gorm.DB {
	if r.OK() {
		return r.Query
	}
	return nil
}

func (r *Result) GetFirstError() *FilterError {
	if r.OK() || len(r.Errors.Errors) == 0 {
		return nil
	}
	return r.Errors.Errors[0]
}

func (r *Result) ToJSONResponse() map[string]any {
	response := map[string]any{
		"success": r.OK(),
	}
	if !r.OK() {
		response["errors"] = r.Errors.ToJSONResponse()["errors"]
	}
	return response
}
