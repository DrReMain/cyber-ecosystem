package audit

import (
	"context"
	"log/slog"

	"github.com/rs/xid"

	"cyber-ecosystem/shared-go/capability/mq"
	kaudit "cyber-ecosystem/shared-go/kratos/audit"
	"cyber-ecosystem/shared-go/utils"

	"cyber-ecosystem/app/services/esports/internal/platform"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

// Repo ----------------------------------------------------------------------------------------------------------------

type auditRP struct {
	shared.RP
	emitter *kaudit.Emitter
}

func NewAuditRP(logger *slog.Logger, p *platform.Platform) (AuditRP, func()) {
	// The producing half only: this service publishes to the shared audit
	// topic; the system service owns the subscribing consumer. The cleanup
	// stops the emitter's drainer after a best-effort flush of queued events.
	rp := &auditRP{RP: shared.NewRP(logger.With("module", "module/audit_rp"), p)}
	emitter, stop := kaudit.NewEmitter(rp.publish, rp.Log)
	rp.emitter = emitter
	return rp, stop
}

// Method --------------------------------------------------------------------------------------------------------------

func (rp *auditRP) Publish(ctx context.Context, ev *kaudit.Event) {
	rp.emitter.Emit(ctx, ev)
}

// Private -------------------------------------------------------------------------------------------------------------

func (rp *auditRP) publish(ctx context.Context, ev *kaudit.Event) error {
	payload, err := utils.Marshal(ev)
	if err != nil {
		return err
	}
	msg := &mq.Message{
		Topic:     kaudit.TopicLog,
		Payload:   payload,
		Headers:   map[string]string{"event_id": xid.New().String()},
		Timestamp: ev.OccurredAt,
	}
	if _, err = rp.Platform.GetMQ().Publisher.Publish(ctx, kaudit.TopicLog, msg); err != nil {
		return rp.Platform.HandleMQError(err)
	}
	return nil
}
