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

// DO ------------------------------------------------------------------------------------------------------------------

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

// Port ----------------------------------------------------------------------------------------------------------------

type CodeRP interface {
	CreateBatch(ctx context.Context, rows []*ActivationCode) error
	List(ctx context.Context, in *ListCodesIn) (*ListCodesOut, error)
	ByCode(ctx context.Context, code string) (*ActivationCode, error) // nil = no such code
	// TakeUse is the redeem's whole arbitration: a conditional UPDATE
	// under used_count < quantity — one use taken per successful call,
	// the same single-arbiter shape the settle path gives money.
	TakeUse(ctx context.Context, code string) (bool, error)
}

type EntitlementRP interface {
	Grant(ctx context.Context, g *EntitlementGrant) error                                             // insert-only; a rival grant for the same scope lands as a no-op, never an error
	ByScope(ctx context.Context, userID, scopeKind string, scopeRef *string) (*EntitlementRow, error) // nil = no grant
	// Exists is the narrow query pay's port asks of the same face.
	Exists(ctx context.Context, userID, scopeKind string, scopeRef *string) (bool, error)
	HasActiveMembership(ctx context.Context, userID string) (bool, error)
}

type PolicyRP interface {
	// Policy reads the operator's override for one family; ok=false means
	// no row, and the shipped default answers.
	Policy(ctx context.Context, family string) (policy string, ok bool, err error)
	SetPolicy(ctx context.Context, family, policy, updatedBy string) error
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

// UC ------------------------------------------------------------------------------------------------------------------

const (
	reportCognitive = "cognitive"
	reportCareer    = "career"
	reportSchool    = "school"

	policyOpen       = "open"
	policyMembership = "membership"
	policyTenant     = "tenant"
	policyClosed     = "closed"

	// The code shape mirrors the legacy corpus: a two-letter prefix a
	// support agent can recognize on the phone plus sixteen characters of
	// secure random over an alphabet with the look-alikes (0/O, 1/I)
	// removed, so a handed-out code survives being read aloud.
	codePrefix   = "AC"
	codeBodyLen  = 16
	codeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
)

var defaultPolicy = map[string]string{ // shipped defaults; report_policy rows are the only override — career open, cognitive the member plane, school codes-only
	reportCareer:    policyOpen,
	reportCognitive: policyMembership,
	reportSchool:    policyClosed,
}

type CodeUC struct {
	shared.UC
	codeRP CodeRP
}

func NewCodeUC(logger *slog.Logger, tm shared.Transaction, codeRP CodeRP) *CodeUC {
	return &CodeUC{UC: shared.NewUC(logger.With("module", "module/entitle_code"), tm), codeRP: codeRP}
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
		code := codePrefix + mintCodeBody()
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

func (uc *CodeUC) List(ctx context.Context, in *ListCodesIn) (*ListCodesOut, error) {
	return uc.codeRP.List(ctx, in)
}

type EntitleUC struct {
	shared.UC
	codeRP   CodeRP
	entRP    EntitlementRP
	policyRP PolicyRP
}

func NewEntitleUC(logger *slog.Logger, tm shared.Transaction, codeRP CodeRP, entRP EntitlementRP, policyRP PolicyRP) *EntitleUC {
	return &EntitleUC{
		UC:       shared.NewUC(logger.With("module", "module/entitle"), tm),
		codeRP:   codeRP,
		entRP:    entRP,
		policyRP: policyRP,
	}
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
	if grant, err := uc.entRP.ByScope(ctx, subject.UserID, shared.ScopeReportType, &family); err != nil {
		return nil, err
	} else if grant != nil {
		return &RedeemOut{ReportType: family, ExpiresAt: grant.ExpiresAt, GrantedAt: grant.CreatedAt}, nil
	}
	taken, err := uc.codeRP.TakeUse(ctx, code)
	if err != nil {
		return nil, err
	}
	if !taken {
		// The conditional update lost: either a rival member took the use,
		// or this member's own concurrent redeem won it — re-reading the
		// grant separates the two.
		grant, err := uc.entRP.ByScope(ctx, subject.UserID, shared.ScopeReportType, &family)
		if err != nil {
			return nil, err
		}
		if grant != nil {
			return &RedeemOut{ReportType: family, ExpiresAt: grant.ExpiresAt, GrantedAt: grant.CreatedAt}, nil
		}
		return nil, esportspb.ErrorEsportsActivationCodeExhausted("")
	}
	ref := family
	expiresAt := row.ExpiresAt
	if err := uc.entRP.Grant(ctx, &EntitlementGrant{
		UserID: subject.UserID, ScopeKind: shared.ScopeReportType, ScopeRef: &ref,
		ExpiresAt: &expiresAt, Source: code,
	}); err != nil {
		return nil, err
	}
	uc.Log.Info("activation code redeemed", "code", code, "report_type", family, "user_id", subject.UserID)
	return &RedeemOut{ReportType: family, ExpiresAt: &expiresAt, GrantedAt: time.Now()}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func mintCodeBody() string {
	b := make([]byte, codeBodyLen)
	alphabetMax := big.NewInt(int64(len(codeAlphabet)))
	for i := range b {
		n, err := rand.Int(rand.Reader, alphabetMax)
		if err != nil {
			panic("entitle: entropy failure: " + err.Error())
		}
		b[i] = codeAlphabet[n.Int64()]
	}
	return string(b)
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
	explicitClosed := ok && policy == policyClosed
	switch policy {
	case policyOpen:
		return &AccessOut{Allowed: true, Reason: esportspb.ReportAccessReason_REPORT_ACCESS_REASON_POLICY_OPEN}, nil
	case policyMembership:
		if ok, err := uc.entRP.HasActiveMembership(ctx, subject.UserID); err != nil {
			return nil, err
		} else if ok {
			return &AccessOut{Allowed: true, Reason: esportspb.ReportAccessReason_REPORT_ACCESS_REASON_MEMBERSHIP_ACTIVE}, nil
		}
	case policyTenant, policyClosed:
	}
	if explicitClosed {
		return &AccessOut{Allowed: false, Reason: esportspb.ReportAccessReason_REPORT_ACCESS_REASON_POLICY_CLOSED}, nil
	}
	return &AccessOut{Allowed: false, Reason: esportspb.ReportAccessReason_REPORT_ACCESS_REASON_NO_ENTITLEMENT}, nil
}

func (uc *EntitleUC) SetReportAccessPolicy(ctx context.Context, family, policy string) error {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("policy write requires an admin session"))
	}
	if policy == policyTenant {
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
