package wechat

// The trading surface over the APIv3 core: one create endpoint per access
// context, trade and refund queries converging on the neutral state model,
// close and refund as the two mutations. The channel's own states map onto
// the contract's; a state this build has never met refuses as unavailability
// rather than guessing what the caller might settle on.

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"cyber-ecosystem/shared-go/capability/payment"
)

// The access contexts this backend serves. Mock mirrors these strings
// declaration-for-declaration so a consumer's channel choice survives the
// dev-to-production swap.
const (
	ChannelJSAPI  payment.Channel = "wechat_jsapi"
	ChannelH5     payment.Channel = "wechat_h5"
	ChannelNative payment.Channel = "wechat_native"
)

// Channels lists the access contexts this backend serves.
func Channels() []payment.Channel {
	return []payment.Channel{ChannelJSAPI, ChannelH5, ChannelNative}
}

// enforceChannel is the provider-knowledge gate the contract does not
// carry: which channels exist and which Payer hints each requires.
func enforceChannel(in *payment.CreatePaymentInput) error {
	switch in.Channel {
	case ChannelJSAPI:
		if in.Payer == nil || in.Payer.OpenID == "" {
			return fmt.Errorf("%w: %s requires payer.openid", payment.ErrInvalidArgument, in.Channel)
		}
	case ChannelH5:
		if in.Payer == nil || in.Payer.ClientIP == "" {
			return fmt.Errorf("%w: %s requires payer.client_ip", payment.ErrInvalidArgument, in.Channel)
		}
	case ChannelNative:
		// A QR rendered on the page — no payer hint applies.
	default:
		return fmt.Errorf("%w: unsupported channel %q", payment.ErrInvalidArgument, in.Channel)
	}
	return nil
}

type gateway struct {
	c     *client
	appID string
}

func (g *gateway) CreatePayment(ctx context.Context, in *payment.CreatePaymentInput) (*payment.PrepayResult, error) {
	if err := payment.ValidateCreatePaymentInput(in); err != nil {
		return nil, err
	}
	if err := enforceChannel(in); err != nil {
		return nil, err
	}
	req := createOrderRequest{
		AppID:       g.appID,
		MchID:       g.c.cfg.MchID,
		Description: in.Description,
		OutTradeNo:  in.OutTradeNo,
		NotifyURL:   in.NotifyURL,
	}
	req.Amount.Total = int64(in.AmountFen)
	req.Amount.Currency = "CNY"
	if !in.ExpireAt.IsZero() {
		req.TimeExpire = in.ExpireAt.Format(time.RFC3339)
	}
	var path string
	cred := payment.Credential{}
	switch in.Channel {
	case ChannelJSAPI:
		path = "/v3/pay/transactions/jsapi"
		req.Payer = &payerWire{OpenID: in.Payer.OpenID}
		cred.Kind = payment.CredentialInvoke
	case ChannelH5:
		path = "/v3/pay/transactions/h5"
		req.SceneInfo = &sceneInfoWire{PayerClientIP: in.Payer.ClientIP}
		req.SceneInfo.H5Info.Type = "Wap"
		cred.Kind = payment.CredentialJumpURL
	case ChannelNative:
		path = "/v3/pay/transactions/native"
		cred.Kind = payment.CredentialQRCode
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("%w: encode order: %w", payment.ErrInvalidArgument, err)
	}
	var resp createOrderResponse
	if err := g.c.do(ctx, http.MethodPost, path, nil, body, &resp); err != nil {
		return nil, err
	}
	switch in.Channel {
	case ChannelJSAPI:
		cred.Value = resp.PrepayID
	case ChannelH5:
		cred.Value = resp.H5URL
	case ChannelNative:
		cred.Value = resp.CodeURL
	}
	return &payment.PrepayResult{OutTradeNo: in.OutTradeNo, Credential: cred}, nil
}

func (g *gateway) QueryPayment(ctx context.Context, outTradeNo string) (*payment.PaymentStatusResult, error) {
	query := url.Values{"mchid": {g.c.cfg.MchID}}
	var resp transactionResponse
	if err := g.c.do(ctx, http.MethodGet,
		"/v3/pay/transactions/out-trade-no/"+outTradeNo, query, nil, &resp); err != nil {
		return nil, err
	}
	state, err := mapTradeState(resp.TradeState)
	if err != nil {
		return nil, err
	}
	out := &payment.PaymentStatusResult{
		OutTradeNo:    outTradeNo,
		TransactionID: resp.TransactionID,
		State:         state,
		AmountFen:     payment.Fen(resp.Amount.Total),
		RefundedFen:   payment.Fen(resp.Amount.RefundTotal),
	}
	if resp.SuccessTime != "" {
		paidAt, err := time.Parse(time.RFC3339, resp.SuccessTime)
		if err != nil {
			return nil, fmt.Errorf("%w: success_time: %w", payment.ErrUnavailable, err)
		}
		out.PaidAt = paidAt
	}
	return out, nil
}

func (g *gateway) ClosePayment(ctx context.Context, outTradeNo string) error {
	// A static structure — encoding cannot fail; ignoring the error keeps
	// the close path free of an unreachable branch.
	body, _ := json.Marshal(closeOrderRequest{MchID: g.c.cfg.MchID})
	return g.c.do(ctx, http.MethodPost,
		"/v3/pay/transactions/out-trade-no/"+outTradeNo+"/close", nil, body, nil)
}

func (g *gateway) Refund(ctx context.Context, in *payment.RefundInput) (*payment.RefundResult, error) {
	if err := payment.ValidateRefundInput(in); err != nil {
		return nil, err
	}
	req := refundRequest{
		OutTradeNo:  in.OutTradeNo,
		OutRefundNo: in.OutRefundNo,
		Reason:      in.Reason,
		NotifyURL:   in.NotifyURL,
	}
	req.Amount.Refund = int64(in.AmountFen)
	req.Amount.Total = int64(in.TotalAmountFen)
	req.Amount.Currency = "CNY"
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("%w: encode refund: %w", payment.ErrInvalidArgument, err)
	}
	var resp refundResponse
	if err := g.c.do(ctx, http.MethodPost, "/v3/refund/domestic/refunds", nil, body, &resp); err != nil {
		return nil, err
	}
	state, err := mapRefundStatus(resp.Status)
	if err != nil {
		return nil, err
	}
	return &payment.RefundResult{OutRefundNo: in.OutRefundNo, RefundID: resp.RefundID, State: state}, nil
}

func (g *gateway) QueryRefund(ctx context.Context, outRefundNo string) (*payment.RefundStatusResult, error) {
	var resp refundResponse
	if err := g.c.do(ctx, http.MethodGet,
		"/v3/refund/domestic/refunds/"+outRefundNo, nil, nil, &resp); err != nil {
		return nil, err
	}
	state, err := mapRefundStatus(resp.Status)
	if err != nil {
		return nil, err
	}
	out := &payment.RefundStatusResult{
		OutRefundNo: outRefundNo,
		RefundID:    resp.RefundID,
		State:       state,
		AmountFen:   payment.Fen(resp.Amount.Refund),
	}
	if resp.SuccessTime != "" {
		at, err := time.Parse(time.RFC3339, resp.SuccessTime)
		if err != nil {
			return nil, fmt.Errorf("%w: success_time: %w", payment.ErrUnavailable, err)
		}
		out.SuccessAt = at
	}
	return out, nil
}

// mapTradeState folds the channel's trade states onto the contract's three.
// REFUND is a paid order with refunds behind it — a derived fact the status
// carries as RefundedFen, not a state of its own.
func mapTradeState(s string) (payment.PaymentState, error) {
	switch s {
	case "SUCCESS", "REFUND":
		return payment.PaymentStatePaid, nil
	case "NOTPAY", "USERPAYING":
		return payment.PaymentStatePending, nil
	case "CLOSED", "REVOKED", "PAYERROR":
		return payment.PaymentStateClosed, nil
	default:
		return "", fmt.Errorf("%w: unknown trade_state %q", payment.ErrUnavailable, s)
	}
}

// mapRefundStatus folds the channel's refund states; CLOSED is the
// channel's terminal refund failure and lands in the manual-intervention
// bucket beside ABNORMAL.
func mapRefundStatus(s string) (payment.RefundState, error) {
	switch s {
	case "PROCESSING":
		return payment.RefundStateProcessing, nil
	case "SUCCESS":
		return payment.RefundStateSucceeded, nil
	case "ABNORMAL", "CLOSED":
		return payment.RefundStateAbnormal, nil
	default:
		return "", fmt.Errorf("%w: unknown refund status %q", payment.ErrUnavailable, s)
	}
}

type createOrderRequest struct {
	AppID       string         `json:"appid"`
	MchID       string         `json:"mchid"`
	Description string         `json:"description"`
	OutTradeNo  string         `json:"out_trade_no"`
	TimeExpire  string         `json:"time_expire,omitempty"`
	NotifyURL   string         `json:"notify_url"`
	Amount      amountWire     `json:"amount"`
	Payer       *payerWire     `json:"payer,omitempty"`
	SceneInfo   *sceneInfoWire `json:"scene_info,omitempty"`
}

type payerWire struct {
	OpenID string `json:"openid"`
}

type sceneInfoWire struct {
	PayerClientIP string     `json:"payer_client_ip"`
	H5Info        h5InfoWire `json:"h5_info"`
}

type h5InfoWire struct {
	Type string `json:"type"`
}

type amountWire struct {
	Total    int64  `json:"total"`
	Currency string `json:"currency"`
}

type createOrderResponse struct {
	PrepayID string `json:"prepay_id"`
	H5URL    string `json:"h5_url"`
	CodeURL  string `json:"code_url"`
}

type closeOrderRequest struct {
	MchID string `json:"mchid"`
}

type transactionResponse struct {
	TransactionID string `json:"transaction_id"`
	TradeState    string `json:"trade_state"`
	SuccessTime   string `json:"success_time"`
	Amount        struct {
		Total       int64 `json:"total"`
		RefundTotal int64 `json:"refund_total"`
	} `json:"amount"`
}

type refundRequest struct {
	OutTradeNo  string `json:"out_trade_no"`
	OutRefundNo string `json:"out_refund_no"`
	Reason      string `json:"reason,omitempty"`
	NotifyURL   string `json:"notify_url,omitempty"`
	Amount      struct {
		Refund   int64  `json:"refund"`
		Total    int64  `json:"total"`
		Currency string `json:"currency"`
	} `json:"amount"`
}

type refundResponse struct {
	RefundID    string `json:"refund_id"`
	Status      string `json:"status"`
	SuccessTime string `json:"success_time,omitempty"`
	Amount      struct {
		Refund int64 `json:"refund"`
	} `json:"amount"`
}
