package order

import (
	"context"
	"log/slog"
	"time"

	"cyber-ecosystem/shared-go/capability/payment"
	"cyber-ecosystem/shared-go/helper"
	"cyber-ecosystem/shared-go/utils"

	"cyber-ecosystem/app/services/esports/internal/ent"
	"cyber-ecosystem/app/services/esports/internal/ent/payorder"
	"cyber-ecosystem/app/services/esports/internal/ent/predicate"
	"cyber-ecosystem/app/services/esports/internal/platform"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

// Repo ----------------------------------------------------------------------------------------------------------------

type orderRP struct {
	shared.RP
}

func NewOrderRP(logger *slog.Logger, p *platform.Platform) OrderRP {
	return &orderRP{RP: shared.NewRP(logger.With("module", "module/order_rp"), p)}
}

type payChannelRP struct {
	shared.RP
}

func NewPayChannelRP(logger *slog.Logger, p *platform.Platform) PayChannelRP {
	return &payChannelRP{RP: shared.NewRP(logger.With("module", "module/order_rp"), p)}
}

// Method --------------------------------------------------------------------------------------------------------------

func (rp *orderRP) Create(ctx context.Context, o *Order) (*Order, error) {
	create := rp.Platform.GetClient(ctx).PayOrder.Create().
		SetOrderNo(o.OrderNo).
		SetUserID(o.UserID).
		SetSkuCode(o.SKUCode).
		SetScopeKind(payorder.ScopeKind(o.ScopeKind)).
		SetProvider(o.Provider).
		SetChannel(o.Channel).
		SetAmountFen(o.AmountFen).
		SetState(payorder.State(o.State)).
		SetCredentialKind(o.CredentialKind).
		SetCredentialValue(o.CredentialValue).
		SetExpiresAt(o.ExpiresAt)
	if o.ScopeRef != nil {
		create = create.SetScopeRef(*o.ScopeRef)
	}
	row, err := create.Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			// The pending-unique lost the same-scope race — the survivor
			// is the order to return; its stored credential is what the
			// payer continues with.
			if survivor, reErr := rp.ActiveByScope(ctx, o.UserID, o.ScopeKind, o.ScopeRef); reErr == nil && survivor != nil {
				return survivor, nil
			}
		}
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapOrder(row), nil
}

func (rp *orderRP) ClosePending(ctx context.Context, orderNo string) (*Order, error) {
	rows, err := rp.Platform.GetClient(ctx).PayOrder.Update().
		Where(payorder.OrderNoEQ(orderNo), payorder.StateEQ(payorder.StatePending)).
		SetState(payorder.StateClosed).
		Save(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	if rows == 0 {
		return nil, nil
	}
	return rp.ByNo(ctx, orderNo)
}

func (rp *orderRP) Settle(ctx context.Context, in *SettleIn) (*Order, bool, error) {
	// The conditional update is the whole arbitration: state IN
	// (pending, closed) → paid, where closed→paid is the channel-truth
	// correction for a close that raced a payment. Zero rows means a rival
	// settle already flipped it — the re-read hands the caller the winner.
	rows, err := rp.Platform.GetClient(ctx).PayOrder.Update().
		Where(
			payorder.OrderNoEQ(in.OrderNo),
			payorder.StateIn(payorder.StatePending, payorder.StateClosed),
		).
		SetState(payorder.StatePaid).
		SetTransactionID(in.TransactionID).
		SetPaidAt(in.PaidAt).
		Save(ctx)
	if err != nil {
		return nil, false, rp.Platform.HandleEntError(err)
	}
	order, err := rp.ByNo(ctx, in.OrderNo)
	if err != nil {
		return nil, false, err
	}
	return order, rows > 0, nil
}

func (rp *orderRP) RefundPaid(ctx context.Context, in *RefundIn) (*Order, bool, error) {
	// The refund twin of the settle arbitration: a paid, never-refunded row
	// takes its one full refund; zero rows means a rival refund already won
	// — the re-read hands the caller the winner (the same channel refund,
	// the out_refund_no being derived from the order).
	var refundID *string
	if in.RefundID != "" {
		refundID = &in.RefundID
	}
	rows, err := rp.Platform.GetClient(ctx).PayOrder.Update().
		Where(
			payorder.OrderNoEQ(in.OrderNo),
			payorder.StateEQ(payorder.StatePaid),
			payorder.RefundNoIsNil(),
		).
		SetRefundNo(in.RefundNo).
		SetNillableRefundID(refundID).
		SetRefundedFen(in.AmountFen).
		SetRefundReason(in.Reason).
		SetRefundState(payorder.RefundState(in.State)).
		SetNillableRefundedAt(in.RefundedAt).
		Save(ctx)
	if err != nil {
		return nil, false, rp.Platform.HandleEntError(err)
	}
	order, err := rp.ByNo(ctx, in.OrderNo)
	if err != nil {
		return nil, false, err
	}
	return order, rows > 0, nil
}

func (rp *orderRP) SettleRefund(ctx context.Context, in *RefundIn) (*Order, bool, error) {
	// Processing → terminal, arbitrated on the row's own refund number: a
	// rival converge that already folded the channel truth leaves zero
	// rows, and the re-read serves what it wrote.
	var refundID *string
	if in.RefundID != "" {
		refundID = &in.RefundID
	}
	rows, err := rp.Platform.GetClient(ctx).PayOrder.Update().
		Where(
			payorder.OrderNoEQ(in.OrderNo),
			payorder.RefundNoEQ(in.RefundNo),
			payorder.RefundStateEQ(payorder.RefundStateProcessing),
		).
		SetRefundState(payorder.RefundState(in.State)).
		SetNillableRefundID(refundID).
		SetNillableRefundedAt(in.RefundedAt).
		Save(ctx)
	if err != nil {
		return nil, false, rp.Platform.HandleEntError(err)
	}
	order, err := rp.ByNo(ctx, in.OrderNo)
	if err != nil {
		return nil, false, err
	}
	return order, rows > 0, nil
}

func (rp *orderRP) List(ctx context.Context, in *ListIn) (*ListOut, error) {
	// The member face narrows by its caller; the operator face arrives
	// with the filter set instead — unset fields simply add no predicate.
	conds := make([]predicate.PayOrder, 0, 5)
	if in.UserID != "" {
		conds = append(conds, payorder.UserIDEQ(in.UserID))
	}
	if len(in.UserIDs) > 0 {
		conds = append(conds, payorder.UserIDIn(in.UserIDs...))
	}
	if in.State != "" {
		conds = append(conds, payorder.StateEQ(payorder.State(in.State)))
	}
	if in.Scope != "" {
		conds = append(conds, payorder.ScopeKindEQ(payorder.ScopeKind(in.Scope)))
	}
	if in.Prefix != "" {
		conds = append(conds, payorder.OrderNoHasPrefix(in.Prefix))
	}
	query := rp.Platform.GetClient(ctx).PayOrder.Query().
		Where(conds...).
		Order(ent.Desc(payorder.FieldCreatedAt))
	total, offset, limit, err := helper.Paginate(ctx, query, in.PageRequest, helper.DefaultPageSizeMax)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	rows, err := query.All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return &ListOut{
		PageResponse: helper.BuildPageResponse(total, offset, limit),
		Orders:       utils.SliceMap(rows, mapOrder),
	}, nil
}

func (rp *orderRP) ListStalePending(ctx context.Context, olderThan time.Duration) ([]*Order, error) {
	rows, err := rp.Platform.GetClient(ctx).PayOrder.Query().
		Where(
			payorder.StateEQ(payorder.StatePending),
			payorder.CreatedAtLT(time.Now().Add(-olderThan)),
		).
		All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return utils.SliceMap(rows, mapOrder), nil
}

func (rp *orderRP) ListPaid(ctx context.Context, from, to time.Time) ([]*Order, error) {
	rows, err := rp.Platform.GetClient(ctx).PayOrder.Query().
		Where(
			payorder.StateEQ(payorder.StatePaid),
			payorder.PaidAtGTE(from),
			payorder.PaidAtLT(to),
		).
		All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return utils.SliceMap(rows, mapOrder), nil
}

func (rp *orderRP) ByNo(ctx context.Context, orderNo string) (*Order, error) {
	row, err := rp.Platform.GetClient(ctx).PayOrder.Query().
		Where(payorder.OrderNoEQ(orderNo)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapOrder(row), nil
}

func (rp *orderRP) ActiveByScope(ctx context.Context, userID, scopeKind string, scopeRef *string) (*Order, error) {
	query := rp.Platform.GetClient(ctx).PayOrder.Query().
		Where(
			payorder.UserIDEQ(userID),
			payorder.ScopeKindEQ(payorder.ScopeKind(scopeKind)),
			payorder.StateEQ(payorder.StatePending),
		).
		Order(ent.Desc(payorder.FieldCreatedAt))
	if scopeRef != nil {
		query = query.Where(payorder.ScopeRefEQ(*scopeRef))
	} else {
		// A period order's null ref: concurrent creations may leave more
		// than one live (NULLS DISTINCT never clashes) — newest wins.
		query = query.Where(payorder.ScopeRefIsNil())
	}
	row, err := query.First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapOrder(row), nil
}

func (rp *payChannelRP) CreatePayment(ctx context.Context, provider string, in *payment.CreatePaymentInput) (*payment.PrepayResult, error) {
	backend, err := rp.Platform.PaymentFor(provider)
	if err != nil {
		return nil, rp.Platform.HandlePaymentError(err)
	}
	out, err := backend.Gateway.CreatePayment(ctx, in)
	if err != nil {
		return nil, rp.Platform.HandlePaymentError(err)
	}
	return out, nil
}

func (rp *payChannelRP) ClosePayment(ctx context.Context, provider, outTradeNo string) error {
	backend, err := rp.Platform.PaymentFor(provider)
	if err != nil {
		return rp.Platform.HandlePaymentError(err)
	}
	if err := backend.Gateway.ClosePayment(ctx, outTradeNo); err != nil {
		return rp.Platform.HandlePaymentError(err)
	}
	return nil
}

func (rp *payChannelRP) QueryPayment(ctx context.Context, provider, outTradeNo string) (*payment.PaymentStatusResult, error) {
	backend, err := rp.Platform.PaymentFor(provider)
	if err != nil {
		return nil, rp.Platform.HandlePaymentError(err)
	}
	out, err := backend.Gateway.QueryPayment(ctx, outTradeNo)
	if err != nil {
		return nil, rp.Platform.HandlePaymentError(err)
	}
	return out, nil
}

func (rp *payChannelRP) Refund(ctx context.Context, provider string, in *payment.RefundInput) (*payment.RefundResult, error) {
	backend, err := rp.Platform.PaymentFor(provider)
	if err != nil {
		return nil, rp.Platform.HandlePaymentError(err)
	}
	out, err := backend.Gateway.Refund(ctx, in)
	if err != nil {
		return nil, rp.Platform.HandlePaymentError(err)
	}
	return out, nil
}

func (rp *payChannelRP) QueryRefund(ctx context.Context, provider, outRefundNo string) (*payment.RefundStatusResult, error) {
	backend, err := rp.Platform.PaymentFor(provider)
	if err != nil {
		return nil, rp.Platform.HandlePaymentError(err)
	}
	out, err := backend.Gateway.QueryRefund(ctx, outRefundNo)
	if err != nil {
		return nil, rp.Platform.HandlePaymentError(err)
	}
	return out, nil
}

func (rp *payChannelRP) ParseNotification(ctx context.Context, provider string, in *payment.NotificationInput) (*payment.NotificationEvent, error) {
	backend, err := rp.Platform.PaymentFor(provider)
	if err != nil {
		return nil, rp.Platform.HandlePaymentError(err)
	}
	event, err := backend.Notifier.Parse(ctx, in)
	if err != nil {
		return nil, rp.Platform.HandlePaymentError(err)
	}
	return event, nil
}

func (rp *payChannelRP) DownloadBill(ctx context.Context, provider string, in *payment.BillInput) (*payment.BillResult, error) {
	backend, err := rp.Platform.PaymentFor(provider)
	if err != nil {
		return nil, rp.Platform.HandlePaymentError(err)
	}
	bill, err := backend.Reports.DownloadBill(ctx, in)
	if err != nil {
		return nil, rp.Platform.HandlePaymentError(err)
	}
	return bill, nil
}

func (rp *payChannelRP) ParseTradeBill(ctx context.Context, provider string, raw []byte) ([]payment.BillLine, error) {
	backend, err := rp.Platform.PaymentFor(provider)
	if err != nil {
		return nil, rp.Platform.HandlePaymentError(err)
	}
	lines, err := backend.Reports.ParseTradeBill(ctx, raw)
	if err != nil {
		return nil, rp.Platform.HandlePaymentError(err)
	}
	return lines, nil
}

func (rp *payChannelRP) Channels(provider string) ([]payment.Channel, error) {
	channels, err := rp.Platform.Payments().ChannelsFor(provider)
	if err != nil {
		return nil, rp.Platform.HandlePaymentError(err)
	}
	return channels, nil
}

func (rp *payChannelRP) DefaultProvider() string { return rp.Platform.Payments().DefaultProvider() }

func (rp *payChannelRP) NotifyBaseURL() string { return rp.Platform.Payments().NotifyBaseURL() }

// Private -------------------------------------------------------------------------------------------------------------

func mapOrder(row *ent.PayOrder) *Order {
	o := &Order{
		OrderNo:         row.OrderNo,
		UserID:          row.UserID,
		SKUCode:         row.SkuCode,
		ScopeKind:       string(row.ScopeKind),
		ScopeRef:        row.ScopeRef,
		Provider:        row.Provider,
		Channel:         row.Channel,
		AmountFen:       row.AmountFen,
		State:           string(row.State),
		CredentialKind:  row.CredentialKind,
		CredentialValue: row.CredentialValue,
		TransactionID:   row.TransactionID,
		CreatedAt:       row.CreatedAt,
		ExpiresAt:       row.ExpiresAt,
		PaidAt:          row.PaidAt,
		RefundNo:        row.RefundNo,
		RefundID:        row.RefundID,
		RefundedFen:     row.RefundedFen,
		RefundReason:    row.RefundReason,
		RefundedAt:      row.RefundedAt,
	}
	if row.RefundState != nil {
		// The column's enum type narrows to the domain's plain string.
		state := string(*row.RefundState)
		o.RefundState = &state
	}
	return o
}
