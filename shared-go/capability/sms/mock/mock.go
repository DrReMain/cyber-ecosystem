// Package mock is the sms family's local backend: it records every message
// and never dials. It exists so dev and e2e stacks run the whole send
// chain with zero channel credentials — the code lands in the service log
// and in the controller the test harness holds.
package mock

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"cyber-ecosystem/shared-go/capability/sms"
)

// Entry is one recorded send, the message verbatim — the dev/e2e face's
// whole point is that the code is readable here.
type Entry struct {
	To           string
	SignName     string
	TemplateCode string
	Params       map[string]string
	BizID        string
	At           time.Time
}

type backend struct {
	mu      sync.Mutex
	entries []Entry
}

func newBackend() *backend {
	return &backend{}
}

// New returns a recording sender. Production consumers reach it only when
// the platform found no credentialed channel — the same single face the
// real backends implement.
func New() sms.Sender {
	return newBackend()
}

// Open returns the sender plus its Controller (read the recorded sends).
// Tests and dev harnesses use Open; production code uses New.
func Open() (sms.Sender, *Controller) {
	b := newBackend()
	return b, &Controller{b: b}
}

// ControllerOf recovers the Controller from a sender this package built —
// the e2e face over a backend a platform assembled from config, where
// production code reaches only the contract.
func ControllerOf(s sms.Sender) (*Controller, bool) {
	b, ok := s.(*backend)
	if !ok {
		return nil, false
	}
	return &Controller{b: b}, true
}

func (b *backend) Send(_ context.Context, msg *sms.Message) (*sms.SendResult, error) {
	if err := sms.ValidateMessage(msg); err != nil {
		return nil, err
	}
	var nonce [4]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return nil, fmt.Errorf("%w: entropy failure: %w", sms.ErrUnavailable, err)
	}
	e := Entry{
		To:           msg.To,
		SignName:     msg.SignName,
		TemplateCode: msg.TemplateCode,
		Params:       msg.Params,
		BizID:        "log-" + hex.EncodeToString(nonce[:]),
		At:           time.Now(),
	}
	b.mu.Lock()
	b.entries = append(b.entries, e)
	b.mu.Unlock()
	// The params carry the verification code by design; this backend only
	// runs where no real channel is configured, and the code in the log is
	// the dev loop's whole value.
	slog.Default().Info("sms mock backend: sent",
		"to", e.To, "sign", e.SignName, "template", e.TemplateCode,
		"params", fmt.Sprintf("%v", e.Params), "biz_id", e.BizID)
	return &sms.SendResult{BizID: e.BizID}, nil
}

type Controller struct {
	b *backend
}

// Sent returns a copy of the recorded sends in order.
func (c *Controller) Sent() []Entry {
	c.b.mu.Lock()
	defer c.b.mu.Unlock()
	out := make([]Entry, len(c.b.entries))
	copy(out, c.b.entries)
	return out
}
