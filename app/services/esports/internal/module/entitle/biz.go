package entitle

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"cyber-ecosystem/shared-go/kratos/security"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
	commonv1 "cyber-ecosystem/gen/go/cyber/shared/common/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"

	"cyber-ecosystem/app/services/esports/internal/shared"
)

const (
	codePrefix   = "AC"
	codeBodyLen  = 16
	codeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
)

var defaultPolicy = map[string]string{ // shipped defaults; report_policy rows are the only override — career open, cognitive the member plane, school codes-only
	shared.ReportCareer:    shared.PolicyOpen,
	shared.ReportCognitive: shared.PolicyMembership,
	shared.ReportSchool:    shared.PolicyClosed,
}

// DO ------------------------------------------------------------------------------------------------------------
type ActivationCode struct {
	Code       string
	ReportType string
	Quantity   int
	UsedCount  int
	ExpiresAt  time.Time
	CreatedBy  string
	CreatedAt  time.Time
}

type GenerateIn struct {
	ReportType string
	Count      int
	ExpireDays int
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
	ScopeKind string
	ScopeRef  *string
	ExpiresAt *time.Time
	Source    string
	CreatedAt time.Time
}

type ListCodesIn struct {
	*commonv1.PageRequest
	ReportType *string
}

type ListCodesOut struct {
	*commonv1.PageResponse
	Codes []*ActivationCode
}

// Port ----------------------------------------------------------------------------------------------------------
type ActivationCodeRP interface {
	CreateBatch(ctx context.Context, rows []*ActivationCode) error
	List(ctx context.Context, in *ListCodesIn) (*ListCodesOut, error)
	ByCode(ctx context.Context, code string) (*ActivationCode, error) // nil = no such code
	TakeUse(ctx context.Context, code string) (bool, error)           // the redeem's whole arbitration: one conditional UPDATE under used_count < quantity — the settle path's single-arbiter shape
}

type EntitlementRP interface {
	Grant(ctx context.Context, g *EntitlementGrant) error                                             // insert-only; a rival grant for the same scope lands as a no-op, never an error
	ByScope(ctx context.Context, userID, scopeKind string, scopeRef *string) (*EntitlementRow, error) // nil = no grant
	Exists(ctx context.Context, userID, scopeKind string, scopeRef *string) (bool, error)             // the narrow query pay's port asks of the same face
	HasActiveMembership(ctx context.Context, userID string) (bool, error)
}

type ReportPolicyRP interface {
	Policy(ctx context.Context, family string) (policy string, ok bool, err error) // ok=false means no row — the shipped default answers
	SetPolicy(ctx context.Context, family, policy, updatedBy string) error
}

// UC ------------------------------------------------------------------------------------------------------------

type CodeUC struct {
	shared.UC
	codeRP ActivationCodeRP
}

func NewCodeUC(logger *slog.Logger, tm shared.Transaction, codeRP ActivationCodeRP) *CodeUC {
	return &CodeUC{UC: shared.NewUC(logger.With("module", "module/entitle"), tm), codeRP: codeRP}
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
func (uc *CodeUC) List(ctx context.Context, in *ListCodesIn) (*ListCodesOut, error) {
	return uc.codeRP.List(ctx, in)
}

func (uc *CodeUC) Generate(ctx context.Context, in *GenerateIn) ([]string, error) {
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
	if err := uc.codeRP.CreateBatch(ctx, rows); err != nil {
		return nil, err
	}
	uc.Log.Info("activation codes generated", "count", len(rows), "report_type", in.ReportType, "created_by", subject.UserID)
	return codes, nil
}

func (uc *EntitleUC) GetReportAccess(ctx context.Context, family string) (*AccessOut, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("report access requires a session"))
	}
	// A redeemed code outranks the policy — a type code is an explicit
	// grant someone handed out; no config switch revokes it.
	ref := family
	grant, err := uc.entRP.ByScope(ctx, subject.UserID, shared.ScopeReportType, &ref)
	if err != nil {
		return nil, err
	}
	if grant != nil && (grant.ExpiresAt == nil || grant.ExpiresAt.After(time.Now())) {
		return &AccessOut{Allowed: true, Reason: esportspb.ReportAccessReason_REPORT_ACCESS_REASON_CODE_REDEEMED}, nil
	}
	policy, ok, err := uc.policyRP.Policy(ctx, family)
	if err != nil {
		return nil, err
	}
	if !ok {
		policy = defaultPolicy[family]
	}
	explicitClosed := ok && policy == shared.PolicyClosed
	switch policy {
	case shared.PolicyOpen:
		return &AccessOut{Allowed: true, Reason: esportspb.ReportAccessReason_REPORT_ACCESS_REASON_POLICY_OPEN}, nil
	case shared.PolicyMembership:
		if ok, err := uc.entRP.HasActiveMembership(ctx, subject.UserID); err != nil {
			return nil, err
		} else if ok {
			return &AccessOut{Allowed: true, Reason: esportspb.ReportAccessReason_REPORT_ACCESS_REASON_MEMBERSHIP_ACTIVE}, nil
		}
	case shared.PolicyTenant, shared.PolicyClosed:
	}
	if explicitClosed {
		return &AccessOut{Allowed: false, Reason: esportspb.ReportAccessReason_REPORT_ACCESS_REASON_POLICY_CLOSED}, nil
	}
	return &AccessOut{Allowed: false, Reason: esportspb.ReportAccessReason_REPORT_ACCESS_REASON_NO_ENTITLEMENT}, nil
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
	if row.ExpiresAt.Before(time.Now()) {
		return nil, esportspb.ErrorEsportsActivationCodeExpired("")
	}
	family := row.ReportType
	// The same member redeeming their own code again is idempotent — the
	// grant already exists and the code's single use is already theirs.
	// A lapsed grant is not that idempotent hit: an expired family row
	// must not strand the member out of every later code for the family,
	// so the redeem falls through and the grant ratchets forward — the
	// access judgement reads the same liveness rule.
	if grant, err := uc.entRP.ByScope(ctx, subject.UserID, shared.ScopeReportType, &family); err != nil {
		return nil, err
	} else if grant != nil && (grant.ExpiresAt == nil || grant.ExpiresAt.After(time.Now())) {
		return &RedeemOut{ReportType: family, ExpiresAt: grant.ExpiresAt, GrantedAt: grant.CreatedAt}, nil
	}
	var out *RedeemOut
	err = uc.Tm.InTx(ctx, func(txCtx context.Context) error {
		// One transaction takes the use and lands the grant: a use taken
		// without its grant is unrecoverable — the retry path would find
		// the code exhausted and no grant to be idempotent against.
		taken, err := uc.codeRP.TakeUse(txCtx, code)
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
				out = &RedeemOut{ReportType: family, ExpiresAt: grant.ExpiresAt, GrantedAt: grant.CreatedAt}
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
		UserID: userID, ScopeKind: shared.ScopeTrial, ScopeRef: &recordID, Source: "membership",
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

func mintCodeBody() (string, error) {
	// Sixteen chars of secure random over an alphabet with the
	// look-alikes (0/O, 1/I) removed — a handed-out code survives being
	// read aloud, the legacy corpus's own alphabet.
	b := make([]byte, codeBodyLen)
	alphabetMax := big.NewInt(int64(len(codeAlphabet)))
	for i := range b {
		n, err := rand.Int(rand.Reader, alphabetMax)
		if err != nil {
			return "", fmt.Errorf("entitle: entropy failure: %w", err)
		}
		b[i] = codeAlphabet[n.Int64()]
	}
	return string(b), nil
}
