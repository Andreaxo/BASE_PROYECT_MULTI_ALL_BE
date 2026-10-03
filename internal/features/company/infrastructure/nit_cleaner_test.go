package infrastructure

import (
	"strconv"
	"strings"
	"testing"
)

func parseNIT(clean string) (int, bool) {
	cleanNIT := strings.Split(clean, "-")[0]
	cleanNIT = strings.ReplaceAll(cleanNIT, ".", "")
	cleanNIT = strings.TrimSpace(cleanNIT)
	nitVal, err := strconv.Atoi(cleanNIT)
	if err != nil {
		return 0, false
	}
	return nitVal, true
}

func TestParseNIT(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		valid    bool
	}{
		{"900123456", 900123456, true},
		{"900.123.456-1", 900123456, true},
		{" 800.555.123 - 4 ", 800555123, true},
		{"12345", 12345, true},
		{"CORP-01", 0, false},
		{"EMPRESA_XYZ", 0, false},
		{"", 0, false},
	}

	for _, tt := range tests {
		val, ok := parseNIT(tt.input)
		if ok != tt.valid {
			t.Errorf("parseNIT(%q) valid = %v, expected %v", tt.input, ok, tt.valid)
		}
		if ok && val != tt.expected {
			t.Errorf("parseNIT(%q) val = %d, expected %d", tt.input, val, tt.expected)
		}
	}
}
