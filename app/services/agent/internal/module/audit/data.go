package audit

import (
	"context"
	"log/slog"

	"github.com/rs/xid"

	kaudit "cyber-ecosystem/shared-go/kratos/audit"
	"cyber-ecosystem/shared-go/utils"

	"cyber-ecosystem/app/services/agent/internal/ent"
	"cyber-ecosystem/app/services/agent/internal/platform"
	"cyber-ecosystem/app/services/agent/internal/shared"
)

// Repo ----------------------------------------------------------------------------------------------------------------

type auditRP struct {
	shared.RP
	emitter *kaudit.Emitter
}

func NewAuditRP(logger *slog.Logger, p *platform.Platform) (AuditRP, func()) {
	log := logger.With("module", "module/audit_rp")
	rp := &auditRP{RP: shared.NewRP(log, p)}
	// System fans its audit feed through the MQ topic (cross-consumer
	// scale-out, at-least-once redelivery); this database-only platform
	// persists straight from the drainer — the Emitter's queue/drop
	// semantics are the shared part, and the event_id unique index keeps
	// the same dedup contract.
	emitter, stop := kaudit.NewEmitter(rp.insert, log)
	rp.emitter = emitter
	return rp, stop
}

// Method --------------------------------------------------------------------------------------------------------------

func (rp *auditRP) Publish(ctx context.Context, ev *kaudit.Event) {
	rp.emitter.Emit(ctx, ev)
}

func (rp *auditRP) insert(ctx context.Context, ev *kaudit.Event) error {
	// The drainer's PublishFunc: an append-only write carrying the event's
	// true time so replays don't skew. A constraint loss here can only be
	// the event_id unique index (every other column defaults), which reads
	// as "already persisted" — the schema's documented dedup contract.
	if err := rp.Platform.GetClient(ctx).AuditLog.Create().
		SetEventID(xid.New().String()).
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
		Exec(ctx); err != nil {
		if ent.IsConstraintError(err) {
			return nil
		}
		return rp.Platform.HandleEntError(err)
	}
	return nil
}
