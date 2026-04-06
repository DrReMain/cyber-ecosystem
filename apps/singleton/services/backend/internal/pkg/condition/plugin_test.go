package condition

import (
	"context"
	"testing"
)

func newTestRegistry() *ConditionRegistry {
	return NewBuiltinConditionRegistry()
}

// --- TimeRange ---

func TestTimeRangePlugin_ValidateConfig(t *testing.T) {
	r := newTestRegistry()
	tests := []struct {
		name    string
		config  string
		wantErr bool
	}{
		{"valid", `{"start":"09:00","end":"18:00"}`, false},
		{"empty", "", true},
		{"invalid_json", `{invalid}`, true},
		{"missing_start", `{"end":"18:00"}`, true},
		{"missing_end", `{"start":"09:00"}`, true},
		{"bad_start", `{"start":"25:00","end":"18:00"}`, true},
		{"bad_end", `{"start":"09:00","end":"25:00"}`, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := r.Validate("time_range", tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate(time_range, %q) error = %v, wantErr %v", tt.config, err, tt.wantErr)
			}
		})
	}
}

// --- IPRange ---

func TestIPRangePlugin_ValidateConfig(t *testing.T) {
	r := newTestRegistry()
	tests := []struct {
		name    string
		config  string
		wantErr bool
	}{
		{"valid", `{"ips":["10.0.0.0/8","192.168.1.0/24"]}`, false},
		{"empty", "", true},
		{"invalid_json", `{invalid}`, true},
		{"no_ips", `{"ips":[]}`, true},
		{"invalid_cidr", `{"ips":["not-a-cidr"]}`, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := r.Validate("ip_range", tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate(ip_range, %q) error = %v, wantErr %v", tt.config, err, tt.wantErr)
			}
		})
	}
}

// --- DayOfWeek ---

func TestDayOfWeekPlugin_ValidateConfig(t *testing.T) {
	r := newTestRegistry()
	tests := []struct {
		name    string
		config  string
		wantErr bool
	}{
		{"valid", `{"days":["monday","friday"]}`, false},
		{"empty", "", true},
		{"invalid_json", `{invalid}`, true},
		{"no_days", `{"days":[]}`, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := r.Validate("day_of_week", tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate(day_of_week, %q) error = %v, wantErr %v", tt.config, err, tt.wantErr)
			}
		})
	}
}

// --- AttributeMatch ---

func TestAttributeMatchPlugin_ValidateConfig(t *testing.T) {
	r := newTestRegistry()
	tests := []struct {
		name    string
		config  string
		wantErr bool
	}{
		{"valid_eq", `{"key":"dept","op":"eq","value":"eng"}`, false},
		{"valid_in", `{"key":"level","op":"in","values":["junior","mid"]}`, false},
		{"valid_gte", `{"key":"age","op":"gte","value":"18"}`, false},
		{"empty", "", true},
		{"invalid_json", `{invalid}`, true},
		{"missing_key", `{"op":"eq","value":"eng"}`, true},
		{"invalid_op", `{"key":"dept","op":"contains","value":"eng"}`, true},
		{"in_empty_values", `{"key":"level","op":"in","values":[]}`, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := r.Validate("attribute_match", tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate(attribute_match, %q) error = %v, wantErr %v", tt.config, err, tt.wantErr)
			}
		})
	}
}

// --- Registry ---

func TestConditionRegistry(t *testing.T) {
	r := newTestRegistry()
	expected := []string{"time_range", "ip_range", "day_of_week", "attribute_match"}
	for _, k := range expected {
		if _, ok := r.Get(k); !ok {
			t.Errorf("registry missing plugin %q", k)
		}
	}
	if _, ok := r.Get("nonexistent"); ok {
		t.Error("expected nonexistent plugin to not be found")
	}
	if err := r.Validate("nonexistent", ""); err == nil {
		t.Error("Validate(nonexistent, '') = nil, want error")
	}
}

// --- Evaluate integration ---

func TestAttributeMatchPlugin_Evaluate(t *testing.T) {
	r := newTestRegistry()
	ctx := context.Background()
	ctx = WithUserAttributes(ctx, map[string]string{"dept": "eng"})

	allowed, err := r.Evaluate(ctx, "attribute_match", `{"key":"dept","op":"eq","value":"eng"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("expected allowed = true")
	}

	allowed, err = r.Evaluate(ctx, "attribute_match", `{"key":"dept","op":"eq","value":"sales"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Error("expected allowed = false")
	}
}
