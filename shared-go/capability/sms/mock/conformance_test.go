package mock

// Conformance probes over the mock backend: the family contract's round
// trip (send → receipt id), the shape check running before anything is
// recorded, the recording order under concurrency, and the controller
// recovery. The memory is the environment — the full list runs with no
// external setup.

import (
	"context"
	"errors"
	"sync"
	"testing"

	"cyber-ecosystem/shared-go/capability/sms"
)

func codeMsg(code string) *sms.Message {
	return &sms.Message{
		To:           "13800138000",
		SignName:     "曦光未来",
		TemplateCode: "SMS_460000000",
		Params:       map[string]string{"code": code},
	}
}

// TestSendRecordsEntry: one send produces exactly one entry, the message
// verbatim, and answers a receipt id distinct from the next send's.
func TestSendRecordsEntry(t *testing.T) {
	s, ctrl := Open()
	r1, err := s.Send(context.Background(), codeMsg("111111"))
	if err != nil {
		t.Fatalf("send 1: %v", err)
	}
	r2, err := s.Send(context.Background(), codeMsg("222222"))
	if err != nil {
		t.Fatalf("send 2: %v", err)
	}
	sent := ctrl.Sent()
	if len(sent) != 2 {
		t.Fatalf("recorded %d sends, want 2", len(sent))
	}
	if sent[0].Params["code"] != "111111" || sent[1].Params["code"] != "222222" {
		t.Fatalf("codes recorded out of order: %+v", sent)
	}
	if sent[0].To != "13800138000" || sent[0].SignName != "曦光未来" || sent[0].TemplateCode != "SMS_460000000" {
		t.Fatalf("entry not verbatim: %+v", sent[0])
	}
	if r1.BizID == "" || r1.BizID == r2.BizID {
		t.Fatalf("receipt ids must be present and distinct: %q vs %q", r1.BizID, r2.BizID)
	}
}

// TestSendRejectsBeforeRecording: the family shape check runs inside Send —
// an invalid message answers ErrInvalidArgument and leaves the recording
// empty, so a misconfigured caller never poisons the harness's reads.
func TestSendRejectsBeforeRecording(t *testing.T) {
	s, ctrl := Open()
	bad := codeMsg("123456")
	bad.To = "123"
	if _, err := s.Send(context.Background(), bad); !errors.Is(err, sms.ErrInvalidArgument) {
		t.Fatalf("want ErrInvalidArgument, got %v", err)
	}
	if n := len(ctrl.Sent()); n != 0 {
		t.Fatalf("invalid send recorded %d entries, want 0", n)
	}
}

// TestConcurrentSendsAllRecorded: the recorder is the shared state — every
// concurrent send must land, none lost to a race.
func TestConcurrentSendsAllRecorded(t *testing.T) {
	s, ctrl := Open()
	const n = 50
	var wg sync.WaitGroup
	for range n {
		wg.Go(func() {
			if _, err := s.Send(context.Background(), codeMsg("000000")); err != nil {
				t.Errorf("concurrent send: %v", err)
			}
		})
	}
	wg.Wait()
	if got := len(ctrl.Sent()); got != n {
		t.Fatalf("recorded %d of %d concurrent sends", got, n)
	}
}

// TestControllerOfRejectsForeignSender: the e2e recovery only succeeds on
// a backend this package built — a swapped-in real channel must report
// false, never panic.
func TestControllerOfRejectsForeignSender(t *testing.T) {
	if _, ok := ControllerOf(foreign{}); ok {
		t.Fatal("ControllerOf must refuse a foreign sender")
	}
}

type foreign struct{}

func (foreign) Send(context.Context, *sms.Message) (*sms.SendResult, error) {
	return &sms.SendResult{}, nil
}

// TestNewRecordsWithoutController: the production face (New) still records
// — a later ControllerOf recovers the same history, so dev and test see
// one backend, not two.
func TestNewRecordsWithoutController(t *testing.T) {
	s := New()
	if _, err := s.Send(context.Background(), codeMsg("333333")); err != nil {
		t.Fatalf("send: %v", err)
	}
	ctrl, ok := ControllerOf(s)
	if !ok {
		t.Fatal("ControllerOf over New's backend must succeed")
	}
	if got := ctrl.Sent(); len(got) != 1 || got[0].Params["code"] != "333333" {
		t.Fatalf("history lost between New and ControllerOf: %+v", got)
	}
}
