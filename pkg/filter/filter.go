package filter

// Filter represents a single filter condition
type Filter struct {
	Value    any    `json:"value"`
	Field    string `json:"field"`
	Operator Clause `json:"operator"`
}
