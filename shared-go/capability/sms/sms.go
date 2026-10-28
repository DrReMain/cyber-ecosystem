package sms

import "context"

// Message is one templated SMS. SignName, TemplateCode and Params are the
// templated-SMS vocabulary the CN channels (Aliyun, Tencent) share: the
// template's own wording holds the message body, Params fills its
// placeholders. A channel with another shape (raw body, per-market
// routes) carries that knowledge in its own backend, never by widening
// this contract.
type Message struct {
	To           string
	SignName     string
	TemplateCode string
	Params       map[string]string
}

// SendResult is the channel's receipt: BizID is the provider's own
// send-receipt id, kept verbatim for audit correlation. Acceptance is not
// delivery — the channel took the message, nothing more.
type SendResult struct {
	BizID string
}

// Sender is the send surface. A send is a mutating call the caller never
// blind-retries: their own rate limiting is the throttle, and a timeout's
// unknown outcome is surfaced as ErrTimeout for exactly that decision.
type Sender interface {
	Send(ctx context.Context, msg *Message) (*SendResult, error)
}
