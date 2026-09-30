package audit

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	kaudit "cyber-ecosystem/shared-go/kratos/audit"
	"cyber-ecosystem/shared-go/kratos/security"

	commonpb "cyber-ecosystem/gen/go/cyber/shared/common/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"

	"cyber-ecosystem/app/services/system/internal/shared"
)

// DO ------------------------------------------------------------------------------------------------------------------

type AuditLog struct {
	ID            string
	CreatedAt     time.Time
	Tenant        string
	Actor         string
	PrincipalType string
	Operation     string
	HTTPMethod    string
	HTTPPath      string
	Status        int
	LatencyMs     int
	IP            string
	UserAgent     string
	DenyReason    string
}

type AuditListIn struct {
	*commonpb.PageRequest
	OrderBy    []string
	Tenant     string
	Actor      *string
	Operation  *string
	Status     *int32
	DenyReason *string
	Denied     *bool
}

type AuditListOut struct {
	*commonpb.PageResponse
	List []*AuditLog
}

// Port ----------------------------------------------------------------------------------------------------------------

type AuditRP interface {
	ListAuditLogs(ctx context.Context, in *AuditListIn) (*AuditListOut, error)
	Publish(ctx context.Context, ev *kaudit.Event)
}

// UC ------------------------------------------------------------------------------------------------------------------

type AuditUC struct {
	shared.UC
	auditRP AuditRP
}

func NewAuditUC(logger *slog.Logger, tm shared.Transaction, auditRP AuditRP) *AuditUC {
	return &AuditUC{
		UC:      shared.NewUC(logger.With("module", "module/audit"), tm),
		auditRP: auditRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *AuditUC) List(ctx context.Context, in *AuditListIn) (*AuditListOut, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("audit list requires a subject"))
	}
	// audit_log carries no TenantMixin (denied and subject-less events must
	// audit), so the read side applies the rule here: the caller's tenant,
	// plus the subject-less public events (login attempts) that belong to
	// no tenant.
	in.Tenant = subject.TenantID
	return uc.auditRP.ListAuditLogs(ctx, in)
}

func (uc *AuditUC) Emit(ctx context.Context, ev *kaudit.Event) {
	uc.auditRP.Publish(ctx, ev)
}
