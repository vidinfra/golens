package filter

import (
	"fmt"
	"strings"
)

// Validator enforces which fields/operators are allowed and validates value shapes.
type Validator struct {
	fieldSet      map[string]struct{}
	opsPerField   map[string]map[Clause]struct{}
	allowedFields []string
	configs       []FilterConfig
}

func NewValidator(allowedFields []string, configs []FilterConfig) *Validator {
	v := &Validator{
		allowedFields: allowedFields,
		configs:       configs,
		fieldSet:      map[string]struct{}{},
		opsPerField:   map[string]map[Clause]struct{}{},
	}

	// If configs are provided, they define both allowed fields and allowed operators.
	if len(configs) > 0 {
		for _, c := range configs {
			// allowed field
			v.fieldSet[c.Field] = struct{}{}

			// per-field operator allowlist (if provided)
			if len(c.AllowedOperators) > 0 {
				opset := make(map[Clause]struct{}, len(c.AllowedOperators))
				for _, op := range c.AllowedOperators {
					opset[op] = struct{}{}
				}
				v.opsPerField[c.Field] = opset
			}
		}
		return v
	}

	// Else, fall back to a simple field allowlist.
	for _, f := range allowedFields {
		v.fieldSet[f] = struct{}{}
	}
	return v
}

// IsFilterAllowed returns whether a filter's FIELD is allowed.
// - If configs are present: only fields present in configs are allowed.
// - Else if allowedFields provided: must be in that list.
// - Else: allow any field (no restriction).
func (v *Validator) IsFilterAllowed(f Filter) bool {
	if len(v.configs) > 0 {
		_, ok := v.fieldSet[f.Field]
		return ok
	}
	if len(v.allowedFields) > 0 {
		_, ok := v.fieldSet[f.Field]
		return ok
	}
	// No allowlists configured → allow all fields
	return true
}

// GetAllowedFields returns the effective allowed fields.
func (v *Validator) GetAllowedFields() []string {
	if len(v.configs) > 0 {
		fields := make([]string, 0, len(v.fieldSet))
		for field := range v.fieldSet {
			fields = append(fields, field)
		}
		return fields
	}
	return v.allowedFields
}

// ValidateFilter performs full validation of a single filter:
// 1) field is allowed
// 2) operator is valid and allowed for that field (when configs provided)
// 3) value is present/has the right shape for the operator
func (v *Validator) ValidateFilter(f Filter) *FilterError {
	// Operator must be defined/known
	if !f.Operator.IsValid() {
		return NewInvalidOperatorError(string(f.Operator))
	}

	// Field must be non-empty
	if strings.TrimSpace(f.Field) == "" {
		return NewValidationError("", string(f.Operator), fmt.Sprintf("%v", f.Value), "Field name cannot be empty")
	}

	// Field allow-check
	if !v.IsFilterAllowed(f) {
		if len(v.configs) > 0 {
			allowed := v.GetAllowedFields()
			// If field exists in configs but operator not allowed, return operator error below.
			// Otherwise, field is not allowed at all:
			return NewFieldNotAllowedError(f.Field, allowed)
		}
		return NewFieldNotAllowedError(f.Field, v.allowedFields)
	}

	// Operator allow-check (only enforced when configs define per-field ops)
	if len(v.configs) > 0 {
		if ops, ok := v.opsPerField[f.Field]; ok && len(ops) > 0 {
			if _, allowed := ops[f.Operator]; !allowed {
				// build suggestions from configured operators
				suggestions := make([]Clause, 0, len(ops))
				for op := range ops {
					suggestions = append(suggestions, op)
				}
				return NewOperatorNotAllowedError(f.Field, string(f.Operator), suggestions)
			}
		}
	}

	// Value shape/semantics by operator
	if err := v.validateValueByOperator(f); err != nil {
		return err
	}

	return nil
}

func (v *Validator) validateValueByOperator(f Filter) *FilterError {
	op := f.Operator
	val := f.Value
	raw := strings.TrimSpace(fmt.Sprint(val))

	switch op {
	case IsNull, IsNotNull:
		// No value required/used
		return nil

	case Between, NotBetween:
		if raw == "" {
			return NewInvalidBetweenValueError(f.Field, raw)
		}
		parts := parseCommaSeparatedValues(raw)
		if len(parts) != 2 {
			return NewInvalidBetweenValueError(f.Field, raw)
		}
		return nil

	case In, NotIn:
		if raw == "" {
			return NewMissingValueError(f.Field, string(op))
		}
		parts := parseCommaSeparatedValues(raw)
		if len(parts) == 0 {
			return NewMissingValueError(f.Field, string(op))
		}
		return nil

	default:
		// All other operators require a non-empty value.
		if raw == "" {
			return NewMissingValueError(f.Field, string(op))
		}
		return nil
	}
}
