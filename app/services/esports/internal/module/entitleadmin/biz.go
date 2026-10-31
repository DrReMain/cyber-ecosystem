package entitleadmin

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"cyber-ecosystem/shared-go/helper"
	"cyber-ecosystem/shared-go/kratos/security"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
	commonv1 "cyber-ecosystem/gen/go/cyber/shared/common/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"

	"cyber-ecosystem/app/services/esports/internal/module/entitle"
	"cyber-ecosystem/app/services/esports/internal/module/member"
	"cyber-ecosystem/app/services/esports/internal/module/profile"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

var displayZone = time.FixedZone("CST", 8*60*60) // "days left" is a calendar fact in the audience's timezone — the same law the pay module's bill window carries

var pageHydrate = ptr(int32(helper.DefaultPageSizeMax)) // one full roster page per list page, never a second round-trip

func ptr[T any](v T) *T { return &v }

const (
	// The four states the lifecycle faces read. TERMINATED is the stored
	// word; the other three fold at read time — the whole point of the
	// read-derived expiry ruling.
	displayActive       = "active"
	displayExpiringSoon = "expiring_soon"
	displayExpired      = "expired"
	displayTerminated   = "terminated"
)

const (
	// Source fold values — the ledger row's artifact mapped to the three
	// provenances the operator filters by.
	sourceOrder      = "order"
	sourceManual     = "manual"
	sourceLegacyCode = "legacy_code"
)

// DO ------------------------------------------------------------------------------------------------------------------

type MembershipFacts struct {
	// The current (or most recent) period of one member, folded for display.
	UserID      string
	GrantedAt   time.Time
	ExpiresAt   time.Time
	Status      string // display four-state
	Source      string // source fold (order / manual / legacy_code)
	SourceRef   string // the raw artifact, for support traceability
	PeriodCount int
	UnlockCount int
	DaysLeft    *int // calendar-day distance in the display zone; nil once over
}

type MembershipHistoryRow struct {
	// One renewal row of the append-only history timeline.
	GrantedAt time.Time
	ExpiresAt time.Time
	Source    string
	SourceRef string
	Status    string
}

type MembershipRow struct {
	// The folded facts with the roster/profile identity beside.
	MembershipFacts
	Nickname    *string
	PhoneMasked *string
}

type MembershipCounts struct {
	Total        int
	Active       int
	ExpiringSoon int
	Expired      int
}

type ListMembershipsIn struct {
	*commonv1.PageRequest
	Status *string
	Source *string
}

type ListMembershipsOut struct {
	*commonv1.PageResponse
	Rows   []*MembershipRow
	Counts MembershipCounts
}

type IdentityFacts struct {
	Nickname    *string
	PhoneMasked *string
	WechatBound bool
	AgeBand     *string
	Sex         *string
}

type MembershipDetail struct {
	Facts    *MembershipFacts
	History  []*MembershipHistoryRow // newest first
	Identity *IdentityFacts
}

type SkuFacts struct {
	// The sku facts a membership grant needs. A local face on purpose:
	// this consumer owns the shape it wants, and the sku side satisfies
	// it without either module importing the other's domain objects.
	Name       string
	PeriodDays int
}

// Port ----------------------------------------------------------------------------------------------------------------

type SkuFactsRP interface {
	// nil = no such code (the caller answers not-found); a kind or shape
	// mismatch fails with the sku-kind error.
	MembershipSku(ctx context.Context, code string) (*SkuFacts, error)
}

// UC -------------------------------------------------------------------------------------------------------------------

type MembershipUC struct {
	shared.UC
	entRP     entitle.EntitlementRP
	skuRP     SkuFactsRP
	systemRP  member.SystemRP
	profileRP profile.ProfileRP
}

func NewMembershipUC(
	logger *slog.Logger, tm shared.Transaction,
	entRP entitle.EntitlementRP, skuRP SkuFactsRP,
	systemRP member.SystemRP, profileRP profile.ProfileRP,
) *MembershipUC {
	return &MembershipUC{
		UC:        shared.NewUC(logger.With("module", "module/entitleadmin"), tm),
		entRP:     entRP,
		skuRP:     skuRP,
		systemRP:  systemRP,
		profileRP: profileRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *MembershipUC) Grant(ctx context.Context, userID, skuCode string) (*MembershipFacts, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("membership grant requires an admin session"))
	}
	sku, err := uc.skuRP.MembershipSku(ctx, skuCode)
	if err != nil {
		return nil, err
	}
	if err := uc.requireOnRoster(ctx, userID); err != nil {
		return nil, err
	}
	err = uc.Tm.InTx(ctx, func(txCtx context.Context) error {
		// The renewal seam: a live window (terminated rows excluded) sets
		// the start, so consecutive periods tile without gaps — a first
		// grant starts now. Two operators racing both read the same seam
		// and land overlapping rows; the operator plane accepts that
		// replay shape (the audit trail shows both), unlike the settle
		// path's pending-unique, because no money fact hangs on it.
		start := time.Now()
		if window, err := uc.entRP.MembershipWindow(txCtx, userID); err != nil {
			return err
		} else if window != nil {
			start = *window.ExpiresAt
		}
		expiry := start.AddDate(0, 0, sku.PeriodDays)
		return uc.entRP.Grant(txCtx, &entitle.EntitlementGrant{
			UserID: userID, ScopeKind: shared.ScopeMembership,
			ExpiresAt: &expiry, Source: shared.SourceManual,
		})
	})
	if err != nil {
		return nil, err
	}
	uc.Log.Info("membership granted by operator", "user_id", userID, "sku_code", skuCode, "period_days", sku.PeriodDays, "by", subject.UserID)
	return uc.foldOne(ctx, userID)
}

func (uc *MembershipUC) Terminate(ctx context.Context, userID string) (*MembershipFacts, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("membership terminate requires an admin session"))
	}
	terminated, err := uc.entRP.TerminateLiveMembership(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !terminated {
		// No live period: nothing to end. The settled answer for both a
		// member who never subscribed and one whose window already closed.
		return nil, esportspb.ErrorEsportsMembershipNotFound("")
	}
	uc.Log.Info("membership terminated by operator", "user_id", userID, "by", subject.UserID)
	return uc.foldOne(ctx, userID)
}

func (uc *MembershipUC) List(ctx context.Context, in *ListMembershipsIn) (*ListMembershipsOut, error) {
	rows, err := uc.entRP.ListMembershipRows(ctx, nil)
	if err != nil {
		return nil, err
	}
	folds := foldAll(rows, time.Now())

	out := &ListMembershipsOut{Rows: make([]*MembershipRow, 0, len(folds))}
	out.Counts.Total = len(folds)
	for _, f := range folds {
		switch f.Status {
		case displayActive:
			out.Counts.Active++
		case displayExpiringSoon:
			out.Counts.ExpiringSoon++
		case displayExpired:
			out.Counts.Expired++
		}
		if in.Status != nil && *in.Status != f.Status {
			continue
		}
		if in.Source != nil && *in.Source != f.Source {
			continue
		}
		out.Rows = append(out.Rows, &MembershipRow{MembershipFacts: *f})
	}
	// Most recently granted first: the freshest operator action tops the
	// list, lapsed members sink.
	sort.SliceStable(out.Rows, func(i, j int) bool {
		return out.Rows[i].GrantedAt.After(out.Rows[j].GrantedAt)
	})

	total := len(out.Rows)
	offset, limit := pageWindow(in.PageRequest, total)
	out.PageResponse = helper.BuildPageResponse(total, offset, limit)
	if offset >= total {
		out.Rows = nil
		return out, nil
	}
	out.Rows = out.Rows[offset:min(offset+limit, total)]

	ids := make([]string, 0, len(out.Rows))
	for _, r := range out.Rows {
		ids = append(ids, r.UserID)
	}
	identity, err := uc.hydrateIdentity(ctx, ids)
	if err != nil {
		return nil, err
	}
	unlocks, err := uc.entRP.ListMembershipUnlocks(ctx, ids)
	if err != nil {
		return nil, err
	}
	for _, r := range out.Rows {
		if id := identity[r.UserID]; id != nil {
			r.Nickname, r.PhoneMasked = id.Nickname, id.PhoneMasked
		}
		for _, u := range unlocks {
			if u.UserID == r.UserID && !u.GrantedAt.Before(r.GrantedAt) && u.GrantedAt.Before(r.ExpiresAt) {
				// An unlock inside this period's window — permanent, so the
				// count survives the period ending.
				r.UnlockCount++
			}
		}
	}
	return out, nil
}

func (uc *MembershipUC) Get(ctx context.Context, userID string) (*MembershipDetail, error) {
	rows, err := uc.entRP.ListMembershipRows(ctx, []string{userID})
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, esportspb.ErrorEsportsMembershipNotFound("")
	}
	now := time.Now()
	folds := foldAll(rows, now)
	detail := &MembershipDetail{Facts: folds[0], Identity: &IdentityFacts{}}
	// Newest first: the timeline reads top-down from today.
	for i := len(rows) - 1; i >= 0; i-- {
		row := rows[i]
		detail.History = append(detail.History, &MembershipHistoryRow{
			GrantedAt: periodStart(rows, i),
			ExpiresAt: derefTime(row.ExpiresAt),
			Source:    foldSource(row.Source),
			SourceRef: row.Source,
			Status:    rowStatus(row, now),
		})
	}
	identity, err := uc.hydrateIdentity(ctx, []string{userID})
	if err != nil {
		return nil, err
	}
	unlocks, err := uc.entRP.ListMembershipUnlocks(ctx, []string{userID})
	if err != nil {
		return nil, err
	}
	for _, u := range unlocks {
		if !u.GrantedAt.Before(detail.Facts.GrantedAt) && u.GrantedAt.Before(detail.Facts.ExpiresAt) {
			detail.Facts.UnlockCount++
		}
	}
	if id := identity[userID]; id != nil {
		detail.Identity = id
	}
	return detail, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func (uc *MembershipUC) foldOne(ctx context.Context, userID string) (*MembershipFacts, error) {
	rows, err := uc.entRP.ListMembershipRows(ctx, []string{userID})
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, esportspb.ErrorEsportsMembershipNotFound("")
	}
	return foldAll(rows, time.Now())[0], nil
}

func (uc *MembershipUC) requireOnRoster(ctx context.Context, userID string) error {
	// A hand grant answers for a real member: an id off the roster (a
	// mistyped one, or an account shown beside the id in the operator
	// list) must refuse rather than mint an orphan ledger row no read
	// plane can ever hydrate.
	roster, err := uc.systemRP.ListRoster(ctx, &member.RosterQuery{
		Population: member.DefaultPopulation,
		UserIDs:    []string{userID},
		Page:       &commonv1.PageRequest{PageSize: pageHydrate},
	})
	if err != nil {
		return err
	}
	if len(roster.Users) == 0 {
		return esportspb.ErrorEsportsMembershipNotFound("")
	}
	return nil
}

func (uc *MembershipUC) hydrateIdentity(ctx context.Context, userIDs []string) (map[string]*IdentityFacts, error) {
	// Resolves the identity header facts from the two read planes the
	// operator face already composes: the system roster (phone mask,
	// wechat binding) and the local profile (nickname, age band, sex). A
	// member absent from either plane keeps null facts — the honest lazy
	// shape, never an error.
	out := make(map[string]*IdentityFacts, len(userIDs))
	if len(userIDs) == 0 {
		return out, nil
	}
	roster, err := uc.systemRP.ListRoster(ctx, &member.RosterQuery{
		Population: member.DefaultPopulation,
		UserIDs:    userIDs,
		// One page of ids hydrates in one call — the page size cap the
		// pagination contract already names.
		Page: &commonv1.PageRequest{PageSize: pageHydrate},
	})
	if err != nil {
		return nil, err
	}
	profiles, err := uc.profileRP.ListByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	for _, id := range userIDs {
		out[id] = &IdentityFacts{}
	}
	for _, u := range roster.Users {
		if f := out[u.ID]; f != nil {
			f.PhoneMasked, f.WechatBound = u.PhoneMasked, u.WechatBound
		}
	}
	for _, p := range profiles {
		if f := out[p.UserID]; f != nil {
			f.Nickname, f.AgeBand, f.Sex = p.Nickname, p.AgeBand, p.Sex
		}
	}
	return out, nil
}

func foldAll(rows []*entitle.EntitlementRow, now time.Time) []*MembershipFacts {
	// Walks the whole membership ledger and lands one folded fact per
	// member. Rows arrive oldest-first; the run length is the period count.
	byUser := make(map[string][]*entitle.EntitlementRow)
	order := make([]string, 0)
	for _, row := range rows {
		if _, seen := byUser[row.UserID]; !seen {
			order = append(order, row.UserID)
		}
		byUser[row.UserID] = append(byUser[row.UserID], row)
	}
	out := make([]*MembershipFacts, 0, len(order))
	for _, userID := range order {
		run := byUser[userID]
		// The current period is the live window running furthest — the
		// same judgement the access plane makes. Settle can land a
		// later-created row with an earlier expiry than a tiled manual
		// window; the newest row would then fold EXPIRED while access is
		// still alive. No live row: the newest row carries the settled
		// terminal state.
		cur := -1
		for i, row := range run {
			if row.Status != shared.StatusActive || row.ExpiresAt == nil || !row.ExpiresAt.After(now) {
				continue
			}
			if cur < 0 || row.ExpiresAt.After(*run[cur].ExpiresAt) {
				cur = i
			}
		}
		if cur < 0 {
			cur = len(run) - 1
		}
		current := run[cur]
		facts := &MembershipFacts{
			UserID:      userID,
			GrantedAt:   periodStart(run, cur),
			ExpiresAt:   derefTime(current.ExpiresAt),
			Status:      rowStatus(current, now),
			Source:      foldSource(current.Source),
			SourceRef:   current.Source,
			PeriodCount: len(run),
		}
		if facts.Status == displayActive || facts.Status == displayExpiringSoon {
			days := calendarDays(now.In(displayZone), facts.ExpiresAt.In(displayZone))
			facts.DaysLeft = &days
		}
		out = append(out, facts)
	}
	return out
}

func periodStart(run []*entitle.EntitlementRow, i int) time.Time {
	// A row's window start: the first period starts at its creation
	// moment, every renewal at the prior period's expiry — the seam the
	// grant read. The row's own created_at stays the operator-action
	// timestamp (an audit fact), so the timeline's start is derived from
	// the chain, never stored twice.
	if i == 0 {
		return run[0].CreatedAt
	}
	return derefTime(run[i-1].ExpiresAt)
}

func rowStatus(row *entitle.EntitlementRow, now time.Time) string {
	// The stored word first (only termination is stored), then time does
	// the rest.
	if row.Status == shared.StatusTerminated {
		return displayTerminated
	}
	expiry := derefTime(row.ExpiresAt)
	if !now.Before(expiry) {
		return displayExpired
	}
	if calendarDays(now.In(displayZone), expiry.In(displayZone)) <= shared.ExpiringSoonDays {
		return displayExpiringSoon
	}
	return displayActive
}

func foldSource(source string) string {
	// Maps the ledger artifact to its provenance: the manual marker and
	// the legacy prefix name themselves, everything else is an order
	// number.
	switch {
	case source == shared.SourceManual:
		return sourceManual
	case len(source) > len(shared.SourceLegacyCode) && source[:len(shared.SourceLegacyCode)] == shared.SourceLegacyCode:
		return sourceLegacyCode
	default:
		return sourceOrder
	}
}

func calendarDays(from, to time.Time) int {
	// Calendar-day distance inside one zone — the number the "N days
	// left" line says, where same-day is zero, not one. Midnight anchors
	// rather than month arithmetic: a fixed-zone day is exactly 24h, so
	// the count stays true across short months (a *31 shortcut drifts a
	// day past November).
	midnight := func(t time.Time) time.Time {
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, displayZone)
	}
	return int(midnight(to).Sub(midnight(from)).Hours() / 24)
}

func pageWindow(req *commonv1.PageRequest, total int) (offset, limit int) {
	// Resolves the request's page onto an in-memory slice — the same
	// defaults and cap the query-level pagination applies (absent page =
	// 1, absent size = the default, oversized = capped).
	pageNo, pageSize := 1, helper.DefaultPageSize
	if req.GetPageNo() > 0 {
		pageNo = int(req.GetPageNo())
	}
	if req.GetPageSize() > 0 {
		pageSize = int(req.GetPageSize())
	}
	if pageSize > helper.DefaultPageSizeMax {
		pageSize = helper.DefaultPageSizeMax
	}
	return (pageNo - 1) * pageSize, pageSize
}

func derefTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}
