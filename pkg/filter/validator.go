package filter

import "slices"

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
