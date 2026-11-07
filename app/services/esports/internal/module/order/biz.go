package order

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/rs/xid"

	"cyber-ecosystem/shared-go/capability/payment"
	"cyber-ecosystem/shared-go/helper"
	"cyber-ecosystem/shared-go/kratos/security"
	"cyber-ecosystem/shared-go/utils"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
	commonv1 "cyber-ecosystem/gen/go/cyber/shared/common/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"

	"cyber-ecosystem/app/services/esports/internal/module/compliance"
	"cyber-ecosystem/app/services/esports/internal/module/entitle"
	"cyber-ecosystem/app/services/esports/internal/module/member"
	"cyber-ecosystem/app/services/esports/internal/module/profile"
	"cyber-ecosystem/app/services/esports/internal/module/sku"
	"cyber-ecosystem/app/services/esports/internal/module/test"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

const (
	ScopeTrial  = "trial"  // order: unlocks one test record permanently
	ScopePeriod = "period" // order: opens the membership window

	EntTrial      = shared.ScopeTrial      // the ledger's scope vocabulary lives in the cross-module kernel
	EntMembership = shared.ScopeMembership // membership: a window with an expiry

	StatePending = "pending"
	StatePaid    = "paid"
	StateClosed  = "closed"

	RefundProcessing = "processing"
	RefundSucceeded  = "succeeded"
	RefundAbnormal   = "abnormal"

	RefundNoPrefix = "R" // mints the refund's channel idempotency key from the order's own — one full refund per order makes it unique by construction

	OrderNoPrefix = "P"
	// The channel's own default deadline, carried explicitly so the row and
	// the channel order can never disagree about expiry.
	payWindow = 2 * time.Hour

	convergeInterval   = time.Minute
	convergeStaleAfter = 10 * time.Minute
	reconcileInterval  = 24 * time.Hour
)

var billZone = time.FixedZone("CST", 8*60*60) // bill days are wall-clock facts in the channel's timezone, not the host's — WeChat publishes in Beijing time; the mock's day-agnostic snapshot takes the same window so both providers reconcile under one rule

var rosterHydrate = utils.Ptr(int32(helper.DefaultPageSizeMax)) // one roster page covers the page's distinct ids

// DO ------------------------------------------------------------------------------------------------------------------

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
	// The derived refund fact; nil RefundNo = never refunded. Paid stays
	// terminal — these columns record what the channel did with the money
	// after the fact, not a state the order machine moves through.
	RefundNo     *string
	RefundID     *string
	RefundedFen  *int64
	RefundReason *string
	RefundState  *string
	RefundedAt   *time.Time
}

type CreateIn struct {
	SKUCode   string
	Channel   payment.Channel
	ScopeRef  *string
	ReturnURL *string
	// The payer's wire address as the transport saw it — a server-derived
	// fact the client never declares; the H5 channel requires it for risk
	// control, so the service face extracts it and the channel validates.
	ClientIP string
}

type CreateOut struct {
	Order      *Order
	Credential payment.Credential
}

type ListIn struct {
	*commonv1.PageRequest
	UserID string
	// The operator face's filters — the member face leaves them unset.
	// UserIDs, when the keyword named people, narrows instead of UserID.
	UserIDs []string
	State   string // "" = any
	Scope   string // "" = any
	Prefix  string // order-no prefix
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

type RefundIn struct {
	OrderNo   string
	RefundNo  string
	RefundID  string
	AmountFen int64
	Reason    string
	State     string
	// RefundedAt is the success stamp the refund answer or the refund query
	// produced; nil while the refund is still processing.
	RefundedAt *time.Time
}

type AdminOrderRow struct {
	// The operator's read model of an order: the wire facts plus the
	// hydrated identity and sku-name the list columns need. Nil hydration
	// fields are the lazy shape (no profile row yet, roster gone) — never
	// an error.
	*Order
	Nickname    *string
	PhoneMasked *string
	SkuName     *string
}

type AdminListIn struct {
	*commonv1.PageRequest
	State   *string
	Scope   *string
	Keyword *string
}

type AdminListOut struct {
	*commonv1.PageResponse
	Rows []*AdminOrderRow
}

type UnlockView struct {
	// What a paid order actually unlocked, folded at read time — trial:
	// the report record, permanently.
	RecordID string
	// Membership: the window this order opened (the grant row whose
	// source is this order number) and the unlocks that landed inside it.
	GrantedAt   time.Time
	ExpiresAt   time.Time
	UnlockCount int
}

// Port ----------------------------------------------------------------------------------------------------------------

type OrderRP interface {
	Create(ctx context.Context, o *Order) (*Order, error)                 // a pending-unique conflict returns the surviving active order
	ClosePending(ctx context.Context, orderNo string) (*Order, error)     // nil, nil = no longer pending
	Settle(ctx context.Context, in *SettleIn) (*Order, bool, error)       // the single write-paid entry; false = a rival settle already won
	RefundPaid(ctx context.Context, in *RefundIn) (*Order, bool, error)   // the single write-refund entry; false = a rival refund already won
	SettleRefund(ctx context.Context, in *RefundIn) (*Order, bool, error) // processing → terminal refund convergence; false = already terminal
	List(ctx context.Context, in *ListIn) (*ListOut, error)
	ListStalePending(ctx context.Context, olderThan time.Duration) ([]*Order, error)               // pending rows past the horizon; expiry is the converger's branch, not the query's
	ListPaid(ctx context.Context, from, to time.Time) ([]*Order, error)                            // paid rows by paid_at window — the reconciliation's local side
	ByNo(ctx context.Context, orderNo string) (*Order, error)                                      // nil, nil = no such order
	ActiveByScope(ctx context.Context, userID, scopeKind string, scopeRef *string) (*Order, error) // nil, nil = none live
}

type PayChannelRP interface { // routes to the configured provider's backend; capability errors arrive classified
	CreatePayment(ctx context.Context, provider string, in *payment.CreatePaymentInput) (*payment.PrepayResult, error)
	ClosePayment(ctx context.Context, provider, outTradeNo string) error
	QueryPayment(ctx context.Context, provider, outTradeNo string) (*payment.PaymentStatusResult, error)
	Refund(ctx context.Context, provider string, in *payment.RefundInput) (*payment.RefundResult, error)
	QueryRefund(ctx context.Context, provider, outRefundNo string) (*payment.RefundStatusResult, error)
	ParseNotification(ctx context.Context, provider string, in *payment.NotificationInput) (*payment.NotificationEvent, error)
	DownloadBill(ctx context.Context, provider string, in *payment.BillInput) (*payment.BillResult, error) // the day's raw statement; NotFound = the channel has not published it yet
	ParseTradeBill(ctx context.Context, provider string, raw []byte) ([]payment.BillLine, error)           // neutralized statement rows in the contract's vocabulary
	Channels(provider string) ([]payment.Channel, error)
	DefaultProvider() string
	NotifyBaseURL() string
}

type SystemRP interface {
	// The caller's own wechat openid via the forwarded session — nil = the
	// payer never bound wechat, which the invoke channel refuses at create.
	WechatOpenID(ctx context.Context) (*string, error)
}

// UC ------------------------------------------------------------------------------------------------------------------

type OrderUC struct {
	shared.UC
	orderRP       OrderRP
	skuRP         sku.SkuRP
	entRP         entitle.EntitlementRP
	testRP        test.TestRP
	channelRP     PayChannelRP
	profileRP     profile.ProfileRP
	complianceRP  compliance.ComplianceRP
	systemRP      SystemRP
	eventBus      shared.DomainEventBus
	loopCancels   []context.CancelFunc
	nextBillDay   time.Time // the day the next reconcile sweep must still see; owned by the reconcile loop's goroutine
	nextBillRetry int
}

func NewOrderUC(logger *slog.Logger, tm shared.Transaction, hooks shared.HookRegistry, eventBus shared.DomainEventBus, orderRP OrderRP, skuRP sku.SkuRP, entRP entitle.EntitlementRP, testRP test.TestRP, channelRP PayChannelRP, profileRP profile.ProfileRP, complianceRP compliance.ComplianceRP, systemRP SystemRP) *OrderUC {
	uc := &OrderUC{
		UC:           shared.NewUC(logger.With("module", "module/order"), tm),
		orderRP:      orderRP,
		skuRP:        skuRP,
		entRP:        entRP,
		testRP:       testRP,
		channelRP:    channelRP,
		profileRP:    profileRP,
		complianceRP: complianceRP,
		systemRP:     systemRP,
		eventBus:     eventBus,
	}
	hooks.OnStart(uc.runLoop("converge", convergeInterval, uc.ConvergeOnce))
	hooks.OnStart(uc.runLoop("reconcile", reconcileInterval, uc.reconcileSweep))
	hooks.OnStop(uc.stopLoops)
	return uc
}

type OrderAdminUC struct {
	// The operator's UC over the same aggregate — a sibling of the member
	// UC: the admin face's whole-order list, the detail with its folded
	// unlock outcome, and the operator's close. Ports only; the owner
	// check the member Close carries is deliberately absent — the admin
	// face's access ladder already settled who may call here.
	shared.UC
	orderRP   OrderRP
	channelRP PayChannelRP
	skuRP     sku.SkuRP
	profileRP profile.ProfileRP
	systemRP  member.SystemRP
	entRP     entitle.EntitlementRP
}

func NewOrderAdminUC(logger *slog.Logger, tm shared.Transaction, orderRP OrderRP, channelRP PayChannelRP, skuRP sku.SkuRP, profileRP profile.ProfileRP, systemRP member.SystemRP, entRP entitle.EntitlementRP) *OrderAdminUC {
	return &OrderAdminUC{
		UC:        shared.NewUC(logger.With("module", "module/order_admin"), tm),
		orderRP:   orderRP,
		channelRP: channelRP,
		skuRP:     skuRP,
		profileRP: profileRP,
		systemRP:  systemRP,
		entRP:     entRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *OrderUC) Create(ctx context.Context, in *CreateIn) (*CreateOut, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("order create requires a session"))
	}
	// The pay gate's authority — the wall's catalog-side pre-check is
	// experience, this refusal is the rule. A member whose birth cannot
	// be read is refused the same way as a reading under the line: the
	// two reasons route differently on the client (backfill vs final).
	gate, err := uc.complianceRP.Get(ctx)
	if err != nil {
		return nil, err
	}
	p, err := uc.profileRP.FindByUserID(ctx, subject.UserID)
	if err != nil {
		return nil, err
	}
	if p == nil || p.BirthYear == nil || p.BirthMonth == nil {
		return nil, esportspb.ErrorEsportsPayBirthMissing("")
	}
	if !lineReached(int(*p.BirthYear), int(*p.BirthMonth), int(gate.PayAgeLine), time.Now()) {
		return nil, esportspb.ErrorEsportsPayAgeBlocked("")
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
		if active.ExpiresAt.After(time.Now()) && active.Channel == string(in.Channel) {
			// Same-order reuse: the stored credential replays with zero
			// channel round trips, so the channel never sees a second
			// create for the same grant target. A channel switch is not
			// a reuse — the old channel's credential is dead weight in
			// the new channel's hands — so the old order retires and a
			// fresh one takes the member's chosen channel instead of
			// handing back a credential it cannot pay with.
			return &CreateOut{Order: active, Credential: credentialOf(active)}, nil
		}
		if _, err := retirePending(ctx, uc.channelRP, uc.orderRP, active); err != nil {
			return nil, err
		}
	}
	orderNo := OrderNoPrefix + xid.New().String()
	expireAt := time.Now().Add(payWindow)
	var payer *payment.Payer
	if strings.HasSuffix(string(in.Channel), "_jsapi") {
		// The invoke channel's payer openid is the caller's wechat binding,
		// never a request field — the client has no way to know one, so the
		// order server resolves it from identity at the only hop that needs
		// it. An account that never bound wechat has nothing to pay with.
		openid, err := uc.systemRP.WechatOpenID(ctx)
		if err != nil {
			return nil, err
		}
		if openid == nil {
			return nil, esportspb.ErrorEsportsPayWechatUnbound("")
		}
		payer = &payment.Payer{OpenID: *openid}
	}
	if in.ReturnURL != nil || in.ClientIP != "" {
		// The caller's landing URL cannot name the order — the number is
		// minted right here — so the caller sends the receipt URL with its
		// own params and this side appends the one fact only it owns; the
		// channel's post-payment redirect then lands on a receipt that
		// reads its own row.
		returnURL := utils.Deref(in.ReturnURL, "")
		if returnURL != "" {
			sep := "?"
			if strings.Contains(returnURL, "?") {
				sep = "&"
			}
			returnURL += sep + "order=" + orderNo
		}
		if payer == nil {
			payer = &payment.Payer{}
		}
		payer.ClientIP = in.ClientIP
		payer.ReturnURL = returnURL
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
	closed, err := retirePending(ctx, uc.channelRP, uc.orderRP, order)
	if err != nil {
		return nil, err
	}
	if closed == nil {
		return nil, esportspb.ErrorEsportsOrderStateConflict("")
	}
	return closed, nil
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
	var grant *entitle.EntitlementGrant
	var periodDays int
	switch order.ScopeKind {
	case ScopeTrial:
		if order.ScopeRef == nil {
			uc.Log.Error("trial order without record reference", "order_no", order.OrderNo)
			break
		}
		grant = &entitle.EntitlementGrant{UserID: order.UserID, ScopeKind: EntTrial, ScopeRef: order.ScopeRef, Source: order.OrderNo}
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
		periodDays = *sku.PeriodDays
		// The expiry is the settle transaction's to set — the renewal seam
		// there tiles a live window's renewal past its standing remainder.
		grant = &entitle.EntitlementGrant{UserID: order.UserID, ScopeKind: EntMembership, Source: order.OrderNo}
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
		if grant.ScopeKind == EntMembership {
			// The membership writers' lock, the operator grant and
			// terminate paths' twin: the seam read below cannot straddle a
			// concurrent terminate's commit — a renewal paid moments
			// before a termination must not outlive it.
			if lerr := uc.entRP.LockScope(txCtx, order.UserID, EntMembership, nil); lerr != nil {
				return lerr
			}
			// The renewal seam, the same in-transaction read the operator
			// grant makes: a live window sets the start, so a renewal paid
			// mid-window tiles after the standing remainder rather than
			// overlapping and forfeiting it; a first purchase starts at the
			// paid moment. The writers' lock serializes concurrent renewal
			// settles too, so the tiles never overlap.
			start := in.PaidAt
			if window, werr := uc.entRP.MembershipWindow(txCtx, order.UserID); werr != nil {
				return werr
			} else if window != nil {
				start = *window.ExpiresAt
			}
			expiry := start.AddDate(0, 0, periodDays)
			grant.ExpiresAt = &expiry
		}
		return uc.entRP.Grant(txCtx, grant)
	})
	if err != nil {
		return false, err
	}
	if settled {
		uc.Log.Info("order settled", "order_no", order.OrderNo, "transaction_id", in.TransactionID)
	}
	// The trial grant named its record and the transaction committed, so the
	// freeze can pin the judgement at unlock time — best-effort, the report's
	// lazy read path backstops a missed one. Membership orders unlock no
	// record and publish nothing: their freezes happen on first read.
	if settled && order.ScopeKind == ScopeTrial && order.ScopeRef != nil && uc.eventBus != nil {
		uc.eventBus.PublishReportUnlocked(ctx, shared.ReportUnlocked{UserID: order.UserID, RecordID: *order.ScopeRef})
	}
	return true, nil
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
		if _, err := retirePending(ctx, uc.channelRP, uc.orderRP, order); err != nil {
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

func (uc *OrderAdminUC) Close(ctx context.Context, orderNo string) (*AdminOrderRow, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("order close requires an admin session"))
	}
	o, err := uc.orderRP.ByNo(ctx, orderNo)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, esportspb.ErrorEsportsOrderNotFound("")
	}
	if o.State != StatePending {
		return nil, esportspb.ErrorEsportsOrderStateConflict("")
	}
	closed, err := retirePending(ctx, uc.channelRP, uc.orderRP, o)
	if err != nil {
		return nil, err
	}
	if closed == nil {
		return nil, esportspb.ErrorEsportsOrderStateConflict("")
	}
	uc.Log.Info("order closed by operator", "order_no", orderNo, "user_id", o.UserID, "by", subject.UserID)
	rows, err := uc.hydrate(ctx, []*Order{closed})
	if err != nil {
		return nil, err
	}
	return rows[0], nil
}

func (uc *OrderAdminUC) Refund(ctx context.Context, orderNo, reason string) (*AdminOrderRow, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("order refund requires an admin session"))
	}
	o, err := uc.orderRP.ByNo(ctx, orderNo)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, esportspb.ErrorEsportsOrderNotFound("")
	}
	if o.State != StatePaid {
		// Only a paid order carries a refundable fact — a pending one has
		// taken no money (close it instead), a closed one never will.
		return nil, esportspb.ErrorEsportsOrderStateConflict("")
	}
	if o.RefundNo != nil {
		return nil, esportspb.ErrorEsportsOrderAlreadyRefunded("")
	}
	// The full-refund tool: both amounts are the order's own — the channel
	// requires the original total beside the refund figure, and a partial
	// refund is not a product shape. No NotifyURL: the refund converges
	// through the refund query, the same authority shape the settle loop
	// already runs on.
	refundNo := RefundNoPrefix + o.OrderNo
	res, err := uc.channelRP.Refund(ctx, o.Provider, &payment.RefundInput{
		OutTradeNo:     o.OrderNo,
		OutRefundNo:    refundNo,
		AmountFen:      payment.Fen(o.AmountFen),
		TotalAmountFen: payment.Fen(o.AmountFen),
		Reason:         reason,
	})
	if err != nil {
		return nil, err
	}
	in := &RefundIn{
		OrderNo:   o.OrderNo,
		RefundNo:  refundNo,
		RefundID:  res.RefundID,
		AmountFen: o.AmountFen,
		Reason:    reason,
		State:     string(res.State),
	}
	if res.State == payment.RefundStateSucceeded {
		// The sync answer carries no success time; the observation moment is
		// the fallback fact — the same rule the settle path applies to a
		// channel that stamps none.
		now := time.Now()
		in.RefundedAt = &now
	}
	refunded, won, err := uc.orderRP.RefundPaid(ctx, in)
	if err != nil {
		return nil, err
	}
	if won {
		uc.Log.Info("order refunded by operator", "order_no", orderNo, "user_id", o.UserID, "refund_no", refundNo, "state", in.State, "by", subject.UserID)
	}
	// A lost arbitration is the idempotent answer, not an error: the rival
	// call minted the same channel refund (the out_refund_no is derived
	// from the order), so the row it left is this caller's truth too.
	rows, err := uc.hydrate(ctx, []*Order{refunded})
	if err != nil {
		return nil, err
	}
	return rows[0], nil
}

func (uc *OrderAdminUC) List(ctx context.Context, in *AdminListIn) (*AdminListOut, error) {
	listIn := &ListIn{PageRequest: in.PageRequest}
	if in.State != nil {
		listIn.State = *in.State
	}
	if in.Scope != nil {
		listIn.Scope = *in.Scope
	}
	if in.Keyword != nil {
		// One box, two readings: order-no shaped input queries orders
		// directly, anything else resolves people first (nickname via the
		// profile plane; phone search is a system-plane gap the ledger
		// tracks). A P-prefixed nickname misroutes — one empty answer,
		// the operator retypes.
		kw := strings.TrimPrefix(strings.TrimSpace(*in.Keyword), "#")
		if strings.HasPrefix(kw, OrderNoPrefix) {
			listIn.Prefix = kw
		} else {
			userIDs, err := uc.profileRP.ListUserIDs(ctx, &kw, nil)
			if err != nil {
				return nil, err
			}
			if len(userIDs) == 0 {
				return &AdminListOut{PageResponse: helper.BuildPageResponse(0, 0, 0)}, nil
			}
			listIn.UserIDs = userIDs
		}
	}
	out, err := uc.orderRP.List(ctx, listIn)
	if err != nil {
		return nil, err
	}
	rows, err := uc.hydrate(ctx, out.Orders)
	if err != nil {
		return nil, err
	}
	return &AdminListOut{PageResponse: out.PageResponse, Rows: rows}, nil
}

func (uc *OrderAdminUC) Get(ctx context.Context, orderNo string) (*AdminOrderRow, *UnlockView, error) {
	o, err := uc.orderRP.ByNo(ctx, orderNo)
	if err != nil {
		return nil, nil, err
	}
	if o == nil {
		return nil, nil, esportspb.ErrorEsportsOrderNotFound("")
	}
	o = uc.convergeRefund(ctx, o)
	rows, err := uc.hydrate(ctx, []*Order{o})
	if err != nil {
		return nil, nil, err
	}
	if o.State != StatePaid {
		// A pending order has unlocked nothing; a closed one never will.
		return rows[0], nil, nil
	}
	view, err := uc.foldUnlock(ctx, o)
	if err != nil {
		return nil, nil, err
	}
	return rows[0], view, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func (uc *OrderUC) resolveScope(ctx context.Context, userID string, s *sku.Sku, ref *string) (string, *string, error) {
	// Fixes what the order's grant points at, enforcing the sku-kind
	// contract: a trial names the caller's own un-granted record, a
	// membership takes no reference.
	switch s.Kind {
	case sku.KindTrial:
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
	case sku.KindMembership:
		if ref != nil {
			return "", nil, esportspb.ErrorEsportsOrderScopeInvalid("").WithCause(
				fmt.Errorf("a membership order takes no record reference"))
		}
		return ScopePeriod, nil, nil
	default:
		return "", nil, esportspb.ErrorEsportsOrderScopeInvalid("").WithCause(
			fmt.Errorf("sku kind %q buys nothing", s.Kind))
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

func (uc *OrderUC) reconcileDay(ctx context.Context, day time.Time) error {
	// One bill per provider, not one bill for the world: a day's paid rows
	// carry the provider they were paid on, and each provider's rows meet
	// that provider's own statement — the mock's snapshot bill and WeChat's
	// day bill meet under one rule, never under each other's.
	start, end := billDayWindow(day)
	orders, err := uc.orderRP.ListPaid(ctx, start, end)
	if err != nil {
		return err
	}
	byProvider := make(map[string][]*Order, 1)
	for _, o := range orders {
		byProvider[o.Provider] = append(byProvider[o.Provider], o)
	}
	providers := make([]string, 0, len(byProvider))
	for p := range byProvider {
		providers = append(providers, p)
	}
	slices.Sort(providers)
	for _, provider := range providers {
		rows := byProvider[provider]
		if err := uc.reconcileProviderDay(ctx, day, start, end, provider, rows); err != nil {
			return err
		}
	}
	return nil
}

func (uc *OrderUC) reconcileProviderDay(ctx context.Context, day, start, end time.Time, provider string, orders []*Order) error {
	bill, err := uc.channelRP.DownloadBill(ctx, provider, &payment.BillInput{Date: day, Kind: payment.BillKindTrade})
	if err != nil {
		if errorspb.IsInfraErrorPaymentUnavailable(err) {
			// Rows paid on a provider this deployment does not serve sit
			// outside this reconciler's authority — the deployment that
			// configures the provider owns their reconciliation.
			uc.Log.Warn("payment bill provider not served; rows skipped", "provider", provider, "rows", len(orders), "day", day.Format("2006-01-02"))
			return nil
		}
		return err
	}
	lines, err := uc.channelRP.ParseTradeBill(ctx, provider, bill.Raw)
	if err != nil {
		return err
	}
	local := make(map[string]*Order, len(orders))
	for _, o := range orders {
		local[o.OrderNo] = o
	}
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
			uc.Log.Error("paid order missing from bill", "order_no", o.OrderNo, "provider", o.Provider, "day", day.Format("2006-01-02"))
			mismatch++
		case int64(l.AmountFen) != o.AmountFen:
			uc.Log.Error("bill amount mismatch", "order_no", o.OrderNo, "row_fen", o.AmountFen, "bill_fen", int64(l.AmountFen))
			mismatch++
		}
		if o.ScopeKind != ScopeTrial && o.ScopeKind != ScopePeriod {
			continue
		}
		// The settle path's own alarm, read where alarms are read: money
		// moved and the row settled, so the grant the order owes must
		// exist — a sku losing its period between order and settle skips
		// the grant while the amount reconciliation above stays green
		// (both sides agree the money moved), leaving this probe the only
		// witness.
		granted, gerr := uc.entRP.ExistsSource(ctx, o.UserID, o.OrderNo)
		if gerr != nil {
			return gerr
		}
		if !granted {
			uc.Log.Error("settled order missing its grant", "order_no", o.OrderNo, "scope_kind", o.ScopeKind, "day", day.Format("2006-01-02"))
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
	uc.Log.Info("payment bill reconciled", "provider", provider, "day", day.Format("2006-01-02"), "local", len(orders), "billed", len(billed), "mismatch", mismatch)
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

func (uc *OrderAdminUC) hydrate(ctx context.Context, orders []*Order) ([]*AdminOrderRow, error) {
	rows := make([]*AdminOrderRow, len(orders))
	ids := make([]string, 0, len(orders))
	// One member may own several orders on the page; the roster's id list
	// is unique-items on the wire, so dedupe before the call.
	seen := make(map[string]struct{}, len(orders))
	for i, o := range orders {
		rows[i] = &AdminOrderRow{Order: o}
		if _, dup := seen[o.UserID]; !dup {
			seen[o.UserID] = struct{}{}
			ids = append(ids, o.UserID)
		}
	}
	skus, err := uc.skuRP.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	names := make(map[string]string, len(skus))
	for _, s := range skus {
		names[s.Code] = s.Name
	}
	if len(ids) > 0 {
		roster, err := uc.systemRP.ListRoster(ctx, &member.RosterQuery{
			Population: member.DefaultPopulation,
			UserIDs:    ids,
			Page:       &commonv1.PageRequest{PageSize: rosterHydrate},
		})
		if err != nil {
			return nil, err
		}
		profiles, err := uc.profileRP.ListByUserIDs(ctx, ids)
		if err != nil {
			return nil, err
		}
		byID := make(map[string]*profile.Profile, len(profiles))
		for _, p := range profiles {
			byID[p.UserID] = p
		}
		for _, u := range roster.Users {
			for _, r := range rows {
				if r.UserID == u.ID {
					r.PhoneMasked = u.PhoneMasked
				}
			}
		}
		for _, r := range rows {
			if p := byID[r.UserID]; p != nil {
				r.Nickname = p.Nickname
			}
			if n, ok := names[r.SKUCode]; ok {
				r.SkuName = &n
			}
		}
	}
	return rows, nil
}

func (uc *OrderAdminUC) foldUnlock(ctx context.Context, o *Order) (*UnlockView, error) {
	if o.ScopeKind == ScopeTrial {
		// The bought record is the order's own scope ref — the settle
		// transaction granted it, so the id needs no second read.
		view := &UnlockView{RecordID: utils.Deref(o.ScopeRef, "")}
		return view, nil
	}
	rows, err := uc.entRP.ListMembershipRows(ctx, []string{o.UserID})
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row.Source != o.OrderNo {
			continue
		}
		view := &UnlockView{GrantedAt: row.CreatedAt, ExpiresAt: utils.Deref(row.ExpiresAt, time.Time{})}
		unlocks, err := uc.entRP.ListMembershipUnlocks(ctx, []string{o.UserID})
		if err != nil {
			return nil, err
		}
		for _, u := range unlocks {
			if !u.GrantedAt.Before(row.CreatedAt) && u.GrantedAt.Before(view.ExpiresAt) {
				view.UnlockCount++
			}
		}
		return view, nil
	}
	// A paid membership order without its grant row — a state no write
	// path can produce; the honest answer is the empty view.
	return &UnlockView{}, nil
}

func (uc *OrderAdminUC) convergeRefund(ctx context.Context, o *Order) *Order {
	// Folds a still-processing refund onto its channel truth; the query is
	// the authority (no refund notification is subscribed). A refused or
	// unchanged answer serves the row as-is — a stale processing tag is
	// honest, a blocked read is not.
	if o.RefundNo == nil || o.RefundState == nil || *o.RefundState != RefundProcessing {
		return o
	}
	status, err := uc.channelRP.QueryRefund(ctx, o.Provider, *o.RefundNo)
	if err != nil {
		uc.Log.Warn("refund converge query failed", "order_no", o.OrderNo, "refund_no", *o.RefundNo, "err", err)
		return o
	}
	if status.State == payment.RefundStateProcessing {
		return o
	}
	in := &RefundIn{
		OrderNo:  o.OrderNo,
		RefundNo: *o.RefundNo,
		RefundID: status.RefundID,
		State:    string(status.State),
	}
	if status.State == payment.RefundStateSucceeded {
		at := status.SuccessAt
		if at.IsZero() {
			at = time.Now()
		}
		in.RefundedAt = &at
	}
	refunded, _, err := uc.orderRP.SettleRefund(ctx, in)
	if err != nil {
		uc.Log.Warn("refund converge settle failed", "order_no", o.OrderNo, "refund_no", *o.RefundNo, "err", err)
		return o
	}
	if refunded != nil {
		return refunded
	}
	return o
}

func retirePending(ctx context.Context, channelRP PayChannelRP, orderRP OrderRP, order *Order) (*Order, error) {
	// Takes a pending order out of play: the channel twin closes first
	// (the channel is the state machine; a channel refusal folds to the
	// truth it reports), then the local conditional close frees the reuse
	// slot. A nil return means the row left pending on its own — a settle
	// or a rival close won the race.
	if err := channelRP.ClosePayment(ctx, order.Provider, order.OrderNo); err != nil {
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
	return orderRP.ClosePending(ctx, order.OrderNo)
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

func lineReached(birthYear, birthMonth, line int, now time.Time) bool {
	// Birth is year+month granularity, so the line clears conservatively
	// only from the month after the birthday month (次月起算): a June
	// birth clears an 18-year line in July, never inside June itself.
	years := now.Year() - birthYear
	if years != line {
		return years > line
	}
	return int(now.Month()) > birthMonth
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
