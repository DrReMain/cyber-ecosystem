package mock

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"cyber-ecosystem/shared-go/capability/payment"
)

// DownloadBill renders the current state as a CSV statement — the same
// shape contract as any backend's bill (raw bytes the consumer reconciles
// against). Mock serves today's snapshot immediately; a real channel
// publishes a day's statement the following day, a timing difference that
// belongs to the consumer's scheduling, not to this surface.
func (b *backend) DownloadBill(_ context.Context, in *payment.BillInput) (*payment.BillResult, error) {
	if err := payment.ValidateBillInput(in); err != nil {
		return nil, err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	var lines []string
	if in.Kind == payment.BillKindFundFlow {
		lines = append(lines, "out_trade_no,direction,amount_fen")
		for _, o := range b.orders {
			netChange := int64(o.in.AmountFen - o.refunded)
			if o.state != payment.PaymentStatePaid || netChange == 0 {
				continue
			}
			lines = append(lines, o.in.OutTradeNo+",net,"+strconv.FormatInt(netChange, 10))
		}
	} else {
		lines = append(lines, "out_trade_no,state,amount_fen,refunded_fen,paid_at")
		for _, o := range b.orders {
			lines = append(lines, strings.Join([]string{
				o.in.OutTradeNo,
				string(o.state),
				strconv.FormatInt(int64(o.in.AmountFen), 10),
				strconv.FormatInt(int64(o.refunded), 10),
				o.paidAt.Format(time.RFC3339),
			}, ","))
		}
	}
	return &payment.BillResult{Raw: []byte(strings.Join(lines, "\n"))}, nil
}

var billHeader = []string{"out_trade_no", "state", "amount_fen", "refunded_fen", "paid_at"}

// ParseTradeBill reads back the CSV this backend renders — the same
// round-trip a consumer's reconciliation runs against a real channel's
// statement. Malformed input folds onto the contract's sentinel, never a
// guessed row.
func (b *backend) ParseTradeBill(_ context.Context, raw []byte) ([]payment.BillLine, error) {
	rows := strings.Split(string(raw), "\n")
	idx := make(map[string]int, len(billHeader))
	for i, h := range strings.Split(rows[0], ",") {
		idx[h] = i
	}
	for _, need := range billHeader {
		if _, ok := idx[need]; !ok {
			return nil, fmt.Errorf("%w: mock bill header lacks %s", payment.ErrInvalidArgument, need)
		}
	}
	lines := make([]payment.BillLine, 0, len(rows)-1)
	for _, row := range rows[1:] {
		if row == "" {
			continue
		}
		rec := strings.Split(row, ",")
		if len(rec) != len(billHeader) {
			return nil, fmt.Errorf("%w: mock bill row shape", payment.ErrInvalidArgument)
		}
		var state payment.BillLineState
		switch payment.PaymentState(rec[idx["state"]]) {
		case payment.PaymentStatePending:
			state = payment.BillLinePending
		case payment.PaymentStatePaid:
			state = payment.BillLinePaid
		case payment.PaymentStateClosed:
			state = payment.BillLineClosed
		default:
			return nil, fmt.Errorf("%w: mock bill state %q", payment.ErrInvalidArgument, rec[idx["state"]])
		}
		amount, err := strconv.ParseInt(rec[idx["amount_fen"]], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%w: mock bill amount: %w", payment.ErrInvalidArgument, err)
		}
		refunded, err := strconv.ParseInt(rec[idx["refunded_fen"]], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%w: mock bill refunded: %w", payment.ErrInvalidArgument, err)
		}
		at, err := time.Parse(time.RFC3339, rec[idx["paid_at"]])
		if err != nil {
			return nil, fmt.Errorf("%w: mock bill time: %w", payment.ErrInvalidArgument, err)
		}
		lines = append(lines, payment.BillLine{
			OutTradeNo: rec[idx["out_trade_no"]],
			State:      state,
			AmountFen:  payment.Fen(amount),
			RefundFen:  payment.Fen(refunded),
			At:         at,
		})
	}
	return lines, nil
}
