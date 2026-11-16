package audit

import (
	"context"
	"io"
	"log/slog"
	"testing"

	kaudit "cyber-ecosystem/shared-go/kratos/audit"
)

// fakeAuditRP records the sink handoff; the insert itself is the
// database-only platform's live arm (e2e owns it).
type fakeAuditRP struct {
	AuditRP
	published []*kaudit.Event
}

func (f *fakeAuditRP) Publish(_ context.Context, ev *kaudit.Event) {
	f.published = append(f.published, ev)
}

// Emit is the ladder's sink face: whatever the middleware hands the UC
// reaches the emitter queue verbatim — added nothing, dropped nothing.
func TestEmitPublishesVerbatim(t *testing.T) {
	rp := &fakeAuditRP{}
	uc := NewAuditUC(slog.New(slog.NewTextHandler(io.Discard, nil)), nil, rp)
	ev := &kaudit.Event{TenantID: "t1", Operation: "agent e2e op"}

	uc.Emit(context.Background(), ev)
	if len(rp.published) != 1 || rp.published[0] != ev {
		t.Fatalf("publish: %+v", rp.published)
	}
}
