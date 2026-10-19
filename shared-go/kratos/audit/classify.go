package audit

import (
	"context"
	"log/slog"
	"strings"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	extv1 "cyber-ecosystem/gen/go/cyber/ext/v1"
)

var actionCache sync.Map // operation string → extv1.Action

// actionOf resolves the effect semantics of an operation through a ladder:
// the desc action annotation (authoritative) → the method-name vocabulary
// (fallback) → the WRITE floor. The floor direction is deliberate — a
// missed read is invisible, an over-collected write is visible and fixable
// — and its hits are counted as a guard-loss signal. Results are cached per
// operation.
func actionOf(operation string) extv1.Action {
	if v, ok := actionCache.Load(operation); ok {
		return v.(extv1.Action)
	}
	act := resolveAction(operation)
	actionCache.Store(operation, act)
	return act
}

func resolveAction(operation string) extv1.Action {
	// "/pkg.Svc/Method" → "pkg.Svc.Method" (protoreflect.FullName)
	name := strings.ReplaceAll(strings.TrimPrefix(operation, "/"), "/", ".")
	if desc, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(name)); err == nil {
		if md, ok := desc.(protoreflect.MethodDescriptor); ok {
			if d := methodDescOf(md); d != nil && d.GetAction() != extv1.Action_ACTION_UNSPECIFIED {
				return d.GetAction()
			}
			if readVerb(string(md.Name())) {
				return extv1.Action_ACTION_READ
			}
		}
	}
	countFloorHit(operation)
	return extv1.Action_ACTION_WRITE
}

func methodDescOf(md protoreflect.MethodDescriptor) *extv1.MethodDesc {
	opts := md.Options()
	if opts == nil || !proto.HasExtension(opts, extv1.E_Method) {
		return nil
	}
	d, _ := proto.GetExtension(opts, extv1.E_Method).(*extv1.MethodDesc)
	return d
}

// readVerb is the fallback vocabulary: the read-shaped verb prefixes of the
// repo's RPC naming. Everything else reads as a write.
func readVerb(method string) bool {
	return strings.HasPrefix(method, "List") || strings.HasPrefix(method, "Get")
}

// The floor counter is created lazily on the first hit so it binds to the
// meter provider installed at app startup, not the pre-init noop one.
var (
	floorOnce    sync.Once
	floorCounter metric.Int64Counter
)

func countFloorHit(operation string) {
	floorOnce.Do(func() {
		c, err := otel.Meter("kratos").Int64Counter("kratos.audit.write_floor")
		if err != nil {
			slog.Warn("audit: write-floor counter creation failed", "error", err)
			return
		}
		floorCounter = c
	})
	if floorCounter != nil {
		floorCounter.Add(context.Background(), 1, metric.WithAttributes(attribute.String("operation", operation)))
	}
}
