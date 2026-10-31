package e2e

// The full-chain live probe: the real database, the mock channel over its
// real controller, and the service's own faces end to end — create, channel
// pay, callback (twice, because at-least-once is the contract), settle,
// entitlement, converger fold, daily reconciliation, and the refusal arms.
// Gated: PAY_E2E=1 with the dev database up runs it; anything else skips.

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"cyber-ecosystem/shared-go/capability/payment"
	"cyber-ecosystem/shared-go/capability/payment/mock"
	"cyber-ecosystem/shared-go/kratos/security"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"

	"cyber-ecosystem/app/services/esports/internal/conf"
	"cyber-ecosystem/app/services/esports/internal/ent"
	"cyber-ecosystem/app/services/esports/internal/ent/entitlement"
	"cyber-ecosystem/app/services/esports/internal/module/entitle"
	"cyber-ecosystem/app/services/esports/internal/module/order"
	"cyber-ecosystem/app/services/esports/internal/module/sku"
	"cyber-ecosystem/app/services/esports/internal/module/test"
	"cyber-ecosystem/app/services/esports/internal/platform"
)

type noopHooks struct{}

func (noopHooks) OnStart(func(context.Context) error) {}
func (noopHooks) OnStop(func(context.Context) error)  {}

func strp(v string) *string { return &v }

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func TestPayLiveE2E(t *testing.T) {
	if os.Getenv("PAY_E2E") != "1" {
		t.Skip("set PAY_E2E=1 with the dev database up to run the payment end-to-end")
	}
	const (
		uid  = "e2epayuser0000000001" // MaxRuneLen 20
		uid2 = "e2epayuser0000000002"
		recA = "e2erecord00000000001"
		recB = "e2erecord00000000002"
	)
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	db, closeDB, err := platform.NewEntClient(&conf.Data{Database: &conf.Data_Database{
		Driver: "postgres", Host: envOr("DB_HOST", "localhost"), Port: 5432,
		User: "postgres", Password: envOr("DB_PASS", "postgres"), DbName: envOr("DB_NAME", "esports"),
		MaxOpenConns: 4,
	}}, logger)
	if err != nil {
		t.Fatalf("ent client: %v", err)
	}
	payments, closePayments, err := platform.NewPayments(&conf.Payment{
		DefaultProvider: "mock",
		NotifyBaseUrl:   "http://localhost:5174",
		Providers:       []*conf.Payment_Provider{{Name: "mock", Kind: "mock"}},
	})
	if err != nil {
		t.Fatalf("payments: %v", err)
	}
	entHandler := platform.NewEntErrorHandler()
	payHandler, err := platform.NewPaymentErrorHandler()
	if err != nil {
		t.Fatalf("pay handler: %v", err)
	}
	pf, err := platform.NewPlatform(db, entHandler, payments, payHandler)
	if err != nil {
		t.Fatalf("platform: %v", err)
	}
	t.Cleanup(func() {
		cleanE2E(t, db, uid)
		closePayments()
		closeDB()
	})
	cleanE2E(t, db, uid)

	backend, err := pf.PaymentFor("mock")
	if err != nil {
		t.Fatalf("backend: %v", err)
	}
	chanRP := order.NewPayChannelRP(logger, pf)
	ctrl, ok := mock.ControllerOf(backend)
	if !ok {
		t.Fatalf("mock controller: not a mock backend")
	}
	uc := order.NewOrderUC(logger, pf, noopHooks{}, nil, order.NewOrderRP(logger, pf), sku.NewSkuRP(logger, pf),
		entitle.NewEntitlementRP(logger, pf), test.NewTestRP(logger, pf), chanRP)
	svc := order.NewOrderService(logger, uc)

	channels, err := chanRP.Channels("mock")
	if err != nil {
		t.Fatalf("channels: %v", err)
	}
	var channel payment.Channel
	for _, c := range channels {
		if strings.HasSuffix(string(c), "_native") { // the QR face: no payer hints, the e2e has no login-side openid
			channel = c
		}
	}
	if channel == "" {
		t.Fatalf("no native channel in %v", channels)
	}
	subject := &security.Subject{UserID: uid, TenantID: "default", Population: "user"}
	ctx := security.WithSubject(context.Background(), subject)
	notify := func(ni *payment.NotificationInput) (*esportspb.NotifyPaymentResponse, error) {
		hdrs := make(map[string]string, len(ni.Headers))
		for k, vs := range ni.Headers {
			if len(vs) > 0 {
				hdrs[k] = vs[0]
			}
		}
		return svc.NotifyPayment(context.Background(), &esportspb.NotifyPaymentRequest{Provider: strp("mock"), Headers: hdrs, Body: ni.Body})
	}
	entCount := func(kind string) (int, *ent.Entitlement) {
		rows, err := pf.GetClient(context.Background()).Entitlement.Query().
			Where(
				entitlement.UserIDEQ(uid),
				entitlement.ScopeKindEQ(entitlement.ScopeKind(kind)),
			).All(context.Background())
		if err != nil {
			t.Fatalf("entitlement query: %v", err)
		}
		var first *ent.Entitlement
		if len(rows) > 0 {
			first = rows[0]
		}
		return len(rows), first
	}

	now := time.Now()
	if _, err := pf.GetClient(context.Background()).TestRecord.Create().
		SetID(recA).SetTestCode("e2e").SetUserID(uid).SetAttemptNo(1).
		SetStartedAt(now).SetEndedAt(now).SetRawDetail(map[string]any{"e2e": true}).
		Save(context.Background()); err != nil {
		t.Fatalf("attempt fixture A: %v", err)
	}
	if _, err := pf.GetClient(context.Background()).TestRecord.Create().
		SetID(recB).SetTestCode("e2e").SetUserID(uid2).SetAttemptNo(1).
		SetStartedAt(now).SetEndedAt(now).SetRawDetail(map[string]any{"e2e": true}).
		Save(context.Background()); err != nil {
		t.Fatalf("attempt fixture B: %v", err)
	}

	// The membership main line: create → channel pay → callback → settle →
	// grant, with the callback delivered twice because at-least-once is the
	// channel contract and idempotence is the settle's to carry.
	mem1, err := uc.Create(ctx, &order.CreateIn{SKUCode: "member_month", Channel: channel})
	if err != nil {
		t.Fatalf("membership create: %v", err)
	}
	if mem1.Order.State != order.StatePending || mem1.Order.AmountFen != 2990 || mem1.Credential.Value == "" {
		t.Fatalf("membership create shape: %+v", mem1.Order)
	}
	memN1, err := ctrl.MarkPaid(mem1.Order.OrderNo)
	if err != nil {
		t.Fatalf("channel pay: %v", err)
	}
	if resp, err := notify(memN1); err != nil || !resp.GetHandled() {
		t.Fatalf("first callback: resp=%v err=%v", resp, err)
	}
	if resp, err := notify(memN1); err != nil || !resp.GetHandled() {
		t.Fatalf("duplicate callback: resp=%v err=%v", resp, err)
	}
	got, err := uc.Get(ctx, mem1.Order.OrderNo)
	if err != nil {
		t.Fatalf("get settled: %v", err)
	}
	if got.State != order.StatePaid || got.TransactionID == nil || got.PaidAt == nil {
		t.Fatalf("settled shape: %+v", got)
	}
	rows, first := entCount(order.EntMembership)
	if rows != 1 {
		t.Fatalf("double delivery granted %d membership rows, want 1", rows)
	}
	if first.TenantID != "default" || first.ScopeRef != nil || first.Source != mem1.Order.OrderNo {
		t.Fatalf("membership grant shape: %+v", first)
	}
	if d := first.ExpiresAt.Sub(*got.PaidAt); d < 31*24*time.Hour-time.Minute || d > 31*24*time.Hour+time.Minute {
		t.Fatalf("membership window: %v off 31 days", d)
	}

	// Renewal stacks: a second paid membership is a second row, never a dedup.
	mem2, err := uc.Create(ctx, &order.CreateIn{SKUCode: "member_month", Channel: channel})
	if err != nil {
		t.Fatalf("renewal create: %v", err)
	}
	memN2, err := ctrl.MarkPaid(mem2.Order.OrderNo)
	if err != nil {
		t.Fatalf("renewal channel pay: %v", err)
	}
	if resp, err := notify(memN2); err != nil || !resp.GetHandled() {
		t.Fatalf("renewal callback: %v", err)
	}
	if rows, _ := entCount(order.EntMembership); rows != 2 {
		t.Fatalf("renewal rows: %d, want 2 (stacking)", rows)
	}

	// The trial line: same-order reuse before payment replays the stored
	// credential with zero channel round trips; after settle the record is
	// granted permanently and a second trial order for it is refused.
	refA := recA
	trial1, err := uc.Create(ctx, &order.CreateIn{SKUCode: "trial_report", Channel: channel, ScopeRef: &refA})
	if err != nil {
		t.Fatalf("trial create: %v", err)
	}
	if trial1.Order.AmountFen != 1990 || trial1.Order.ScopeKind != order.ScopeTrial || *trial1.Order.ScopeRef != recA {
		t.Fatalf("trial shape: %+v", trial1.Order)
	}
	reuse, err := uc.Create(ctx, &order.CreateIn{SKUCode: "trial_report", Channel: channel, ScopeRef: &refA})
	if err != nil {
		t.Fatalf("trial reuse: %v", err)
	}
	if reuse.Order.OrderNo != trial1.Order.OrderNo {
		t.Fatalf("trial reuse minted a second order: %s vs %s", reuse.Order.OrderNo, trial1.Order.OrderNo)
	}
	trialN, err := ctrl.MarkPaid(trial1.Order.OrderNo)
	if err != nil {
		t.Fatalf("trial channel pay: %v", err)
	}
	if resp, err := notify(trialN); err != nil || !resp.GetHandled() {
		t.Fatalf("trial callback: %v", err)
	}
	rows, tfirst := entCount(order.EntTrial)
	if rows != 1 || *tfirst.ScopeRef != recA || tfirst.ExpiresAt != nil {
		t.Fatalf("trial grant shape: rows=%d row=%+v", rows, tfirst)
	}
	if _, err := uc.Create(ctx, &order.CreateIn{SKUCode: "trial_report", Channel: channel, ScopeRef: &refA}); !esportspb.IsEsportsOrderScopeAlreadyGranted(err) {
		t.Fatalf("re-granted trial: err=%v", err)
	}
	refB := recB
	if _, err := uc.Create(ctx, &order.CreateIn{SKUCode: "trial_report", Channel: channel, ScopeRef: &refB}); !esportspb.IsEsportsOrderScopeInvalid(err) {
		t.Fatalf("foreign record accepted: err=%v", err)
	}
	if _, err := uc.Create(ctx, &order.CreateIn{SKUCode: "no-such-sku", Channel: channel}); !esportspb.IsEsportsSkuNotFound(err) {
		t.Fatalf("unknown sku: err=%v", err)
	}
	if _, err := uc.Close(ctx, mem1.Order.OrderNo); !esportspb.IsEsportsOrderStateConflict(err) {
		t.Fatalf("close paid: err=%v", err)
	}
	forge := &payment.NotificationInput{Headers: memN1.Headers.Clone(), Body: memN1.Body}
	forge.Body = append([]byte{}, forge.Body...)
	forge.Body[len(forge.Body)-1] ^= 0x01
	if _, err := notify(forge); !errorspb.IsInfraErrorPaymentInvalidSignature(err) {
		t.Fatalf("tampered envelope: err=%v", err)
	}
	if _, err := svc.NotifyPayment(context.Background(), &esportspb.NotifyPaymentRequest{Provider: strp("nope"), Headers: map[string]string{"X": "y"}, Body: []byte("{}")}); !errorspb.IsInfraErrorPaymentUnavailable(err) {
		t.Fatalf("unknown provider: err=%v", err)
	}

	// The converger folds an aged pending order the payer abandoned.
	aged, err := uc.Create(ctx, &order.CreateIn{SKUCode: "member_month", Channel: channel})
	if err != nil {
		t.Fatalf("aged create: %v", err)
	}
	if _, err := db.ExecContext(context.Background(),
		`UPDATE pay_order SET created_at = now() - interval '30 minutes', expires_at = now() - interval '5 minutes' WHERE order_no = $1`,
		aged.Order.OrderNo); err != nil {
		t.Fatalf("age the order: %v", err)
	}
	if err := uc.ConvergeOnce(context.Background()); err != nil {
		t.Fatalf("converge: %v", err)
	}
	if got, err := uc.Get(ctx, aged.Order.OrderNo); err != nil || got.State != order.StateClosed {
		t.Fatalf("aged fold: state=%v err=%v", got, err)
	}

	// The day's reconciliation meets the channel's own book on every settled
	// order and raises nothing.
	if err := uc.ReconcileOnce(context.Background(), time.Now()); err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if !strings.Contains(logBuf.String(), "mismatch=0") {
		t.Fatalf("reconcile log:\n%s", logBuf.String())
	}
}

func cleanE2E(t *testing.T, db *ent.Client, uid string) {
	t.Helper()
	ctx := context.Background()
	for _, stmt := range []struct {
		q    string
		args []any
	}{
		{`DELETE FROM pay_order WHERE user_id = $1`, []any{uid}},
		{`DELETE FROM entitlement WHERE user_id = $1`, []any{uid}},
		{`DELETE FROM test_record WHERE user_id LIKE 'e2epayuser%'`, nil},
	} {
		if _, err := db.ExecContext(ctx, stmt.q, stmt.args...); err != nil {
			t.Fatalf("e2e clean: %v", err)
		}
	}
}
