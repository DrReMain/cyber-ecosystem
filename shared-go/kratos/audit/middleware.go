package audit

import (
	"context"
	"net/http"
	"strings"
	"time"

	kerrors "github.com/go-kratos/kratos/v3/errors"
	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/transport"
	ktranshttp "github.com/go-kratos/kratos/v3/transport/http"

	"cyber-ecosystem/shared-go/kratos/security"
	"cyber-ecosystem/shared-go/kratos/transport/connect"

	extv1 "cyber-ecosystem/gen/go/cyber/ext/v1"
)

// Audit emits one event per collected operation outcome. Collection
// semantics: allow-side WRITE actions and every authorization denial — a
// read that passes is not audit material, a denied read always is (probe
// detection is the point).
//
// Position contract: inside the ADMIN audience block it must sit between
// SessionAuth and OperationAuthz — after the former so subject attribution
// is in the context, outside the latter so denial errors return through it.
// A 401 from the outer SessionAuth never reaches this middleware. Should an
// unauthenticated audience ever carry it, subject-less events keep the
// attribution columns empty.
func Audit(sink Sink) middleware.Middleware {
	if sink == nil {
		// Wiring breakage: an audit middleware with nowhere to send events
		// should fail loudly at assembly, not silently collect nothing.
		panic("Audit: sink is required")
	}
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			start := time.Now()
			reply, err := handler(ctx, req)
			deny := denyReasonOf(err)
			if deny == "" && actionOf(operationOf(ctx)) != extv1.Action_ACTION_WRITE {
				return reply, err
			}
			sink.Emit(ctx, newEvent(ctx, start, err, deny))
			return reply, err
		}
	}
}

// operationOf returns the RPC operation of the inbound server transport, or
// "" outside one (classified then by the WRITE floor — fail-safe).
func operationOf(ctx context.Context) string {
	if tr, ok := transport.FromServerContext(ctx); ok {
		return tr.Operation()
	}
	return ""
}

func newEvent(ctx context.Context, start time.Time, err error, deny string) *Event {
	ev := &Event{
		OccurredAt: start,
		Status:     statusOf(err),
		DenyReason: deny,
		LatencyMs:  time.Since(start).Milliseconds(),
	}
	if s, ok := security.SubjectFromCtx(ctx); ok {
		ev.TenantID = s.TenantID
		ev.Actor = s.UserID
		ev.PrincipalType = "user"
	}
	if tr, ok := transport.FromServerContext(ctx); ok {
		ev.Operation = tr.Operation()
		ev.UserAgent = tr.RequestHeader().Get("User-Agent")
		ev.IP = clientIP(tr.RequestHeader())
		switch tr.Kind() {
		case transport.KindHTTP:
			if ht, ok := tr.(*ktranshttp.Transport); ok && ht.Request() != nil && ht.Request().URL != nil {
				ev.HTTPMethod = ht.Request().Method
				ev.HTTPPath = ht.Request().URL.Path
			}
		case connect.KindConnect:
			ev.HTTPMethod = http.MethodPost
			ev.HTTPPath = tr.Operation()
		}
	}
	return ev
}

// clientIP reads the client address from the standard forwarding headers;
// the edge sets X-Forwarded-For and the first hop is the client. Direct
// connections without a proxy in front leave the field empty.
func clientIP(h transport.Header) string {
	if v := h.Get("X-Forwarded-For"); v != "" {
		if i := strings.IndexByte(v, ','); i >= 0 {
			v = v[:i]
		}
		return strings.TrimSpace(v)
	}
	return h.Get("X-Real-Ip")
}

// statusOf maps the outcome to its HTTP status; kratos errors carry it as
// their code, anything else is an internal failure.
func statusOf(err error) int {
	if err == nil {
		return http.StatusOK
	}
	if e := kerrors.FromError(err); e != nil && e.Code > 0 {
		return int(e.Code)
	}
	return http.StatusInternalServerError
}
