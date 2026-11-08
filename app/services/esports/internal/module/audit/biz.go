package audit

import (
	"context"
	"log/slog"

	kaudit "cyber-ecosystem/shared-go/kratos/audit"

	"cyber-ecosystem/app/services/esports/internal/shared"
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
	// Satisfies kaudit.Sink for the server ladder; this module is the
	// producing half only — the read faces (list, export) live with the
	// persisting consumer in the system service.
	uc.auditRP.Publish(ctx, ev)
}
