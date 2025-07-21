package filter

import (
	"fmt"
	"slices"
)

// Validator handles validation of filters against configurations
type Validator struct {
	allowedFields []string
	configs       []FilterConfig
	useConfigs    bool
}

func NewValidator(allowedFields []string, configs []FilterConfig) *Validator {
	return &Validator{
		allowedFields: allowedFields,
		configs:       configs,
		useConfigs:    len(configs) > 0,
	}
}

func (v *Validator) IsFilterAllowed(filter Filter) bool {
	if v.useConfigs {
		return v.isFilterAllowedByConfig(filter)
	} else if v.allowedFields != nil {
		return slices.Contains(v.allowedFields, filter.Field)
	}
	return true
}

func (v *Validator) isFilterAllowedByConfig(filter Filter) bool {
	if v.configs == nil {
		return true
	}

	for _, config := range v.configs {
		if config.Field == filter.Field {
			return slices.Contains(config.AllowedOperators, filter.Operator)
		}
	}

	return false
}

func (v *Validator) GetAllowedFields() []string {
	if v.useConfigs && len(v.configs) > 0 {
		fields := make([]string, len(v.configs))
		for i, config := range v.configs {
			fields[i] = config.Field
		}
		return fields
	}

	return v.allowedFields
}

func (v *Validator) ValidateFilter(filter Filter) *FilterError {
	if !filter.Operator.IsValid() {
		return NewInvalidOperatorError(string(filter.Operator))
	}

	if filter.Field == "" {
		return NewValidationError("", string(filter.Operator), fmt.Sprintf("%v", filter.Value), "Field name cannot be empty")
	}

	// Check if value is empty for operators that require it
	if filter.Value == "" || filter.Value == nil {
		switch filter.Operator {
		case IsNull, IsNotNull:
			// These operators don't need values
		default:
			return NewMissingValueError(filter.Field, string(filter.Operator))
		}
	}

	switch filter.Operator {
	case Between, NotBetween:
		valueStr := fmt.Sprintf("%v", filter.Value)
		values := parseCommaSeparatedValues(valueStr)
		if len(values) != 2 {
			return NewInvalidBetweenValueError(filter.Field, valueStr)
		}
	}

	if !v.IsFilterAllowed(filter) {
		if v.useConfigs {
			for _, config := range v.configs {
				if config.Field == filter.Field {
					return NewOperatorNotAllowedError(filter.Field, string(filter.Operator), config.AllowedOperators)
				}
			}

			allowedFields := make([]string, len(v.configs))
			for i, config := range v.configs {
				allowedFields[i] = config.Field
			}
			return NewFieldNotAllowedError(filter.Field, allowedFields)
		} else {
			return NewFieldNotAllowedError(filter.Field, v.allowedFields)
		}
	}

	return nil
}
