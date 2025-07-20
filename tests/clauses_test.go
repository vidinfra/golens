package filter_test

import (
	"testing"

	"github.com/vidinfra/golens/filter"
)

func TestClause_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		clause   filter.Clause
		expected bool
	}{
		{"valid equals", filter.Equals, true},
		{"valid contains", filter.Contains, true},
		{"valid greater than", filter.GreaterThan, true},
		{"valid in", filter.In, true},
		{"valid is null", filter.IsNull, true},
		{"invalid clause", filter.Clause("invalid"), false},
		{"empty clause", filter.Clause(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.clause.IsValid(); got != tt.expected {
				t.Errorf("Clause.IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestClause_String(t *testing.T) {
	tests := []struct {
		name     string
		clause   filter.Clause
		expected string
	}{
		{"equals", filter.Equals, "eq"},
		{"contains", filter.Contains, "like"},
		{"greater than", filter.GreaterThan, "gt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.clause.String(); got != tt.expected {
				t.Errorf("Clause.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}
