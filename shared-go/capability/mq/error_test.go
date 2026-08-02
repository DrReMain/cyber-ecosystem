package mq

import (
	"errors"
	"strings"
	"testing"

	kratoserrors "github.com/go-kratos/kratos/v3/errors"
)

func TestHandleMQError(t *testing.T) {
	errs := &MQDefaultError{
		InvalidArgument: kratoserrors.New(400, "INVALID_ARGUMENT", "x"),
		Unavailable:     kratoserrors.New(503, "UNAVAILABLE", "x"),
		Timeout:         kratoserrors.New(504, "TIMEOUT", "x"),
	}
	cases := []struct {
		in   error
		want int32
	}{
		{ErrInvalidArgument, 400},
		{ErrUnavailable, 503},
		{ErrTimeout, 504},
		{errors.New("boom"), 503}, // unknown → Unavailable
	}
	for _, c := range cases {
		got := HandleMQError(c.in, errs)
		var ke *kratoserrors.Error
		if !errors.As(got, &ke) || ke.Code != c.want {
			t.Errorf("HandleMQError(%v): code=%v, want %d", c.in, ke.Code, c.want)
		}
	}
}

func TestValidateMQDefaultError(t *testing.T) {
	if err := ValidateMQDefaultError(&MQDefaultError{}); err == nil {
		t.Fatal("want error for all-nil slots")
	}
	if err := ValidateMQDefaultError(&MQDefaultError{
		InvalidArgument: kratoserrors.New(400, "i", ""),
		Unavailable:     kratoserrors.New(503, "u", ""),
	}); err != nil {
		t.Fatalf("nil Timeout should be allowed: %v", err)
	}
}

// TestValidateTopicGroupCharset is an exhaustive charset check for ValidateTopic
// and ValidateGroup. Both share the isSubjectToken rule (a token safe as BOTH a
// NATS subject literal and a stream/consumer-name fragment: the stream name is
// "mq-"+topic and the durable name is group+"-"+topic, so '.', '>', '*', '/',
// whitespace, control chars and any non-ASCII rune are illegal). The table is
// run against both functions; each function only differs in its own length cap.
func TestValidateTopicGroupCharset(t *testing.T) {
	// Build over-length inputs at exactly one past each cap.
	tooLongTopic := strings.Repeat("a", maxTopicLen+1)
	tooLongGroup := strings.Repeat("a", maxGroupLen+1)

	cases := []struct {
		name  string
		input string
		want  error // nil == accept; ErrInvalidArgument == reject
	}{
		// --- REJECT: empty / over-long ---
		{name: "empty", input: "", want: ErrInvalidArgument},
		{name: "overlong_topic", input: tooLongTopic, want: ErrInvalidArgument},
		{name: "overlong_group", input: tooLongGroup, want: ErrInvalidArgument},

		// --- REJECT: NATS wildcard / separator chars ---
		{name: "dot", input: "a.b", want: ErrInvalidArgument},
		{name: "gt_wildcard", input: "a>b", want: ErrInvalidArgument},
		{name: "star_wildcard", input: "a*b", want: ErrInvalidArgument},
		{name: "slash", input: "a/b", want: ErrInvalidArgument},

		// --- REJECT: whitespace ---
		{name: "space", input: "a b", want: ErrInvalidArgument},
		{name: "tab", input: "a\tb", want: ErrInvalidArgument},
		{name: "newline", input: "a\nb", want: ErrInvalidArgument},

		// --- REJECT: non-ASCII (CJK, emoji) and control chars ---
		{name: "cjk", input: "a.中", want: ErrInvalidArgument},
		{name: "emoji", input: "a🎉b", want: ErrInvalidArgument},
		{name: "nul_control", input: "a\x00b", want: ErrInvalidArgument},

		// --- ACCEPT: plain ASCII tokens ---
		{name: "single", input: "a", want: nil},
		{name: "dash", input: "a-b", want: nil},
		{name: "underscore", input: "a_b", want: nil},
		{name: "alphanumeric_mixed", input: "A1", want: nil},
		{name: "mixed_token", input: "grp_1-topicA", want: nil},
	}

	for _, c := range cases {
		// At-cap is still ACCEPT (boundary): confirm the cap itself passes for the
		// at-max form of an otherwise-valid token, then the over-cap form rejects.
		t.Run("topic/"+c.name, func(t *testing.T) {
			if c.name == "overlong_group" {
				return // overlong_group is a group-only fixture here
			}
			if got := ValidateTopic(c.input); !errors.Is(got, c.want) {
				t.Errorf("ValidateTopic(%q): want %v, got %v", c.input, c.want, got)
			}
		})
		t.Run("group/"+c.name, func(t *testing.T) {
			if c.name == "overlong_topic" {
				return // overlong_topic is a topic-only fixture here
			}
			if got := ValidateGroup(c.input); !errors.Is(got, c.want) {
				t.Errorf("ValidateGroup(%q): want %v, got %v", c.input, c.want, got)
			}
		})
	}

	// Boundary: a token exactly at the length cap is accepted (one byte over rejects).
	atMaxTopic := strings.Repeat("a", maxTopicLen)
	atMaxGroup := strings.Repeat("a", maxGroupLen)
	if err := ValidateTopic(atMaxTopic); err != nil {
		t.Errorf("ValidateTopic at maxTopicLen: want nil, got %v", err)
	}
	if err := ValidateGroup(atMaxGroup); err != nil {
		t.Errorf("ValidateGroup at maxGroupLen: want nil, got %v", err)
	}
}
