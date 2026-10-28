package sms

import (
	"errors"
	"strings"
	"testing"
)

// TestIsPhone is the exhaustive shape check: optional leading plus, digits
// only, at least five, at most twenty. Market validity is deliberately
// absent — the channel that carries the number owns that verdict.
func TestIsPhone(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "empty", input: "", want: false},
		{name: "too_short", input: "1234", want: false},
		{name: "mainland_mobile", input: "13800138000", want: true},
		{name: "e164", input: "+8613800138000", want: true},
		{name: "plus_not_leading", input: "138+00", want: false},
		{name: "dashes_rejected", input: "138-0013-8000", want: false},
		{name: "spaces_rejected", input: "138 0013", want: false},
		{name: "letters_rejected", input: "1380013800a", want: false},
		{name: "at_cap_20", input: strings.Repeat("1", 20), want: true},
		{name: "over_cap_21", input: strings.Repeat("1", 21), want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isPhone(tc.input); got != tc.want {
				t.Fatalf("isPhone(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func validMessage() *Message {
	return &Message{
		To:           "13800138000",
		SignName:     "曦光未来",
		TemplateCode: "SMS_460000000",
		Params:       map[string]string{"code": "123456"},
	}
}

// TestValidateMessage walks the shape contract: each violation alone flips
// the verdict, and the well-formed message passes. Template existence and
// deliverability stay out — provider state, never a shape check.
func TestValidateMessage(t *testing.T) {
	if err := ValidateMessage(validMessage()); err != nil {
		t.Fatalf("valid message rejected: %v", err)
	}
	mutators := []struct {
		name string
		mut  func(m *Message)
	}{
		{"nil_message", func(m *Message) { *m = Message{} }},
		{"empty_to", func(m *Message) { m.To = "" }},
		{"short_to", func(m *Message) { m.To = "123" }},
		{"empty_sign", func(m *Message) { m.SignName = "" }},
		{"overlong_sign", func(m *Message) { m.SignName = strings.Repeat("s", 65) }},
		{"empty_template", func(m *Message) { m.TemplateCode = "" }},
		{"overlong_template", func(m *Message) { m.TemplateCode = strings.Repeat("T", 65) }},
		{"too_many_params", func(m *Message) {
			p := map[string]string{}
			for i := range maxParams + 1 {
				p["k"+string(rune('a'+i%26))+string(rune('0'+i/26))] = "v"
			}
			m.Params = p
		}},
		{"empty_param_key", func(m *Message) { m.Params = map[string]string{"": "v"} }},
		{"overlong_param_key", func(m *Message) { m.Params = map[string]string{strings.Repeat("k", 33): "v"} }},
		{"overlong_param_value", func(m *Message) { m.Params = map[string]string{"code": strings.Repeat("v", 257)} }},
	}
	for _, tc := range mutators {
		t.Run(tc.name, func(t *testing.T) {
			m := validMessage()
			tc.mut(m)
			if err := ValidateMessage(m); !errors.Is(err, ErrInvalidArgument) {
				t.Fatalf("want ErrInvalidArgument, got %v", err)
			}
		})
	}
	// Nil params are legal: templates without placeholders exist.
	m := validMessage()
	m.Params = nil
	if err := ValidateMessage(m); err != nil {
		t.Fatalf("nil params rejected: %v", err)
	}
}
