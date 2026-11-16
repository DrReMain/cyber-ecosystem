package audit

import (
	"context"
	"log/slog"

	kaudit "cyber-ecosystem/shared-go/kratos/audit"

	"cyber-ecosystem/app/services/agent/internal/shared"
)

// Port ----------------------------------------------------------------------------------------------------------------

type AuditRP interface {
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

func (uc *AuditUC) Emit(ctx context.Context, ev *kaudit.Event) {
	uc.auditRP.Publish(ctx, ev)
}
