package mock

import (
	"context"
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
