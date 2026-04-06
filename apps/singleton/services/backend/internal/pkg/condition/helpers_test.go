package condition

import "testing"

func TestParseHHMM(t *testing.T) {
	tests := []struct {
		input string
		want  int
		ok    bool
	}{
		{"00:00", 0, true},
		{"09:30", 570, true},
		{"23:59", 1439, true},
		{"12:00", 720, true},
		{"25:00", 0, false},
		{"12:60", 0, false},
		{"-1:00", 0, false},
		{"09", 0, false},
		{"", 0, false},
		{"ab:cd", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := parseHHMM(tt.input)
			if ok != tt.ok {
				t.Errorf("parseHHMM(%q) ok = %v, want %v", tt.input, ok, tt.ok)
			}
			if ok && got != tt.want {
				t.Errorf("parseHHMM(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestAtoi(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"0", 0},
		{"42", 42},
		{"123", 123},
		{"-1", -1},
		{"12a", -1},
		{"", 0},
		{" 7 ", 7},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := atoi(tt.input); got != tt.want {
				t.Errorf("atoi(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestCompareOrdered(t *testing.T) {
	tests := []struct {
		actual   string
		expected string
		op       string
		want     bool
	}{
		// Numeric comparison
		{"10", "5", "gt", true},
		{"5", "10", "gt", false},
		{"10", "10", "gte", true},
		{"9", "10", "gte", false},
		{"5", "10", "lt", true},
		{"10", "5", "lt", false},
		{"10", "10", "lte", true},
		{"11", "10", "lte", false},
		// String fallback
		{"b", "a", "gt", true},
		{"a", "b", "gt", false},
	}
	for _, tt := range tests {
		name := tt.actual + tt.op + tt.expected
		t.Run(name, func(t *testing.T) {
			got, err := compareOrdered(tt.actual, tt.expected, tt.op)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("compareOrdered(%q, %q, %q) = %v, want %v", tt.actual, tt.expected, tt.op, got, tt.want)
			}
		})
	}
}
