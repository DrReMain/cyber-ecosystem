package mq

import (
	"errors"
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

func TestValidateTopicGroup(t *testing.T) {
	// rejected: empty, over-long, or any byte unsafe as a NATS subject literal or
	// stream/consumer-name fragment (the stream name is "mq-"+topic and the
	// durable name is group+"-"+topic, so '.', '>', '*', '/', whitespace and
	// control chars are illegal).
	for _, bad := range []string{
		"", "orders.events", "a b", "a.>", "a*", "a/b", "中文", "x\ty", "tab\t",
	} {
		if err := ValidateTopic(bad); err != ErrInvalidArgument {
			t.Errorf("ValidateTopic(%q): want ErrInvalidArgument, got %v", bad, err)
		}
	}
	// valid tokens
	for _, ok := range []string{"orders_events", "orders-events", "topic123", "ABC"} {
		if err := ValidateTopic(ok); err != nil {
			t.Errorf("ValidateTopic(%q): want nil, got %v", ok, err)
		}
	}
	if err := ValidateGroup(""); err != ErrInvalidArgument {
		t.Errorf("empty group: %v", err)
	}
	if err := ValidateGroup("notify-svc"); err != nil {
		t.Errorf("valid group: %v", err)
	}
	if err := ValidateGroup("notify.svc"); err != ErrInvalidArgument {
		t.Errorf("group with '.' should be rejected: %v", err)
	}
}
