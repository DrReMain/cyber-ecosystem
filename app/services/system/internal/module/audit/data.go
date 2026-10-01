package audit

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/rs/xid"

	"cyber-ecosystem/shared-go/capability/mq"
	"cyber-ecosystem/shared-go/helper"
	kaudit "cyber-ecosystem/shared-go/kratos/audit"
	"cyber-ecosystem/shared-go/utils"

	"cyber-ecosystem/app/services/system/internal/ent"
	"cyber-ecosystem/app/services/system/internal/ent/auditexport"
	"cyber-ecosystem/app/services/system/internal/ent/auditlog"
	entfile "cyber-ecosystem/app/services/system/internal/ent/file"
	"cyber-ecosystem/app/services/system/internal/ent/predicate"
	"cyber-ecosystem/app/services/system/internal/module/file"
	"cyber-ecosystem/app/services/system/internal/platform"
	"cyber-ecosystem/app/services/system/internal/shared"
)

const (
	topic         = "audit_log"
	consumerGroup = "audit" // group identity; survives restarts — only a brand-new group replays
)

// Repo ----------------------------------------------------------------------------------------------------------------

type auditRP struct {
	shared.RP
	emitter *kaudit.Emitter
}

func NewAuditRP(logger *slog.Logger, p *platform.Platform) (AuditRP, func()) {
	log := logger.With("module", "module/audit")
	rp := &auditRP{RP: shared.NewRP(log, p)}
	emitter, stop := kaudit.NewEmitter(rp.publish, log)
	rp.emitter = emitter
	// Subscribe failure degrades to no persistence with a loud log — the
	// service stays up; audit availability yields to service availability.
	sub, err := p.GetMQ().Consumer.Subscribe(context.Background(), topic, consumerGroup, rp.handle)
	if err != nil {
		log.Error("audit consumer subscribe failed; events are not being persisted", "topic", topic, "error", err)
		return rp, stop
	}
	return rp, func() { _ = sub.Close(); stop() }
}

// Method --------------------------------------------------------------------------------------------------------------

func (rp *auditRP) ListAuditLogs(ctx context.Context, in *AuditListIn) (*AuditListOut, error) {
	query := rp.Platform.GetClient(ctx).AuditLog.Query()
	helper.Where(query, in.Tenant != "", func() predicate.AuditLog { return auditlog.TenantIn(in.Tenant, "") })
	helper.Where(query, in.Actor != nil, func() predicate.AuditLog { return auditlog.ActorEQ(*in.Actor) })
	helper.Where(query, in.Operation != nil, func() predicate.AuditLog { return auditlog.OperationContainsFold(*in.Operation) })
	helper.Where(query, in.Status != nil, func() predicate.AuditLog {
		return auditlog.StatusEQ(utils.ConvNum[int](*in.Status))
	})
	helper.Where(query, in.DenyReason != nil, func() predicate.AuditLog { return auditlog.DenyReasonEQ(*in.DenyReason) })
	helper.Where(query, in.Denied != nil, func() predicate.AuditLog {
		if *in.Denied {
			return auditlog.DenyReasonNEQ("")
		}
		return auditlog.DenyReasonEQ("")
	})
	helper.WherePtr(query, utils.FromTimestamp(in.CreatedAtA), auditlog.CreatedAtGTE)
	helper.WherePtr(query, utils.FromTimestamp(in.CreatedAtZ), auditlog.CreatedAtLTE)
	helper.ApplyOrderBy(helper.ParseOrderBy(in.OrderBy), ent.Asc, ent.Desc, helper.FOMapping{
		"createdAt": func(sel helper.SQLSelector) { query.Order(sel(auditlog.FieldCreatedAt)) },
	})
	total, offset, limit, err := shared.Paginate(ctx, query, in.PageRequest, helper.DefaultPageSizeMax)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	logs, err := query.All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return &AuditListOut{
		PageResponse: helper.BuildPageResponse(total, offset, limit),
		List:         utils.SliceMap(logs, mapAuditLog),
	}, nil
}

func (rp *auditRP) Publish(ctx context.Context, ev *kaudit.Event) {
	rp.emitter.Emit(ctx, ev)
}

func (rp *auditRP) CreateExport(ctx context.Context, fileID, ownerID string) error {
	if err := rp.Platform.GetClient(ctx).AuditExport.Create().
		SetFileID(fileID).
		SetOwnerID(ownerID).
		Exec(ctx); err != nil {
		return rp.Platform.HandleEntError(err)
	}
	return nil
}

func (rp *auditRP) ListExports(ctx context.Context, in *ExportListIn) (*ExportListOut, error) {
	query := rp.Platform.GetClient(ctx).AuditExport.Query()
	helper.ApplyOrderBy(helper.ParseOrderBy(in.OrderBy), ent.Asc, ent.Desc, helper.FOMapping{
		"createdAt": func(sel helper.SQLSelector) { query.Order(sel(auditexport.FieldCreatedAt)) },
	})
	total, offset, limit, err := shared.Paginate(ctx, query, in.PageRequest, helper.DefaultPageSizeUnlimit)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	rows, err := query.All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	// Hydration is a plain management-plane read: a retired file renders as
	// absent. Datascope on both tables is owner-aligned by construction —
	// the export row's owner is the file's owner.
	ids := make([]string, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.FileID)
	}
	files, err := rp.Platform.GetClient(ctx).File.Query().Where(entfile.IDIn(ids...)).All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	byID := make(map[string]*file.File, len(files))
	for _, d := range files {
		byID[d.ID] = file.MapFile(d)
	}
	list := make([]*AuditExport, 0, len(rows))
	for _, r := range rows {
		list = append(list, &AuditExport{
			ID:        r.ID,
			CreatedAt: r.CreatedAt,
			FileID:    r.FileID,
			OwnerID:   r.OwnerID,
			File:      byID[r.FileID],
		})
	}
	return &ExportListOut{
		PageResponse: helper.BuildPageResponse(total, offset, limit),
		List:         list,
	}, nil
}

func (rp *auditRP) ListLogsAfter(ctx context.Context, tenant, after string, limit int) ([]*AuditLog, error) {
	// Keyset iteration: id order is total and stable, so cursor resumption
	// never skips or repeats a row.
	query := rp.Platform.GetClient(ctx).AuditLog.Query().
		Where(auditlog.TenantIn(tenant, "")).
		Order(ent.Asc(auditlog.FieldID)).
		Limit(limit)
	if after != "" {
		query.Where(auditlog.IDGT(after))
	}
	logs, err := query.All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return utils.SliceMap(logs, mapAuditLog), nil
}

func (rp *auditRP) FindLogsByIDs(ctx context.Context, tenant string, ids []string) ([]*AuditLog, error) {
	// Unknown ids simply export fewer rows, mirroring the skip semantics
	// of a missing-key map read.
	logs, err := rp.Platform.GetClient(ctx).AuditLog.Query().
		Where(auditlog.TenantIn(tenant, ""), auditlog.IDIn(ids...)).
		Order(ent.Asc(auditlog.FieldID)).
		All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return utils.SliceMap(logs, mapAuditLog), nil
}

// Private -------------------------------------------------------------------------------------------------------------

func (rp *auditRP) publish(ctx context.Context, ev *kaudit.Event) error {
	payload, err := utils.Marshal(ev)
	if err != nil {
		return err
	}
	msg := &mq.Message{
		Topic:     topic,
		Payload:   payload,
		Headers:   map[string]string{"event_id": xid.New().String()},
		Timestamp: ev.OccurredAt,
	}
	_, err = rp.Platform.GetMQ().Publisher.Publish(ctx, topic, msg)
	return err
}

func (rp *auditRP) handle(ctx context.Context, msg mq.Message) error {
	// At-least-once consumer: errors nak (retry, then DLQ after the
	// backend's cap). Redelivery dedup is dialect-free: the common path is
	// a replay whose row already exists (the Exist pre-check acks
	// directly); the race of two concurrent deliveries both inserting falls
	// into the event_id unique index — the only constraint an insert here
	// can violate (every other column has a default) — so
	// ent.IsConstraintError reads as "already persisted". No pg driver
	// types cross into the orm or business layers.
	eventID := msg.Headers["event_id"]
	if eventID == "" {
		// A contract violation no retry can fix; DLQ keeps it visible.
		return fmt.Errorf("audit message without event_id header")
	}
	ev, err := utils.Unmarshal[kaudit.Event](msg.Payload)
	if err != nil {
		return fmt.Errorf("decode audit event: %w", err)
	}
	// Non-ent failures above stay raw so the DLQ row names the cause; ent
	// errors go through the choke point (the original stays in the cause
	// chain, and the consumer retries on any non-nil error either way).
	// created_at carries the event's true time so replays don't skew.
	client := rp.Platform.GetClient(ctx)
	exists, err := client.AuditLog.Query().Where(auditlog.EventIDEQ(eventID)).Exist(ctx)
	if err != nil {
		return rp.Platform.HandleEntError(fmt.Errorf("audit exist check: %w", err))
	}
	if exists {
		return nil
	}
	_, err = client.AuditLog.Create().
		SetEventID(eventID).
		SetTenant(ev.TenantID).
		SetActor(ev.Actor).
		SetPrincipalType(ev.PrincipalType).
		SetOperation(ev.Operation).
		SetHTTPMethod(ev.HTTPMethod).
		SetHTTPPath(ev.HTTPPath).
		SetStatus(ev.Status).
		SetLatencyMs(utils.ConvNum[int](ev.LatencyMs)).
		SetIP(ev.IP).
		SetUserAgent(ev.UserAgent).
		SetDenyReason(ev.DenyReason).
		SetCreatedAt(ev.OccurredAt).
		Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil
		}
		return rp.Platform.HandleEntError(fmt.Errorf("insert audit log: %w", err))
	}
	return nil
}

func mapAuditLog(d *ent.AuditLog) *AuditLog {
	return &AuditLog{
		ID:            d.ID,
		CreatedAt:     d.CreatedAt,
		Tenant:        d.Tenant,
		Actor:         d.Actor,
		PrincipalType: d.PrincipalType,
		Operation:     d.Operation,
		HTTPMethod:    d.HTTPMethod,
		HTTPPath:      d.HTTPPath,
		Status:        d.Status,
		LatencyMs:     d.LatencyMs,
		IP:            d.IP,
		UserAgent:     d.UserAgent,
		DenyReason:    d.DenyReason,
	}
}
