package pay

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/rs/xid"

	"cyber-ecosystem/shared-go/capability/payment"
	"cyber-ecosystem/shared-go/kratos/security"
	"cyber-ecosystem/shared-go/utils"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
	commonv1 "cyber-ecosystem/gen/go/cyber/shared/common/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"

	"cyber-ecosystem/app/services/esports/internal/module/test"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

const (
	SkuKindTrial      = "trial"
	SkuKindMembership = "membership"

	ScopeTrial    = "trial"      // order: unlocks one test record permanently
	ScopePeriod   = "period"     // order: opens the membership window
	EntTrial      = "trial"      // entitlement: one record, no deadline
	EntMembership = "membership" // entitlement: a window with an expiry

	StatePending = "pending"
	StatePaid    = "paid"
	StateClosed  = "closed"

	orderNoPrefix = "P"
	// The channel's own default deadline, carried explicitly so the row and
	// the channel order can never disagree about expiry.
	payWindow = 2 * time.Hour

	convergeInterval   = time.Minute
	convergeStaleAfter = 10 * time.Minute
	reconcileInterval  = 24 * time.Hour
)

var billZone = time.FixedZone("CST", 8*60*60) // bill days are wall-clock facts in the channel's timezone, not the host's — WeChat publishes in Beijing time; the mock's day-agnostic snapshot takes the same window so both providers reconcile under one rule

// DO ------------------------------------------------------------------------------------------------------------------

type Sku struct {
	Code         string
	Kind         string
	Name         string
	Description  *string
	PriceFen     int64
	ListPriceFen int64
	PeriodDays   *int
	OnShelf      bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Order struct {
	OrderNo         string
	UserID          string
	SKUCode         string
	ScopeKind       string
	ScopeRef        *string
	Provider        string
	Channel         string
	AmountFen       int64
	State           string
	CredentialKind  string
	CredentialValue string
	TransactionID   *string
	CreatedAt       time.Time
	ExpiresAt       time.Time
	PaidAt          *time.Time
}

type CreateIn struct {
	SKUCode     string
	Channel     payment.Channel
	ScopeRef    *string
	PayerOpenID *string
	ReturnURL   *string
}

type CreateOut struct {
	Order      *Order
	Credential payment.Credential
}

type ListIn struct {
	*commonv1.PageRequest
	UserID string
}

type ListOut struct {
	*commonv1.PageResponse
	Orders []*Order
}

type SettleIn struct {
	OrderNo       string
	Provider      string
	TransactionID string
	AmountFen     int64
	PaidAt        time.Time
}

type EntitlementGrant struct {
	UserID    string
	ScopeKind string
	ScopeRef  *string
	ExpiresAt *time.Time
	Source    string
}

// Port ----------------------------------------------------------------------------------------------------------------

type SkuRP interface {
	List(ctx context.Context) ([]*Sku, error)              // on-shelf face for the member plane
	ByCode(ctx context.Context, code string) (*Sku, error) // shelf ignored — nil, nil = no such code
}

type OrderRP interface {
	Create(ctx context.Context, o *Order) (*Order, error)                                          // a pending-unique conflict returns the surviving active order
	ByNo(ctx context.Context, orderNo string) (*Order, error)                                      // nil, nil = no such order
	ActiveByScope(ctx context.Context, userID, scopeKind string, scopeRef *string) (*Order, error) // nil, nil = none live
	ClosePending(ctx context.Context, orderNo string) (*Order, error)                              // nil, nil = no longer pending
	Settle(ctx context.Context, in *SettleIn) (*Order, bool, error)                                // the single write-paid entry; false = a rival settle already won
	List(ctx context.Context, in *ListIn) (*ListOut, error)
	ListStalePending(ctx context.Context, olderThan time.Duration) ([]*Order, error) // pending rows past the horizon; expiry is the converger's branch, not the query's
	ListPaid(ctx context.Context, from, to time.Time) ([]*Order, error)              // paid rows by paid_at window — the reconciliation's local side
}

type EntitlementRP interface {
	Grant(ctx context.Context, g *EntitlementGrant) error // insert-only; a rival grant for the same scope lands as a no-op, never an error
	Exists(ctx context.Context, userID, scopeKind string, scopeRef *string) (bool, error)
}

type PayChannelRP interface { // routes to the configured provider's backend; capability errors arrive classified
	CreatePayment(ctx context.Context, provider string, in *payment.CreatePaymentInput) (*payment.PrepayResult, error)
	QueryPayment(ctx context.Context, provider, outTradeNo string) (*payment.PaymentStatusResult, error)
	ClosePayment(ctx context.Context, provider, outTradeNo string) error
	ParseNotification(ctx context.Context, provider string, in *payment.NotificationInput) (*payment.NotificationEvent, error)
	DownloadBill(ctx context.Context, provider string, in *payment.BillInput) (*payment.BillResult, error) // the day's raw statement; NotFound = the channel has not published it yet
	ParseTradeBill(ctx context.Context, provider string, raw []byte) ([]payment.BillLine, error)           // neutralized statement rows in the contract's vocabulary
	Channels(provider string) ([]payment.Channel, error)
	DefaultProvider() string
	NotifyBaseURL() string
}

// UC ------------------------------------------------------------------------------------------------------------------

type SkuUC struct {
	shared.UC
	skuRP SkuRP
}

func NewSkuUC(logger *slog.Logger, tm shared.Transaction, skuRP SkuRP) *SkuUC {
	return &SkuUC{UC: shared.NewUC(logger.With("module", "module/pay"), tm), skuRP: skuRP}
}

type OrderUC struct {
	shared.UC
	orderRP       OrderRP
	skuRP         SkuRP
	entRP         EntitlementRP
	testRP        test.TestRP
	channelRP     PayChannelRP
	loopCancels   []context.CancelFunc
	nextBillDay   time.Time // the day the next reconcile sweep must still see; owned by the reconcile loop's goroutine
	nextBillRetry int
}

func NewOrderUC(logger *slog.Logger, tm shared.Transaction, hooks shared.HookRegistry, orderRP OrderRP, skuRP SkuRP, entRP EntitlementRP, testRP test.TestRP, channelRP PayChannelRP) *OrderUC {
	uc := &OrderUC{
		UC:        shared.NewUC(logger.With("module", "module/pay"), tm),
		orderRP:   orderRP,
		skuRP:     skuRP,
		entRP:     entRP,
		testRP:    testRP,
		channelRP: channelRP,
	}
	hooks.OnStart(uc.runLoop("converge", convergeInterval, uc.ConvergeOnce))
	hooks.OnStart(uc.runLoop("reconcile", reconcileInterval, uc.reconcileSweep))
	hooks.OnStop(uc.stopLoops)
	return uc
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *SkuUC) List(ctx context.Context) ([]*Sku, error) {
	return uc.skuRP.List(ctx)
}

func (uc *OrderUC) Create(ctx context.Context, in *CreateIn) (*CreateOut, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("order create requires a session"))
	}
	sku, err := uc.skuRP.ByCode(ctx, in.SKUCode)
	if err != nil {
		return nil, err
	}
	if sku == nil {
		return nil, esportspb.ErrorEsportsSkuNotFound("")
	}
	if !sku.OnShelf {
		return nil, esportspb.ErrorEsportsSkuOffShelf("")
	}
	provider := uc.channelRP.DefaultProvider()
	scopeKind, scopeRef, err := uc.resolveScope(ctx, subject.UserID, sku, in.ScopeRef)
	if err != nil {
		return nil, err
	}
	channels, err := uc.channelRP.Channels(provider)
	if err != nil {
		return nil, err
	}
	if !channelIn(channels, in.Channel) {
		return nil, esportspb.ErrorEsportsOrderChannelInvalid("").WithCause(
			fmt.Errorf("provider %s does not serve channel %s", provider, in.Channel))
	}
	if active, err := uc.orderRP.ActiveByScope(ctx, subject.UserID, scopeKind, scopeRef); err != nil {
		return nil, err
	} else if active != nil {
		if active.ExpiresAt.After(time.Now()) {
			// Same-order reuse: the stored credential replays with zero
			// channel round trips, so the channel never sees a second
			// create for the same grant target.
			return &CreateOut{Order: active, Credential: credentialOf(active)}, nil
		}
		if _, err := uc.retire(ctx, active); err != nil {
			return nil, err
		}
	}
	orderNo := orderNoPrefix + xid.New().String()
	expireAt := time.Now().Add(payWindow)
	var payer *payment.Payer
	if in.PayerOpenID != nil || in.ReturnURL != nil {
		payer = &payment.Payer{
			OpenID:    utils.Deref(in.PayerOpenID, ""),
			ReturnURL: utils.Deref(in.ReturnURL, ""),
		}
	}
	prepay, err := uc.channelRP.CreatePayment(ctx, provider, &payment.CreatePaymentInput{
		Channel:     in.Channel,
		OutTradeNo:  orderNo,
		AmountFen:   payment.Fen(sku.PriceFen),
		Description: sku.Name,
		NotifyURL:   uc.channelRP.NotifyBaseURL() + fmt.Sprintf("/api/pay/%s/notify", provider),
		ExpireAt:    expireAt,
		Payer:       payer,
	})
	if err != nil {
		return nil, err
	}
	order := &Order{
		OrderNo:         orderNo,
		UserID:          subject.UserID,
		SKUCode:         sku.Code,
		ScopeKind:       scopeKind,
		ScopeRef:        scopeRef,
		Provider:        provider,
		Channel:         string(in.Channel),
		AmountFen:       sku.PriceFen,
		State:           StatePending,
		CredentialKind:  string(prepay.Credential.Kind),
		CredentialValue: prepay.Credential.Value,
		ExpiresAt:       expireAt,
	}
	created, err := uc.orderRP.Create(ctx, order)
	if err != nil {
		return nil, err
	}
	if created.OrderNo != orderNo {
		// Lost the pending-unique race — the survivor replaced ours. The
		// loser's channel order is an orphan nobody can pay into knowingly;
		// close it best-effort, and it expires on its own either way.
		if cerr := uc.channelRP.ClosePayment(ctx, provider, orderNo); cerr != nil {
			uc.Log.Warn("close orphan channel order", "order_no", orderNo, "err", cerr)
		}
	}
	// The row is the truth for the answer: on the create race the survivor
	// replaces ours, and its stored credential is what the payer continues
	// with — never the loser's fresh one.
	return &CreateOut{Order: created, Credential: credentialOf(created)}, nil
}

func (uc *OrderUC) Close(ctx context.Context, orderNo string) (*Order, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("order close requires a session"))
	}
	order, err := uc.orderRP.ByNo(ctx, orderNo)
	if err != nil {
		return nil, err
	}
	if order == nil || order.UserID != subject.UserID {
		return nil, esportspb.ErrorEsportsOrderNotFound("")
	}
	if order.State != StatePending {
		return nil, esportspb.ErrorEsportsOrderStateConflict("")
	}
	closed, err := uc.retire(ctx, order)
	if err != nil {
		return nil, err
	}
	if closed == nil {
		return nil, esportspb.ErrorEsportsOrderStateConflict("")
	}
	return closed, nil
}

func (uc *OrderUC) List(ctx context.Context, page *commonv1.PageRequest) (*ListOut, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("order list requires a session"))
	}
	return uc.orderRP.List(ctx, &ListIn{PageRequest: page, UserID: subject.UserID})
}

func (uc *OrderUC) Get(ctx context.Context, orderNo string) (*Order, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("order get requires a session"))
	}
	order, err := uc.orderRP.ByNo(ctx, orderNo)
	if err != nil {
		return nil, err
	}
	if order == nil || order.UserID != subject.UserID {
		return nil, esportspb.ErrorEsportsOrderNotFound("")
	}
	return order, nil
}

func (uc *OrderUC) HandleNotification(ctx context.Context, provider string, headers map[string]string, body []byte) (bool, error) {
	h := http.Header{}
	for k, v := range headers {
		h.Set(k, v)
	}
	event, err := uc.channelRP.ParseNotification(ctx, provider, &payment.NotificationInput{Headers: h, Body: body})
	if err != nil {
		return false, err
	}
	switch event.Type {
	case payment.EventPaymentSucceeded:
		return uc.SettlePaid(ctx, &SettleIn{
			OrderNo:       event.OutTradeNo,
			Provider:      provider,
			TransactionID: event.TransactionID,
			AmountFen:     int64(event.AmountFen),
			PaidAt:        event.SuccessAt,
		})
	case payment.EventRefundSucceeded, payment.EventRefundAbnormal:
		// M1 folds nothing on refunds — the event is the audit trail; the
		// order stays paid and entitlements never roll back.
		uc.Log.Info("payment refund event", "channel_type", event.ChannelType, "order_no", event.OutTradeNo, "refund_no", event.OutRefundNo, "refund_id", event.RefundID)
		return true, nil
	default:
		// A channel event this build has never met: no fold, and no ack
		// either — the channel's retry pressure keeps it visible instead of
		// swallowing an unknown money movement.
		uc.Log.Warn("payment event type unrecognized", "channel_type", event.ChannelType, "order_no", event.OutTradeNo)
		return false, nil
	}
}

func (uc *OrderUC) SettlePaid(ctx context.Context, in *SettleIn) (bool, error) {
	if in.PaidAt.IsZero() {
		// A channel that stamps no success time still settles — the local
		// clock is the fallback fact, both settle paths meeting here.
		in.PaidAt = time.Now()
	}
	order, err := uc.orderRP.ByNo(ctx, in.OrderNo)
	if err != nil {
		return false, err
	}
	if order == nil {
		// A paid channel order with no row — the crash window between the
		// channel create and the local insert, paid into after the orphan
		// close missed it. Nothing exists to settle or grant; the alarm and
		// the reconciliation ledger own the money.
		uc.Log.Error("payment settled for unknown order", "order_no", in.OrderNo, "provider", in.Provider, "transaction_id", in.TransactionID)
		return true, nil
	}
	if order.Provider != in.Provider {
		uc.Log.Error("payment provider mismatch", "order_no", in.OrderNo, "row_provider", order.Provider, "notify_provider", in.Provider)
		return true, nil
	}
	if order.AmountFen != in.AmountFen {
		// Money moved at an amount the row never knew. The money stays
		// taken, so the channel may stop retrying — but the grant never
		// follows a mismatched amount.
		uc.Log.Error("payment amount mismatch", "order_no", in.OrderNo, "row_fen", order.AmountFen, "notify_fen", in.AmountFen)
		return true, nil
	}
	var grant *EntitlementGrant
	switch order.ScopeKind {
	case ScopeTrial:
		if order.ScopeRef == nil {
			uc.Log.Error("trial order without record reference", "order_no", order.OrderNo)
			break
		}
		grant = &EntitlementGrant{UserID: order.UserID, ScopeKind: EntTrial, ScopeRef: order.ScopeRef, Source: order.OrderNo}
	case ScopePeriod:
		sku, err := uc.skuRP.ByCode(ctx, order.SKUCode)
		if err != nil {
			return false, err
		}
		if sku == nil || sku.PeriodDays == nil {
			// The sku lost its period between order and settle. The money is
			// taken, so the order still settles — the missing grant is the
			// alarm reconciliation reads.
			uc.Log.Error("membership settle without period", "order_no", order.OrderNo, "sku_code", order.SKUCode)
			break
		}
		expiry := in.PaidAt.AddDate(0, 0, *sku.PeriodDays)
		grant = &EntitlementGrant{UserID: order.UserID, ScopeKind: EntMembership, ExpiresAt: &expiry, Source: order.OrderNo}
	default:
		uc.Log.Error("order scope unknown at settle", "order_no", order.OrderNo, "scope_kind", order.ScopeKind)
	}
	settled := false
	err = uc.Tm.InTx(ctx, func(txCtx context.Context) error {
		_, updated, err := uc.orderRP.Settle(txCtx, in)
		if err != nil {
			return err
		}
		settled = updated
		if !updated || grant == nil {
			return nil
		}
		return uc.entRP.Grant(txCtx, grant)
	})
	if err != nil {
		return false, err
	}
	if settled {
		uc.Log.Info("order settled", "order_no", order.OrderNo, "transaction_id", in.TransactionID)
	}
	return true, nil
}

func (uc *OrderUC) ConvergeOnce(ctx context.Context) error {
	stale, err := uc.orderRP.ListStalePending(ctx, convergeStaleAfter)
	if err != nil {
		return err
	}
	for _, order := range stale {
		if order.ExpiresAt.After(time.Now()) {
			uc.convergeQuery(ctx, order)
			continue
		}
		if _, err := uc.retire(ctx, order); err != nil {
			if esportspb.IsEsportsOrderStateConflict(err) {
				// The channel refused the close — paid in the expiry race.
				// The query is the truth that settles it.
				uc.convergeQuery(ctx, order)
				continue
			}
			uc.Log.Warn("converge retire failed", "order_no", order.OrderNo, "err", err)
		}
	}
	return nil
}

func (uc *OrderUC) ReconcileOnce(ctx context.Context, day time.Time) error {
	err := uc.reconcileDay(ctx, day)
	if errorspb.IsInfraErrorPaymentNotFound(err) {
		// The channel has not published the day yet — a recent day's
		// absence is the channel being the channel.
		uc.Log.Info("payment bill not published", "day", day.Format("2006-01-02"))
		return nil
	}
	return err
}

func (uc *OrderUC) reconcileDay(ctx context.Context, day time.Time) error {
	provider := uc.channelRP.DefaultProvider()
	bill, err := uc.channelRP.DownloadBill(ctx, provider, &payment.BillInput{Date: day, Kind: payment.BillKindTrade})
	if err != nil {
		return err
	}
	lines, err := uc.channelRP.ParseTradeBill(ctx, provider, bill.Raw)
	if err != nil {
		return err
	}
	start, end := billDayWindow(day)
	orders, err := uc.orderRP.ListPaid(ctx, start, end)
	if err != nil {
		return err
	}
	local := make(map[string]*Order, len(orders))
	for _, o := range orders {
		local[o.OrderNo] = o
	}
	// The reconciliation compares day facts: channel rows the bill dates into
	// this window against local rows paid in it — the mock's snapshot bill and
	// WeChat's day bill meet under one rule.
	billed := make(map[string]payment.BillLine, len(lines))
	for _, l := range lines {
		if (l.State == payment.BillLinePaid || l.State == payment.BillLineRefund) && !l.At.Before(start) && l.At.Before(end) {
			billed[l.OutTradeNo] = l
		}
	}
	mismatch := 0
	for _, o := range orders {
		l, ok := billed[o.OrderNo]
		switch {
		case !ok:
			uc.Log.Error("paid order missing from bill", "order_no", o.OrderNo, "day", day.Format("2006-01-02"))
			mismatch++
		case int64(l.AmountFen) != o.AmountFen:
			uc.Log.Error("bill amount mismatch", "order_no", o.OrderNo, "row_fen", o.AmountFen, "bill_fen", int64(l.AmountFen))
			mismatch++
		}
	}
	for no, l := range billed {
		if _, ok := local[no]; ok {
			continue
		}
		row, err := uc.orderRP.ByNo(ctx, no)
		if err != nil {
			return err
		}
		if row == nil {
			uc.Log.Error("bill order unknown locally", "order_no", no, "bill_fen", int64(l.AmountFen))
		} else {
			uc.Log.Error("bill settled an unsettled order", "order_no", no, "local_state", row.State, "bill_fen", int64(l.AmountFen))
		}
		mismatch++
	}
	uc.Log.Info("payment bill reconciled", "day", day.Format("2006-01-02"), "local", len(orders), "billed", len(billed), "mismatch", mismatch)
	return nil
}

// Private -------------------------------------------------------------------------------------------------------------

func (uc *OrderUC) resolveScope(ctx context.Context, userID string, sku *Sku, ref *string) (string, *string, error) {
	// Fixes what the order's grant points at, enforcing the sku-kind
	// contract: a trial names the caller's own un-granted record, a
	// membership takes no reference.
	switch sku.Kind {
	case SkuKindTrial:
		if ref == nil {
			return "", nil, esportspb.ErrorEsportsOrderScopeInvalid("").WithCause(
				fmt.Errorf("a trial order requires the record it unlocks"))
		}
		attempt, err := uc.testRP.GetAttempt(ctx, *ref)
		if err != nil {
			return "", nil, err
		}
		if attempt == nil || attempt.UserID != userID {
			return "", nil, esportspb.ErrorEsportsOrderScopeInvalid("").WithCause(
				fmt.Errorf("record %s is not the caller's attempt", *ref))
		}
		granted, err := uc.entRP.Exists(ctx, userID, EntTrial, ref)
		if err != nil {
			return "", nil, err
		}
		if granted {
			return "", nil, esportspb.ErrorEsportsOrderScopeAlreadyGranted("")
		}
		return ScopeTrial, ref, nil
	case SkuKindMembership:
		if ref != nil {
			return "", nil, esportspb.ErrorEsportsOrderScopeInvalid("").WithCause(
				fmt.Errorf("a membership order takes no record reference"))
		}
		return ScopePeriod, nil, nil
	default:
		return "", nil, esportspb.ErrorEsportsOrderScopeInvalid("").WithCause(
			fmt.Errorf("sku kind %q buys nothing", sku.Kind))
	}
}

func (uc *OrderUC) runLoop(name string, interval time.Duration, sweep func(context.Context) error) func(context.Context) error {
	// One loop shape for every background sweep: a boot sweep first —
	// anything that aged while the process was down folds before the first
	// tick — then the ticker; cancellation rides the lifecycle's context.
	return func(ctx context.Context) error {
		loopCtx, cancel := context.WithCancel(ctx)
		uc.loopCancels = append(uc.loopCancels, cancel)
		go func() {
			uc.Log.Info("payment loop started", "loop", name, "interval", interval)
			if err := sweep(loopCtx); err != nil {
				uc.Log.Error("payment loop boot sweep failed", "loop", name, "err", err)
			}
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			for {
				select {
				case <-loopCtx.Done():
					return
				case <-ticker.C:
				}
				if err := sweep(loopCtx); err != nil {
					uc.Log.Error("payment loop sweep failed", "loop", name, "err", err)
				}
			}
		}()
		return nil
	}
}

func (uc *OrderUC) stopLoops(context.Context) error {
	for _, cancel := range uc.loopCancels {
		cancel()
	}
	return nil
}

func (uc *OrderUC) reconcileSweep(ctx context.Context) error {
	// A day advances only after a sweep handles it: a transient download
	// failure or an unpublished bill is retried on the next tick rather
	// than skipped — a day's bill is the only record of what the channel
	// believes happened. One retry, then the day is abandoned with an
	// ERROR: a trade-less day answers not-found forever, and no single
	// day may wedge the loop.
	yesterday := billDayBefore(time.Now())
	if uc.nextBillDay.IsZero() || uc.nextBillDay.After(yesterday) {
		uc.nextBillDay, uc.nextBillRetry = yesterday, 0
	}
	for !uc.nextBillDay.After(yesterday) {
		day := uc.nextBillDay
		err := uc.reconcileDay(ctx, day)
		if errorspb.IsInfraErrorPaymentNotFound(err) {
			uc.Log.Info("payment bill not published yet", "day", day.Format("2006-01-02"))
		}
		if err != nil {
			uc.nextBillRetry++
			if uc.nextBillRetry < 2 {
				return err
			}
			uc.Log.Error("payment bill day abandoned", "day", day.Format("2006-01-02"), "err", err)
			uc.nextBillRetry = 0
		} else {
			uc.nextBillRetry = 0
		}
		uc.nextBillDay = day.AddDate(0, 0, 1)
	}
	return nil
}

func (uc *OrderUC) convergeQuery(ctx context.Context, order *Order) {
	status, err := uc.channelRP.QueryPayment(ctx, order.Provider, order.OrderNo)
	if err != nil {
		uc.Log.Warn("converge query failed", "order_no", order.OrderNo, "err", err)
		return
	}
	if status.State != payment.PaymentStatePaid {
		return
	}
	paidAt := status.PaidAt
	if paidAt.IsZero() {
		paidAt = time.Now()
	}
	if _, err := uc.SettlePaid(ctx, &SettleIn{
		OrderNo:       order.OrderNo,
		Provider:      order.Provider,
		TransactionID: status.TransactionID,
		AmountFen:     int64(status.AmountFen),
		PaidAt:        paidAt,
	}); err != nil {
		uc.Log.Warn("converge settle failed", "order_no", order.OrderNo, "err", err)
	}
}

func (uc *OrderUC) retire(ctx context.Context, order *Order) (*Order, error) {
	// Takes a pending order out of play: the channel twin closes first
	// (the channel is the state machine; a channel refusal folds to the
	// truth it reports), then the local conditional close frees the reuse
	// slot. A nil return means the row left pending on its own — a settle
	// or a rival close won the race.
	if err := uc.channelRP.ClosePayment(ctx, order.Provider, order.OrderNo); err != nil {
		switch {
		case errorspb.IsInfraErrorPaymentNotFound(err):
			// The channel never knew this trade (a crash between the
			// channel create and the local insert) — the local close is
			// the whole close.
		case errorspb.IsInfraErrorPaymentChannelRejected(err):
			// The channel refuses the close — paid in the expiry race. The
			// converger settles it; a fresh pending order must wait for
			// this row to leave pending first.
			return nil, esportspb.ErrorEsportsOrderStateConflict("").WithCause(err)
		default:
			return nil, err
		}
	}
	return uc.orderRP.ClosePending(ctx, order.OrderNo)
}

func credentialOf(o *Order) payment.Credential {
	return payment.Credential{Kind: payment.CredentialKind(o.CredentialKind), Value: o.CredentialValue}
}

func channelIn(channels []payment.Channel, want payment.Channel) bool {
	for _, c := range channels {
		if c == want {
			return true
		}
	}
	return false
}

func billDayWindow(day time.Time) (time.Time, time.Time) {
	d := day.In(billZone)
	start := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, billZone)
	return start, start.Add(24 * time.Hour)
}

func billDayBefore(t time.Time) time.Time {
	// A day's statement publishes the following day — the sweep always asks
	// for yesterday's bill.
	d := t.In(billZone).AddDate(0, 0, -1)
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, billZone)
}
