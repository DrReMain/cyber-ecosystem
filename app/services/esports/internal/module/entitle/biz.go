package entitle

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"cyber-ecosystem/shared-go/kratos/security"
	"cyber-ecosystem/shared-go/utils"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
	commonv1 "cyber-ecosystem/gen/go/cyber/shared/common/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"

	"cyber-ecosystem/app/services/esports/internal/module/profile"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

const (
	codePrefix   = "AC"
	codeBodyLen  = 16
	codeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
)

var reportFamilies = []string{shared.ReportCognitive, shared.ReportCareer, shared.ReportSchool}

// DO ------------------------------------------------------------------------------------------------------------
type ActivationCode struct {
	Code       string
	ReportType string
	Quantity   int
	UsedCount  int
	ExpiresAt  time.Time
	CreatedBy  string
	CreatedAt  time.Time
	// The minting batch; nil only on rows predating the batch model.
	BatchID *string
	// BatchNo/BatchDisabled hydrate from the batch row the code points at —
	// the redeem's disable check reads the latter.
	BatchNo       *string
	BatchDisabled bool
	// The redeem's trail, written by the take-use arbitration itself: nil
	// on an unredeemed code. The grant the redeem produced is not this
	// trail's source — a family ratchet rewrites it.
	RedeemedBy *string
	RedeemedAt *time.Time
	// Profile fold the operator reads beside the redeem; hydrated in the UC.
	Nickname *string
}

type ActivationBatch struct {
	// Human-shape batch number (B-YYMM-XXXX) — the string support quotes.
	BatchNo string
	// The report family every code in the batch opens.
	ReportType string
	// The shared validity deadline.
	ExpiresAt time.Time
	// Nil = live batch; the flip is the whole mutable truth on this row.
	DisabledAt *time.Time
	CreatedAt  time.Time
	// The admin identity that minted the batch (audit join).
	CreatedBy string
	// Aggregates over the code rows, computed at read time so the batch
	// row and the codes can never drift.
	Count     int
	UsedCount int
}

type PolicyRow struct {
	Family     string
	Policy     string
	Overridden bool // false = the shipped default answering, true = a report_policy row governs
	UpdatedBy  string
	UpdatedAt  time.Time
}

type GenerateIn struct {
	ReportType string
	Count      int
	ExpireDays int
}

type GenerateOut struct {
	Codes   []string
	BatchNo string
}

type RedeemOut struct {
	ReportType string
	ExpiresAt  *time.Time
	GrantedAt  time.Time
}

type AccessOut struct {
	Allowed bool
	Reason  esportspb.ReportAccessReason
}

type EntitlementGrant struct {
	UserID    string
	ScopeKind string
	ScopeRef  *string
	ExpiresAt *time.Time
	Source    string
}

type EntitlementRow struct {
	UserID       string
	ScopeKind    string
	ScopeRef     *string
	ExpiresAt    *time.Time
	Source       string
	CreatedAt    time.Time
	Status       string
	TerminatedAt *time.Time
}

const (
	CodeStatusUnused   = "unused" // batch-level facts (disabled, expired) never enter the row state — they stay on the batch
	CodeStatusRedeemed = "redeemed"
)

type ListCodesIn struct {
	*commonv1.PageRequest
	ReportType *string
	// Present = one batch's code table; absent = the flat all-codes view.
	BatchNo *string
	Status  *string
	// Present = the exact code the detail page's search box chases.
	Code *string
	// Present = the member's own view — the codes this member redeemed.
	User *string
}

type ListCodesOut struct {
	*commonv1.PageResponse
	Codes []*ActivationCode
}

type ListBatchesIn struct {
	*commonv1.PageRequest
	ReportType *string
	// Nil = every batch; true/false = disabled/live only.
	Disabled *bool
	// Exact-match lookup — the detail page's header facts in one answer.
	BatchNo *string
}

type ListBatchesOut struct {
	*commonv1.PageResponse
	Batches []*ActivationBatch
	// Fold over every batch, filters ignored — the footer's one line.
	TotalBatches int
	TotalCodes   int
	UsedCodes    int
}

type WindowUnlockRow struct {
	// The grant time is the whole fact — which period it falls inside is
	// the caller's window arithmetic.
	UserID    string
	GrantedAt time.Time
}

// Port ----------------------------------------------------------------------------------------------------------
type ActivationCodeRP interface {
	List(ctx context.Context, in *ListCodesIn) (*ListCodesOut, error)
	// The export's whole-filter shape: every matching row, no paging.
	ListAll(ctx context.Context, in *ListCodesIn) ([]*ActivationCode, error)
	ByCode(ctx context.Context, code string) (*ActivationCode, error)   // nil = no such code
	TakeUse(ctx context.Context, code, redeemedBy string) (bool, error) // the redeem's whole arbitration: one conditional UPDATE under used_count < quantity and a live batch, landing the use and its trail together — the settle path's single-arbiter shape
	// One transaction lands the batch row and its codes — a batch whose
	// codes never landed must not exist.
	MintBatch(ctx context.Context, batch *ActivationBatch, rows []*ActivationCode) error
	// The status flip's whole arbitration: one conditional UPDATE on the
	// batch row — disabling only a live batch, re-enabling only a disabled
	// one. Zero rows = idempotent repeat; the batch re-reads either way.
	SetBatchDisabled(ctx context.Context, batchNo string, disabled bool) (*ActivationBatch, error)
	ListBatches(ctx context.Context, in *ListBatchesIn) (*ListBatchesOut, error)
}

type EntitlementRP interface {
	Grant(ctx context.Context, g *EntitlementGrant) error                                             // insert-only; a rival grant for the same scope lands as a no-op, never an error
	ByScope(ctx context.Context, userID, scopeKind string, scopeRef *string) (*EntitlementRow, error) // nil = no grant
	Exists(ctx context.Context, userID, scopeKind string, scopeRef *string) (bool, error)             // the narrow query pay's port asks of the same face
	HasActiveMembership(ctx context.Context, userID string) (bool, error)
	MembershipWindow(ctx context.Context, userID string) (*EntitlementRow, error) // the live membership row with the furthest expiry; nil = none
	// The history-telling twin above: the most recent non-terminated
	// window's end, past or future; nil = never a member.
	LastMembershipUntil(ctx context.Context, userID string) (*time.Time, error)
	HasAnyScope(ctx context.Context, userID, scopeKind string) (bool, error)      // any row of the kind, ref-agnostic
	// Membership rows for the lifecycle faces: every member's whole history
	// when ids is nil, the named members' rows otherwise — the operator
	// plane folds statuses in the use case, so the read stays one shape.
	ListMembershipRows(ctx context.Context, userIDs []string) ([]*EntitlementRow, error)
	// The terminate action's whole arbitration: one conditional UPDATE —
	// status still active, window still open — flipping exactly the live
	// row. False = no live period to terminate (a settled fact, not a race
	// to report: the caller answers not-found).
	TerminateLiveMembership(ctx context.Context, userID string) (bool, error)
	// Serializes the rival writers to one member's scope: the redeem's
	// read-judge-take window and the membership writers' seam read share
	// this transaction-scoped advisory lock, so no two transactions race
	// the same grant target. Advisory because the first grant to a scope
	// has no row to lock yet; xact-scoped because it releases with the
	// caller's commit.
	LockScope(ctx context.Context, userID, scopeKind string, scopeRef *string) error
	// The reconciliation's missing-grant probe: whether any row names this
	// source — a settled order owes exactly one. Narrower than ByScope on
	// purpose: a member's standing rows from other sources must not mask
	// this order's missing grant.
	ExistsSource(ctx context.Context, userID, source string) (bool, error)
	// Member-window trial grants for the named members — the count behind
	// "unlocked inside this period, permanently".
	ListMembershipUnlocks(ctx context.Context, userIDs []string) ([]*WindowUnlockRow, error)
}

type ReportPolicyRP interface {
	Policy(ctx context.Context, family string) (policy string, ok bool, err error) // ok=false means no row — the shipped default answers
	SetPolicy(ctx context.Context, family, policy, updatedBy string) error
	List(ctx context.Context) ([]*PolicyRow, error) // override rows only — defaults fold in at the UC
}

// UC ------------------------------------------------------------------------------------------------------------

type CodeUC struct {
	shared.UC
	codeRP    ActivationCodeRP
	profileRP profile.ProfileRP
}

func NewCodeUC(logger *slog.Logger, tm shared.Transaction, codeRP ActivationCodeRP, profileRP profile.ProfileRP) *CodeUC {
	return &CodeUC{
		UC:        shared.NewUC(logger.With("module", "module/entitle"), tm),
		codeRP:    codeRP,
		profileRP: profileRP,
	}
}

type EntitleUC struct {
	shared.UC
	codeRP   ActivationCodeRP
	entRP    EntitlementRP
	policyRP ReportPolicyRP
	eventBus shared.DomainEventBus
}

func NewEntitleUC(logger *slog.Logger, tm shared.Transaction, codeRP ActivationCodeRP, entRP EntitlementRP, policyRP ReportPolicyRP, eventBus shared.DomainEventBus) *EntitleUC {
	uc := &EntitleUC{
		UC:       shared.NewUC(logger.With("module", "module/entitle"), tm),
		codeRP:   codeRP,
		entRP:    entRP,
		policyRP: policyRP,
		eventBus: eventBus,
	}
	// The subscription is this module's whole coupling to the completion
	// fact: the handler runs on the bus, the writer of the grant stays
	// here, and no module imports another.
	eventBus.SubscribeTrialCompleted(uc.onTrialCompleted)
	return uc
}

// Method --------------------------------------------------------------------------------------------------------
func (uc *CodeUC) Generate(ctx context.Context, in *GenerateIn) (*GenerateOut, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("code generation requires an admin session"))
	}
	expiresAt := time.Now().AddDate(0, 0, in.ExpireDays)
	rows := make([]*ActivationCode, 0, in.Count)
	codes := make([]string, 0, in.Count)
	seen := make(map[string]struct{}, in.Count)
	for len(rows) < in.Count {
		// Two-letter prefix a support agent can recognize on the phone,
		// then the random body — the code shape mirrors the legacy corpus.
		body, err := mintCodeBody()
		if err != nil {
			return nil, err
		}
		code := codePrefix + body
		if _, dup := seen[code]; dup {
			continue
		}
		seen[code] = struct{}{}
		codes = append(codes, code)
		rows = append(rows, &ActivationCode{
			Code: code, ReportType: in.ReportType, Quantity: 1, UsedCount: 0,
			ExpiresAt: expiresAt, CreatedBy: subject.UserID,
		})
	}
	// The batch number rides a collision-retry loop: the unique index is
	// the only arbiter of "this number is taken".
	var minted *ActivationBatch
	var lastErr error
	for range 3 {
		no, err := mintBatchNo()
		if err != nil {
			return nil, err
		}
		batch := &ActivationBatch{BatchNo: no, ReportType: in.ReportType, ExpiresAt: expiresAt, CreatedBy: subject.UserID}
		err = uc.Tm.InTx(ctx, func(txCtx context.Context) error {
			return uc.codeRP.MintBatch(txCtx, batch, rows)
		})
		if err == nil {
			minted = batch
			break
		}
		// Identity on the choke point's classification, via the generated
		// reason predicate — never a fresh ent-class at the call site.
		if !errorspb.IsInfraErrorDbConstraint(err) {
			return nil, err
		}
		lastErr = err
	}
	if minted == nil {
		return nil, fmt.Errorf("minting the batch hit a constraint three times in a row: %w", lastErr)
	}
	uc.Log.Info("activation codes generated", "count", len(rows), "report_type", in.ReportType, "created_by", subject.UserID, "batch_no", minted.BatchNo)
	return &GenerateOut{Codes: codes, BatchNo: minted.BatchNo}, nil
}

func (uc *CodeUC) List(ctx context.Context, in *ListCodesIn) (*ListCodesOut, error) {
	out, err := uc.codeRP.List(ctx, in)
	if err != nil {
		return nil, err
	}
	if err := uc.hydrateRedeems(ctx, out.Codes); err != nil {
		return nil, err
	}
	return out, nil
}

func (uc *CodeUC) Export(ctx context.Context, in *ListCodesIn) ([]*ActivationCode, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("code export requires an admin session"))
	}
	codes, err := uc.codeRP.ListAll(ctx, in)
	if err != nil {
		return nil, err
	}
	if err := uc.hydrateRedeems(ctx, codes); err != nil {
		return nil, err
	}
	// The export hands out usable code values — its own trace line even
	// before the audit middleware joins this service's ladder.
	uc.Log.Info("activation codes exported", "count", len(codes), "batch_no", utils.Deref(in.BatchNo, ""), "by", subject.UserID)
	return codes, nil
}

func (uc *CodeUC) SetBatchStatus(ctx context.Context, batchNo string, disabled bool) (*ActivationBatch, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("batch status flip requires an admin session"))
	}
	batch, err := uc.codeRP.SetBatchDisabled(ctx, batchNo, disabled)
	if err != nil {
		return nil, err
	}
	uc.Log.Info("activation batch status flipped", "batch_no", batchNo, "disabled", disabled, "by", subject.UserID)
	return batch, nil
}

func (uc *CodeUC) ListBatches(ctx context.Context, in *ListBatchesIn) (*ListBatchesOut, error) {
	return uc.codeRP.ListBatches(ctx, in)
}

func (uc *EntitleUC) GetReportAccess(ctx context.Context, family string) (*AccessOut, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("report access requires a session"))
	}
	// The three facts, then the order's one home: the shared verdict
	// this face and the report cluster faces both quote, so a drift
	// between them is structurally impossible.
	ref := family
	grant, err := uc.entRP.ByScope(ctx, subject.UserID, shared.ScopeReportType, &ref)
	if err != nil {
		return nil, err
	}
	codeActive := grant != nil && (grant.ExpiresAt == nil || grant.ExpiresAt.After(time.Now()))
	policy, rowPresent, err := uc.policyRP.Policy(ctx, family)
	if err != nil {
		return nil, err
	}
	memberActive, err := uc.entRP.HasActiveMembership(ctx, subject.UserID)
	if err != nil {
		return nil, err
	}
	allowed, reason := shared.ReportAccessVerdict(family, codeActive, rowPresent, policy, memberActive)
	return &AccessOut{Allowed: allowed, Reason: reason}, nil
}

func (uc *EntitleUC) MemberUntil(ctx context.Context, userID string) (*time.Time, error) {
	// The profile face's greet port (verdict-port shape): profile owns the
	// RPC face but already sits under this module's imports, so this module
	// implements the port profile declares. Membership rows are born
	// bounded — trial nulls never reach this read, the window query's GT
	// filter drops them — so the guard below keeps the port's nil contract
	// honest, it is not a fourth liveness state.
	window, err := uc.entRP.MembershipWindow(ctx, userID)
	if err != nil {
		return nil, err
	}
	if window == nil || window.ExpiresAt == nil {
		return nil, nil
	}
	return window.ExpiresAt, nil
}

func (uc *EntitleUC) LastMembershipUntil(ctx context.Context, userID string) (*time.Time, error) {
	// The greet port's history twin — the same verdict-port shape, the
	// nil contract already the RP's own answer.
	return uc.entRP.LastMembershipUntil(ctx, userID)
}

func (uc *EntitleUC) Redeem(ctx context.Context, code string) (*RedeemOut, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("redeem requires a session"))
	}
	row, err := uc.codeRP.ByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, esportspb.ErrorEsportsActivationCodeInvalid("")
	}
	family := row.ReportType
	// The same member redeeming their own code again is idempotent — the
	// grant already exists and the code's single use is already theirs,
	// and that answer outranks both refusals below: an expired code or a
	// disabled batch voids unredeemed uses, neither reopens a grant
	// already landed. A lapsed grant is not that idempotent hit: an
	// expired family row must not strand the member out of every later
	// code for the family, so the redeem falls through and the grant
	// ratchets forward — the access judgement reads the same liveness
	// rule.
	if grant, err := uc.entRP.ByScope(ctx, subject.UserID, shared.ScopeReportType, &family); err != nil {
		return nil, err
	} else if grantStanding(grant) {
		return idempotentOut(family, grant), nil
	}
	if row.ExpiresAt.Before(time.Now()) {
		return nil, esportspb.ErrorEsportsActivationCodeExpired("")
	}
	if row.BatchDisabled {
		// The operator pulled the batch — distinct from expired and from
		// invalid, so support can say exactly which of the three it was.
		return nil, esportspb.ErrorEsportsActivationCodeDisabled("")
	}
	var out *RedeemOut
	err = uc.Tm.InTx(ctx, func(txCtx context.Context) error {
		// The fast read above cannot close the two-code race: two codes of
		// one family redeeming concurrently past an expired grant both read
		// "lapsed" outside and both take their use — silently burning one
		// code's option that the sequential shape keeps in the member's
		// hand. The advisory lock serializes the family's redeem writers;
		// the re-read inside answers the race — the winner's committed
		// grant is visible here, the loser returns idempotent without
		// taking any use.
		if err := uc.entRP.LockScope(txCtx, subject.UserID, shared.ScopeReportType, &family); err != nil {
			return err
		}
		if grant, err := uc.entRP.ByScope(txCtx, subject.UserID, shared.ScopeReportType, &family); err != nil {
			return err
		} else if grantStanding(grant) {
			out = idempotentOut(family, grant)
			return nil
		}
		// One transaction takes the use and lands the grant: a use taken
		// without its grant is unrecoverable — the retry path would find
		// the code exhausted and no grant to be idempotent against.
		taken, err := uc.codeRP.TakeUse(txCtx, code, subject.UserID)
		if err != nil {
			return err
		}
		if !taken {
			// The conditional update lost: either a rival member took the
			// use, or this member's own concurrent redeem won it — the
			// re-read separates the two.
			grant, err := uc.entRP.ByScope(txCtx, subject.UserID, shared.ScopeReportType, &family)
			if err != nil {
				return err
			}
			if grant != nil {
				out = idempotentOut(family, grant)
				return nil
			}
			return esportspb.ErrorEsportsActivationCodeExhausted("")
		}
		ref := family
		expiresAt := row.ExpiresAt
		if err := uc.entRP.Grant(txCtx, &EntitlementGrant{
			UserID: subject.UserID, ScopeKind: shared.ScopeReportType, ScopeRef: &ref,
			ExpiresAt: &expiresAt, Source: code,
		}); err != nil {
			return err
		}
		out = &RedeemOut{ReportType: family, ExpiresAt: &expiresAt, GrantedAt: time.Now()}
		return nil
	})
	if err != nil {
		return nil, err
	}
	uc.Log.Info("activation code redeemed", "code", code, "report_type", family, "user_id", subject.UserID)
	return out, nil
}

func (uc *EntitleUC) ListReportAccessPolicies(ctx context.Context) ([]*PolicyRow, error) {
	rows, err := uc.policyRP.List(ctx)
	if err != nil {
		return nil, err
	}
	byFamily := make(map[string]*PolicyRow, len(rows))
	for _, r := range rows {
		byFamily[r.Family] = r
	}
	// The absent families answer with their shipped defaults — the operator
	// sees one settled table, never "no row" as a third state.
	out := make([]*PolicyRow, 0, len(reportFamilies))
	for _, family := range reportFamilies {
		if r, ok := byFamily[family]; ok {
			out = append(out, r)
			continue
		}
		out = append(out, &PolicyRow{Family: family, Policy: shared.DefaultReportPolicy[family]})
	}
	return out, nil
}

func (uc *EntitleUC) SetReportAccessPolicy(ctx context.Context, family, policy string) error {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("policy write requires an admin session"))
	}
	if policy == shared.PolicyTenant {
		// The tenant plane is not here yet — refusing the value keeps the
		// operator from setting a policy this build would silently answer
		// as closed.
		return esportspb.ErrorEsportsReportPolicyNotReady("")
	}
	if err := uc.policyRP.SetPolicy(ctx, family, policy, subject.UserID); err != nil {
		return err
	}
	uc.Log.Info("report access policy set", "report_type", family, "policy", policy, "updated_by", subject.UserID)
	return nil
}

// Private -------------------------------------------------------------------------------------------------------

// grantStanding: the liveness read the redeem's idempotent answer rides —
// a family row with no expiry, or one still in the future.
func grantStanding(grant *EntitlementRow) bool {
	return grant != nil && (grant.ExpiresAt == nil || grant.ExpiresAt.After(time.Now()))
}

// idempotentOut: the answer a standing family grant gives — the grant's
// own facts, never a fresh mint.
func idempotentOut(family string, grant *EntitlementRow) *RedeemOut {
	return &RedeemOut{ReportType: family, ExpiresAt: grant.ExpiresAt, GrantedAt: grant.CreatedAt}
}

func (uc *EntitleUC) onTrialCompleted(ctx context.Context, e shared.TrialCompleted) {
	// The member plane's derived grant: a valid trial finished while a
	// membership is active unlocks that record permanently. The ruling
	// names no retroactivity — the subscription firing at completion time
	// is what enforces it. Best-effort by contract: the attempt is stored
	// truth no grant failure may destroy, so errors land as logs, never on
	// the publisher.
	userID, recordID := e.UserID, e.RecordID
	active, err := uc.entRP.HasActiveMembership(ctx, userID)
	if err != nil {
		uc.Log.Error("membership trial grant check failed", "user_id", userID, "record_id", recordID, "err", err)
		return
	}
	if !active {
		return
	}
	if err := uc.entRP.Grant(ctx, &EntitlementGrant{
		UserID: userID, ScopeKind: shared.ScopeTrial, ScopeRef: &recordID, Source: shared.SourceMembership,
	}); err != nil {
		uc.Log.Error("membership trial grant failed", "user_id", userID, "record_id", recordID, "err", err)
		return
	}
	if uc.eventBus != nil {
		// The grant named its record, so the freeze can happen now — the
		// judgement pins at unlock time rather than at first read.
		uc.eventBus.PublishReportUnlocked(ctx, shared.ReportUnlocked{UserID: userID, RecordID: recordID})
	}
}

func (uc *CodeUC) hydrateRedeems(ctx context.Context, codes []*ActivationCode) error {
	// Folds the local profile nickname onto each redeemed code's member —
	// the who is the code row's own fact; only the readable name comes
	// from outside it.
	ids := make([]string, 0, len(codes))
	seen := make(map[string]struct{}, len(codes))
	for _, c := range codes {
		if c.RedeemedBy != nil {
			if _, dup := seen[*c.RedeemedBy]; !dup {
				seen[*c.RedeemedBy] = struct{}{}
				ids = append(ids, *c.RedeemedBy)
			}
		}
	}
	if len(ids) == 0 {
		return nil
	}
	profiles, err := uc.profileRP.ListByUserIDs(ctx, ids)
	if err != nil {
		return err
	}
	nicknames := make(map[string]*string, len(profiles))
	for _, p := range profiles {
		if p.Nickname != nil && *p.Nickname != "" {
			nicknames[p.UserID] = p.Nickname
		}
	}
	for _, c := range codes {
		if c.RedeemedBy != nil {
			c.Nickname = nicknames[*c.RedeemedBy]
		}
	}
	return nil
}

func mintCodeBody() (string, error) {
	// Sixteen chars of secure random over an alphabet with the
	// look-alikes (0/O, 1/I) removed — a handed-out code survives being
	// read aloud, the legacy corpus's own alphabet.
	b := make([]byte, codeBodyLen)
	if err := fillRandom(b); err != nil {
		return "", err
	}
	return string(b), nil
}

func mintBatchNo() (string, error) {
	// B-YYMM-XXXX: the prefix and month a support agent can parse on the
	// phone, four random chars over the code alphabet — a batch number
	// is spoken far more often than it is typed.
	b := make([]byte, 4)
	if err := fillRandom(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("B-%s-%s", time.Now().Format("0601"), string(b)), nil
}

func fillRandom(b []byte) error {
	alphabetMax := big.NewInt(int64(len(codeAlphabet)))
	for i := range b {
		n, err := rand.Int(rand.Reader, alphabetMax)
		if err != nil {
			return fmt.Errorf("entitle: entropy failure: %w", err)
		}
		b[i] = codeAlphabet[n.Int64()]
	}
	return nil
}
