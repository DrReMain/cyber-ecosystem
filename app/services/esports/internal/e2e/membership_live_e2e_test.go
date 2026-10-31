package e2e

// The membership lifecycle's live probe: the operator's grant (first
// period from now, renewal from the live window's expiry), the access
// verdict flipping with the window, the member-window unlock count, and
// termination — access falls back the same moment while permanent unlocks
// stay. Gated: ENT_E2E=1 with the dev database up; anything else skips.

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
	commonv1 "cyber-ecosystem/gen/go/cyber/shared/common/v1"

	"cyber-ecosystem/app/services/esports/internal/bootstrap"
	"cyber-ecosystem/app/services/esports/internal/conf"
	"cyber-ecosystem/app/services/esports/internal/ent"
	"cyber-ecosystem/app/services/esports/internal/ent/entitlement"
	"cyber-ecosystem/app/services/esports/internal/module/entitle"
	"cyber-ecosystem/app/services/esports/internal/module/entitleadmin"
	"cyber-ecosystem/app/services/esports/internal/module/member"
	"cyber-ecosystem/app/services/esports/internal/module/pay"
	"cyber-ecosystem/app/services/esports/internal/module/profile"
	"cyber-ecosystem/app/services/esports/internal/platform"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

// The roster and profile planes are cross-service reads the lifecycle page
// hydrates best-effort. The roster stub echoes the asked-for ids back (the
// grant's roster check must pass); the profile plane answers empty so the
// identity fields stay honestly null.
type stubRoster struct{}

func (stubRoster) ListRoster(_ context.Context, q *member.RosterQuery) (*member.RosterPage, error) {
	users := make([]*member.RosterUser, 0, len(q.UserIDs))
	for _, id := range q.UserIDs {
		users = append(users, &member.RosterUser{ID: id})
	}
	return &member.RosterPage{PageResponse: &commonv1.PageResponse{}, Users: users}, nil
}

type stubProfile struct{}

func (stubProfile) FindByUserID(context.Context, string) (*profile.Profile, error) { return nil, nil }
func (stubProfile) ListByUserIDs(context.Context, []string) ([]*profile.Profile, error) {
	return nil, nil
}
func (stubProfile) ListUserIDs(context.Context, *string, *string) ([]string, error) { return nil, nil }
func (stubProfile) Save(context.Context, *profile.Profile) (*profile.Profile, error) {
	return nil, nil
}

func TestMembershipLiveE2E(t *testing.T) {
	if os.Getenv("ENT_E2E") != "1" {
		t.Skip("set ENT_E2E=1 with the dev database up to run the membership lifecycle end-to-end")
	}
	const (
		admin1 = "memadmuser0000000001"
		userA  = "meme2euser0000000001"
		userB  = "meme2euser0000000002"
	)
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	db, closeDB, err := platform.NewEntClient(&conf.Data{Database: &conf.Data_Database{
		Driver: "postgres", Host: entEnvOr("DB_HOST", "localhost"), Port: 5432,
		User: "postgres", Password: entEnvOr("DB_PASS", "postgres"), DbName: entEnvOr("DB_NAME", "esports"),
		MaxOpenConns: 8,
	}}, logger)
	if err != nil {
		t.Fatalf("ent client: %v", err)
	}
	// Registered the moment the closer exists: a platform-wiring failure
	// below must not leak the open client for the rest of the binary's run.
	t.Cleanup(func() {
		cleanMembershipE2E(t, db)
		closeDB()
	})
	pf, err := platform.NewPlatform(db, platform.NewEntErrorHandler(), nil, nil)
	if err != nil {
		t.Fatalf("platform: %v", err)
	}
	cleanMembershipE2E(t, db)

	entRP := entitle.NewEntitlementRP(logger, pf)
	skuRP := entitleadmin.NewSkuFactsRP(pay.NewSkuRP(logger, pf))
	uc := entitleadmin.NewMembershipUC(logger, pf, entRP, skuRP, stubRoster{}, stubProfile{})
	adminCtx := entSubject(admin1)
	userCtx := entSubject(userA)

	// The refusal arm first: a trial sku never hand-sells a window.
	if _, err := uc.Grant(adminCtx, userA, "trial_report"); err == nil || !esportspb.IsEsportsSkuKindInvalid(err) {
		t.Fatalf("trial sku grant: err=%v, want sku kind invalid", err)
	}
	if _, err := uc.Grant(adminCtx, userA, "no_such_sku"); err == nil || !esportspb.IsEsportsSkuNotFound(err) {
		t.Fatalf("unknown sku grant: err=%v, want not found", err)
	}

	// The dev ledger is shared (walkthrough rows live beside probe rows), so
	// the count assertions below speak in deltas over this baseline.
	baseOut, err := uc.List(adminCtx, &entitleadmin.ListMembershipsIn{PageRequest: &commonv1.PageRequest{}})
	if err != nil {
		t.Fatalf("baseline list: %v", err)
	}
	baseTotal, baseActive := baseOut.Counts.Total, baseOut.Counts.Active

	// First grant: a period starting now, active, a full month left.
	first, err := uc.Grant(adminCtx, userA, "member_month")
	if err != nil {
		t.Fatalf("first grant: %v", err)
	}
	if first.PeriodCount != 1 || first.Source != "manual" || first.DaysLeft == nil || *first.DaysLeft < 30 || *first.DaysLeft > 31 {
		t.Fatalf("first grant fold: %+v", first)
	}
	windowEnd := first.ExpiresAt

	// Renewal: the seam is the live window's expiry — the second period
	// starts where the first ends, no gap, no overlap.
	second, err := uc.Grant(adminCtx, userA, "member_month")
	if err != nil {
		t.Fatalf("renewal grant: %v", err)
	}
	if second.PeriodCount != 2 || !second.GrantedAt.Equal(windowEnd) {
		t.Fatalf("renewal fold: %+v, want granted_at == prior expiry %v", second, windowEnd)
	}

	// Access: the membership family flips to MEMBERSHIP_ACTIVE while the
	// window stands.
	accessUC := entitle.NewEntitleUC(logger, pf, entitle.NewActivationCodeRP(logger, pf), entRP, entitle.NewReportPolicyRP(logger, pf), bootstrap.NewEventSubscriptions(logger))
	if out, err := accessUC.GetReportAccess(userCtx, shared.ReportCognitive); err != nil || !out.Allowed || out.Reason != esportspb.ReportAccessReason_REPORT_ACCESS_REASON_MEMBERSHIP_ACTIVE {
		t.Fatalf("access under membership: out=%+v err=%v", out, err)
	}

	// A member-window unlock inside the *current* period counts; one made
	// in an earlier period belongs to that period and never leaks forward.
	// userA renewed, so its window moved on — a fresh unlock now falls
	// outside the current period. userB holds one period, so the same
	// unlock counts there.
	recordA := "meme2erecord0000001"
	if err := entRP.Grant(userCtx, &entitle.EntitlementGrant{
		UserID: userA, ScopeKind: shared.ScopeTrial, ScopeRef: &recordA, Source: shared.SourceMembership,
	}); err != nil {
		t.Fatalf("window unlock grant: %v", err)
	}
	if _, err := uc.Grant(adminCtx, userB, "member_month"); err != nil {
		t.Fatalf("userB grant: %v", err)
	}
	recordB := "meme2erecord0000002"
	if err := entRP.Grant(entSubject(userB), &entitle.EntitlementGrant{
		UserID: userB, ScopeKind: shared.ScopeTrial, ScopeRef: &recordB, Source: shared.SourceMembership,
	}); err != nil {
		t.Fatalf("userB unlock grant: %v", err)
	}
	probePage := int32(100) // wide page: the probe rows must land on it even beside walkthrough data
	listOut, err := uc.List(adminCtx, &entitleadmin.ListMembershipsIn{PageRequest: &commonv1.PageRequest{PageSize: &probePage}})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if listOut.Counts.Total-baseTotal != 2 || listOut.Counts.Active-baseActive != 2 {
		t.Fatalf("list counts delta: got %+v, base total=%d active=%d", listOut.Counts, baseTotal, baseActive)
	}
	for _, row := range listOut.Rows {
		if row.UserID == userA && row.UnlockCount != 0 {
			t.Fatalf("userA prior-period unlock leaked into current period: %+v", row)
		}
		if row.UserID == userB && row.UnlockCount != 1 {
			t.Fatalf("userB current-period unlock count: %+v", row)
		}
	}

	// Terminate: access falls back the same moment; the unlock row and
	// the terminated period both stay in history.
	terminated, err := uc.Terminate(adminCtx, userA)
	if err != nil {
		t.Fatalf("terminate: %v", err)
	}
	if terminated.Status != "terminated" {
		t.Fatalf("terminate fold: %+v", terminated)
	}
	if out, err := accessUC.GetReportAccess(userCtx, shared.ReportCognitive); err != nil || out.Allowed || out.Reason == esportspb.ReportAccessReason_REPORT_ACCESS_REASON_MEMBERSHIP_ACTIVE {
		t.Fatalf("access after terminate: out=%+v err=%v", out, err)
	}
	rows, err := db.Entitlement.Query().
		Where(entitlement.UserIDEQ(userA), entitlement.ScopeKindEQ(shared.ScopeTrial)).
		All(userCtx)
	if err != nil || len(rows) != 1 {
		t.Fatalf("permanent unlock rows after terminate: rows=%d err=%v", len(rows), err)
	}
	detail, err := uc.Get(adminCtx, userA)
	if err != nil || len(detail.History) != 2 {
		t.Fatalf("history after terminate: detail=%+v err=%v", detail, err)
	}
	if detail.Facts.UnlockCount != 0 {
		t.Fatalf("current-period unlock count after terminate: %+v", detail.Facts)
	}
	if detail.History[0].Status != "terminated" || detail.History[1].Status != "terminated" {
		t.Fatalf("history rows after terminate: %+v %+v", detail.History[0], detail.History[1])
	}

	// The settled refusal: nothing live left to terminate.
	if _, err := uc.Terminate(adminCtx, userA); err == nil || !esportspb.IsEsportsMembershipNotFound(err) {
		t.Fatalf("re-terminate: err=%v, want membership not found", err)
	}

	// The settle shape: pay computes expiry from PaidAt, so a settled row
	// can arrive later than a tiled manual window yet expire earlier — the
	// fold answers with the furthest live window, never the newest row.
	shortExpiry := time.Now().Add(5 * 24 * time.Hour)
	if err := db.Entitlement.Create().
		SetUserID(userB).
		SetScopeKind(shared.ScopeMembership).
		SetExpiresAt(shortExpiry).
		SetSource("E2E-SETTLE-SHORT").
		Exec(userCtx); err != nil {
		t.Fatalf("settle-shaped row: %v", err)
	}
	settled, err := uc.Get(adminCtx, userB)
	if err != nil {
		t.Fatalf("settle-shaped get: %v", err)
	}
	if settled.Facts.PeriodCount != 2 || !settled.Facts.ExpiresAt.After(shortExpiry) {
		t.Fatalf("fold must answer with the furthest live window, not the newest row: %+v", settled.Facts)
	}
}

func cleanMembershipE2E(t *testing.T, db *ent.Client) {
	t.Helper()
	// Raw SQL on purpose: the tenant interceptor scopes typed queries to
	// the caller's tenant, and a cleaner with no subject must still reach
	// every row the probe wrote.
	if _, err := db.ExecContext(context.Background(),
		`DELETE FROM entitlement WHERE user_id LIKE 'meme2e%' OR user_id LIKE 'memadm%'`); err != nil {
		t.Fatalf("clean entitlement: %v", err)
	}
}
