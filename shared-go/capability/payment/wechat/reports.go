package wechat

// The statement face: the channel hands out a signed download ticket for a
// statement day, and the bill itself rides a separate unsigned host — the
// protocol's documented exception, whose integrity rides the hash the ticket
// carried instead of a signature.

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"cyber-ecosystem/shared-go/capability/payment"
)

type reports struct {
	c *client
}

func (r *reports) DownloadBill(ctx context.Context, in *payment.BillInput) (*payment.BillResult, error) {
	if err := payment.ValidateBillInput(in); err != nil {
		return nil, err
	}
	var path string
	switch in.Kind {
	case payment.BillKindTrade:
		path = "/v3/bill/tradebill"
	case payment.BillKindFundFlow:
		path = "/v3/bill/fundflowbill"
	default:
		// ValidateBillInput narrows the kind set today; this arm is the
		// provider-side guard for the day the contract grows a kind this
		// backend does not serve.
		return nil, fmt.Errorf("%w: unknown bill kind %q", payment.ErrInvalidArgument, in.Kind)
	}
	query := url.Values{
		"bill_date":    {in.Date.Format("2006-01-02")},
		"account_type": {"BASIC"},
	}
	var ticket billTicket
	if err := r.c.do(ctx, http.MethodGet, path, query, nil, &ticket); err != nil {
		return nil, err
	}
	if ticket.DownloadURL == "" {
		return nil, fmt.Errorf("%w: channel returned no download url", payment.ErrUnavailable)
	}
	raw, err := r.c.fetchBill(ctx, ticket.DownloadURL)
	if err != nil {
		return nil, err
	}
	if ticket.HashValue != "" {
		sum := sha256.Sum256(raw)
		if !strings.EqualFold(hex.EncodeToString(sum[:]), ticket.HashValue) {
			return nil, fmt.Errorf("%w: bill hash mismatch", payment.ErrInvalidSignature)
		}
	}
	return &payment.BillResult{Raw: raw}, nil
}

// fetchBill retrieves the bill bytes from the channel-issued absolute URL.
// The download host is not the API host, so the signature covers the URI
// only; the response carries no Wechatpay signature by protocol design, and
// a 2xx body that is JSON is the channel's way of wrapping a late error
// (an expired ticket) in a success status.
func (c *client) fetchBill(ctx context.Context, downloadURL string) ([]byte, error) {
	u, err := url.Parse(downloadURL)
	if err != nil {
		return nil, fmt.Errorf("%w: parse download url: %w", payment.ErrInvalidArgument, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: build request: %w", payment.ErrInvalidArgument, err)
	}
	ts := time.Now().Unix()
	nonce, err := newNonce()
	if err != nil {
		return nil, fmt.Errorf("%w: nonce: %w", payment.ErrUnavailable, err)
	}
	sig, err := SignSHA256RSA(c.mchKey, BuildRequestMessage(http.MethodGet, u.RequestURI(), ts, nonce, nil))
	if err != nil {
		return nil, fmt.Errorf("%w: sign: %w", payment.ErrInvalidArgument, err)
	}
	req.Header.Set("Authorization", BuildAuthorizationHeader(c.cfg.MchID, c.cfg.MchSerial, ts, nonce, sig))
	req.Header.Set("Accept", "application/json")
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", payment.ErrUnavailable, err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: read response: %w", payment.ErrUnavailable, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		a := &APIError{Status: resp.StatusCode, RequestID: resp.Header.Get("Request-ID")}
		var e struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		if json.Unmarshal(body, &e) == nil {
			a.Code, a.Message = e.Code, e.Message
		}
		return nil, classifyAPIError(a)
	}
	if strings.Contains(resp.Header.Get("Content-Type"), "json") {
		var e struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		if json.Unmarshal(body, &e) == nil && e.Code != "" {
			return nil, fmt.Errorf("%w: %s: %s", payment.ErrUnavailable, e.Code, e.Message)
		}
	}
	if strings.Contains(resp.Header.Get("Content-Type"), "gzip") {
		zr, err := gzip.NewReader(bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("%w: bill gzip: %w", payment.ErrUnavailable, err)
		}
		defer func() { _ = zr.Close() }()
		plain, err := io.ReadAll(zr)
		if err != nil {
			return nil, fmt.Errorf("%w: bill gunzip: %w", payment.ErrUnavailable, err)
		}
		return plain, nil
	}
	return body, nil
}

type billTicket struct {
	DownloadURL string `json:"download_url"`
	HashValue   string `json:"hash_value"`
	HashType    string `json:"hash_type"`
}
