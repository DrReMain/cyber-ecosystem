// Package dysmsapi is the sms family's Aliyun backend: a hand-rolled RPC
// client over the 2017-05-25 SendSms action (stdlib only — the official
// SDK drags a dependency tree in for one call). The wire protocol is the
// classic Aliyun RPC signature v1: sorted percent-encoded parameters,
// HMAC-SHA1 under secret&, the signature itself riding the POST form.
package dysmsapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"cyber-ecosystem/shared-go/capability/sms"
)

// Sender implements sms.Sender against the Aliyun Dysmsapi edge.
type Sender struct {
	cfg Config
}

type rpcResponse struct {
	RequestId string `json:"RequestId"`
	Code      string `json:"Code"`
	Message   string `json:"Message"`
	BizId     string `json:"BizId"`
}

func (s *Sender) Send(ctx context.Context, msg *sms.Message) (*sms.SendResult, error) {
	if err := sms.ValidateMessage(msg); err != nil {
		return nil, err
	}
	params, err := s.baseParams()
	if err != nil {
		return nil, err
	}
	params["PhoneNumbers"] = msg.To
	params["SignName"] = msg.SignName
	params["TemplateCode"] = msg.TemplateCode
	if len(msg.Params) > 0 {
		// The channel takes the template variables as one JSON object; a
		// string value is the only kind this family's vocabulary carries,
		// so the encoding is fixed here — no caller-side marshaling.
		enc, err := json.Marshal(msg.Params)
		if err != nil {
			return nil, fmt.Errorf("%w: template params: %w", sms.ErrInvalidArgument, err)
		}
		params["TemplateParam"] = string(enc)
	}
	sts := stringToSign(canonicalQuery(params))
	params["Signature"] = signStringToSign(s.cfg.AccessKeySecret, sts)

	form := make(url.Values, len(params))
	for k, v := range params {
		form.Set(k, v)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.Endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("%w: endpoint invalid: %w", sms.ErrUnavailable, err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := s.cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, classifyTransport(err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if err != nil {
		return nil, classifyTransport(err)
	}
	var out rpcResponse
	if err = json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("%w: status %d, unparseable body", sms.ErrUnavailable, resp.StatusCode)
	}
	// Business answers ride an HTTP 200; the channel's own Code is the
	// verdict. A non-200 with a parseable Code (signature errors,
	// throttling) is still a channel answer — the vocabulary survives.
	if out.Code != "OK" {
		return nil, &sms.ChannelError{Code: out.Code, Message: out.Message}
	}
	if out.BizId == "" {
		return nil, fmt.Errorf("%w: channel answered OK without a receipt id", sms.ErrUnavailable)
	}
	return &sms.SendResult{BizID: out.BizId}, nil
}

func (s *Sender) baseParams() (map[string]string, error) {
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		// Entropy exhaustion is the channel being down — the sibling mock
		// backend answers the same way.
		return nil, fmt.Errorf("%w: entropy failure: %w", sms.ErrUnavailable, err)
	}
	return map[string]string{
		"AccessKeyId":      s.cfg.AccessKeyID,
		"Action":           "SendSms",
		"Format":           "JSON",
		"RegionId":         s.cfg.RegionID,
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureNonce":   hex.EncodeToString(nonce[:]),
		"SignatureVersion": "1.0",
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"Version":          "2017-05-25",
	}, nil
}

// classifyTransport separates a deadline (the message's fate is unknown —
// the caller's limiter decides any resend) from a plain failure to reach
// the channel.
func classifyTransport(err error) error {
	if errors.Is(err, context.DeadlineExceeded) || os.IsTimeout(err) {
		return fmt.Errorf("%w: %w", sms.ErrTimeout, err)
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return fmt.Errorf("%w: %w", sms.ErrTimeout, err)
	}
	return fmt.Errorf("%w: %w", sms.ErrUnavailable, err)
}
