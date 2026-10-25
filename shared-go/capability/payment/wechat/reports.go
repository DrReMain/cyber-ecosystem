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
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"cyber-ecosystem/shared-go/capability/payment"
)

// cstZone is the channel's statement timezone — bill days and their row
// stamps are wall-clock facts in UTC+8.
var cstZone = time.FixedZone("CST", 8*60*60)

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

// The trade bill's column vocabulary — looked up by name, never by position,
// so a column the channel appends or reorders cannot silently shift meaning.
const (
	colBillTime      = "交易时间"
	colTransactionID = "微信订单号"
	colOutTradeNo    = "商户订单号"
	colTradeState    = "交易状态"
	colRefundAmount  = "退款金额"
	colFeeAmount     = "手续费金额"
	colOrderAmount   = "订单金额"
	colSettlementAmt = "应结订单金额"
	billTimeLayout   = "2006-01-02 15:04:05"
)

// ParseTradeBill neutralizes the (already gunzipped) trade statement CSV
// into contract rows. Every wire word outside the mapped state set and
// every amount off the yuan-dot-fen format is ErrInvalidArgument — a
// statement this parser does not understand is data to flag, never to
// guess at.
func (r *reports) ParseTradeBill(_ context.Context, raw []byte) ([]payment.BillLine, error) {
	body := raw
	if bytes.HasPrefix(body, []byte("\ufeff")) {
		body = body[len("\ufeff"):]
	}
	recs, err := csv.NewReader(bytes.NewReader(body)).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("%w: bill csv: %w", payment.ErrInvalidArgument, err)
	}
	if len(recs) < 1 {
		return nil, fmt.Errorf("%w: bill has no header row", payment.ErrInvalidArgument)
	}
	col, err := billColumnIndex(recs[0])
	if err != nil {
		return nil, err
	}
	lines := make([]payment.BillLine, 0, len(recs)-1)
	for _, rec := range recs[1:] {
		if len(rec) < len(recs[0]) {
			return nil, fmt.Errorf("%w: bill row shorter than header", payment.ErrInvalidArgument)
		}
		state, err := mapBillState(rec[col[colTradeState]])
		if err != nil {
			return nil, err
		}
		amount, err := parseYuanFen(rec[col[colOrderAmount]])
		if err != nil {
			return nil, err
		}
		refund, err := parseYuanFen(rec[col[colRefundAmount]])
		if err != nil {
			return nil, err
		}
		fee, err := parseYuanFen(rec[col[colFeeAmount]])
		if err != nil {
			return nil, err
		}
		income, err := parseYuanFen(rec[col[colSettlementAmt]])
		if err != nil {
			return nil, err
		}
		at, err := time.ParseInLocation(billTimeLayout, rec[col[colBillTime]], cstZone)
		if err != nil {
			return nil, fmt.Errorf("%w: bill time %q: %w", payment.ErrInvalidArgument, rec[col[colBillTime]], err)
		}
		lines = append(lines, payment.BillLine{
			OutTradeNo:    rec[col[colOutTradeNo]],
			TransactionID: rec[col[colTransactionID]],
			State:         state,
			AmountFen:     amount,
			RefundFen:     refund,
			FeeFen:        fee,
			IncomeFen:     income,
			At:            at,
		})
	}
	return lines, nil
}

func billColumnIndex(header []string) (map[string]int, error) {
	col := make(map[string]int, len(header))
	for i, h := range header {
		col[strings.TrimSpace(h)] = i
	}
	for _, need := range []string{colBillTime, colTransactionID, colOutTradeNo, colTradeState, colRefundAmount, colFeeAmount, colOrderAmount, colSettlementAmt} {
		if _, ok := col[need]; !ok {
			return nil, fmt.Errorf("%w: bill header lacks column %s", payment.ErrInvalidArgument, need)
		}
	}
	return col, nil
}

func mapBillState(word string) (payment.BillLineState, error) {
	switch strings.TrimSpace(word) {
	case "SUCCESS":
		return payment.BillLinePaid, nil
	case "REFUND":
		return payment.BillLineRefund, nil
	case "REVOKED", "CLOSED", "PAYERROR":
		return payment.BillLineClosed, nil
	case "USERPAYING":
		return payment.BillLinePending, nil
	default:
		return "", fmt.Errorf("%w: bill trade state %q", payment.ErrInvalidArgument, word)
	}
}

// parseYuanFen reads the channel's yuan-dot-fen decimal string ("19.90",
// possibly negative for reversal rows) into fen on integer arithmetic — no
// float ever touches an amount.
func parseYuanFen(s string) (payment.AmountFen, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	neg := false
	if strings.HasPrefix(s, "-") {
		neg, s = true, s[1:]
	}
	yuan, frac, dotted := strings.Cut(s, ".")
	// The channel's amounts are plain digit strings — a signed or partial
	// shape ("-1" parsing to a positive, "+1", "5.", ".5") is a corruption
	// of the cell, not a value to guess.
	if !allDigits(yuan) || (dotted && (!allDigits(frac) || len(frac) > 2)) {
		return 0, fmt.Errorf("%w: bill amount %q", payment.ErrInvalidArgument, s)
	}
	frac += strings.Repeat("0", 2-len(frac))
	y, err := strconv.ParseInt(yuan, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: bill amount %q: %w", payment.ErrInvalidArgument, s, err)
	}
	f, err := strconv.ParseInt(frac, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: bill amount %q: %w", payment.ErrInvalidArgument, s, err)
	}
	if y > (math.MaxInt64-f)/100 {
		return 0, fmt.Errorf("%w: bill amount %q overflows", payment.ErrInvalidArgument, s)
	}
	n := y*100 + f
	if neg {
		n = -n
	}
	return payment.Fen(n), nil
}

func allDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
