package audit

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strconv"
	"time"

	kaudit "cyber-ecosystem/shared-go/kratos/audit"
	"cyber-ecosystem/shared-go/kratos/security"

	commonpb "cyber-ecosystem/gen/go/cyber/shared/common/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"

	"cyber-ecosystem/app/services/system/internal/module/file"
	"cyber-ecosystem/app/services/system/internal/shared"
)

const auditExportBatch = 500

var auditCSVHeader = []string{
	"id", "created_at", "tenant", "actor", "principal_type", "operation",
	"http_method", "http_path", "status", "latency_ms", "ip", "user_agent", "deny_reason",
}

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

type AuditExport struct {
	ID        string
	CreatedAt time.Time
	FileID    string
	OwnerID   string
	File      *file.File
}

type ExportIn struct {
	IDs []string
}

type ExportListIn struct {
	*commonpb.PageRequest
	OrderBy []string
}

type ExportListOut struct {
	*commonpb.PageResponse
	List []*AuditExport
}

// Port ----------------------------------------------------------------------------------------------------------------

type AuditRP interface {
	ListAuditLogs(ctx context.Context, in *AuditListIn) (*AuditListOut, error)
	Publish(ctx context.Context, ev *kaudit.Event)
	CreateExport(ctx context.Context, fileID, ownerID string) error
	ListExports(ctx context.Context, in *ExportListIn) (*ExportListOut, error)
	ListLogsAfter(ctx context.Context, tenant, after string, limit int) ([]*AuditLog, error)
	FindLogsByIDs(ctx context.Context, tenant string, ids []string) ([]*AuditLog, error)
}

// UC ------------------------------------------------------------------------------------------------------------------

type AuditUC struct {
	shared.UC
	auditRP AuditRP
	fileUC  *file.FileUC
}

func NewAuditUC(logger *slog.Logger, tm shared.Transaction, auditRP AuditRP, fileUC *file.FileUC) *AuditUC {
	return &AuditUC{
		UC:      shared.NewUC(logger.With("module", "module/audit"), tm),
		auditRP: auditRP,
		fileUC:  fileUC,
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

func (uc *AuditUC) Export(ctx context.Context, in *ExportIn) (*file.File, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("audit export requires a subject"))
	}
	spec := &file.GenerationSpec{
		Name:        fmt.Sprintf("audit-logs-%s.csv", time.Now().UTC().Format("20060102T150405Z")),
		ContentType: "text/csv",
		Produce: func(gctx context.Context, w io.Writer) error {
			return uc.writeAuditCSV(gctx, w, subject.TenantID, in.IDs)
		},
	}
	return uc.fileUC.Generate(ctx, spec, func(tctx context.Context, f *file.File) error {
		return uc.auditRP.CreateExport(tctx, f.ID, subject.UserID)
	})
}

func (uc *AuditUC) ListExports(ctx context.Context, in *ExportListIn) (*ExportListOut, error) {
	return uc.auditRP.ListExports(ctx, in)
}

func (uc *AuditUC) Emit(ctx context.Context, ev *kaudit.Event) {
	uc.auditRP.Publish(ctx, ev)
}

// Private -------------------------------------------------------------------------------------------------------------

func (uc *AuditUC) writeAuditCSV(ctx context.Context, w io.Writer, tenant string, ids []string) error {
	// Batching keeps memory O(batch) whatever the row count; the tenant
	// rule mirrors List — the caller's tenant plus the subject-less
	// public events.
	cw := csv.NewWriter(w)
	if err := cw.Write(auditCSVHeader); err != nil {
		return err
	}
	emit := func(logs []*AuditLog) error {
		for _, l := range logs {
			rec := []string{
				l.ID, l.CreatedAt.UTC().Format(time.RFC3339Nano), l.Tenant, l.Actor,
				l.PrincipalType, l.Operation, l.HTTPMethod, l.HTTPPath,
				strconv.Itoa(l.Status), strconv.Itoa(l.LatencyMs),
				l.IP, l.UserAgent, l.DenyReason,
			}
			if err := cw.Write(rec); err != nil {
				return err
			}
		}
		return nil
	}
	var err error
	if len(ids) == 0 {
		after := ""
		for {
			var batch []*AuditLog
			if batch, err = uc.auditRP.ListLogsAfter(ctx, tenant, after, auditExportBatch); err != nil || len(batch) == 0 {
				break
			}
			if err = emit(batch); err != nil {
				break
			}
			if len(batch) < auditExportBatch {
				break
			}
			after = batch[len(batch)-1].ID
		}
	} else {
		for chunk := range slices.Chunk(ids, auditExportBatch) {
			var rows []*AuditLog
			if rows, err = uc.auditRP.FindLogsByIDs(ctx, tenant, chunk); err != nil {
				break
			}
			if err = emit(rows); err != nil {
				break
			}
		}
	}
	if err != nil {
		return err
	}
	cw.Flush()
	return cw.Error()
}
