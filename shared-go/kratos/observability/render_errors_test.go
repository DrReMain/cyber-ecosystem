package observability

import (
	"context"
	stderrors "errors"
	"fmt"
	"log/slog"
	"reflect"
	"slices"
	"testing"
	"time"

	"connectrpc.com/connect"
	kratoserrors "github.com/go-kratos/kratos/v3/errors"

	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"

	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"
)

func TestViewError(t *testing.T) {
	entErr := stderrors.New(`ent: constraint violation: pq: duplicate key value violates unique constraint "dept_name_key" (sqlstate 23505)`)
	cases := []struct {
		name string
		err  error
		want errorView
	}{
		{
			// A domain rule violation: kratos-only chain, the reason is the
			// whole story and there is no factual tail.
			name: "domain rule carries no root",
			err:  errorspb.ErrorGeneralErrorInvalidArgument(""),
			want: errorView{code: 400, reason: "GENERAL_ERROR_INVALID_ARGUMENT"},
		},
		{
			name: "identity over a raw fact",
			err:  errorspb.ErrorInfraErrorDbConstraint("").WithCause(entErr),
			want: errorView{code: 409, reason: "INFRA_ERROR_DB_CONSTRAINT", root: entErr.Error()},
		},
		{
			// Upstream transposition: an identity wrapping another identity
			// wrapping the fact. The inner identity is preserved as an
			// internal classification; the root stays the fact.
			name: "double identity keeps the classification and the fact",
			err: errorspb.ErrorGeneralErrorInternal("").WithCause(
				errorspb.ErrorInfraErrorDbConstraint("").WithCause(entErr)),
			want: errorView{code: 500, reason: "GENERAL_ERROR_INTERNAL", internal: []string{"INFRA_ERROR_DB_CONSTRAINT"}, root: entErr.Error()},
		},
		{
			// Deliberate ambiguity: a fuzzy wire identity over a specific
			// internal classification. The client sees the fuzzy reason
			// (cause never crosses the wire); collection sees the specific
			// one.
			name: "fuzzy wire identity over a specific internal classification",
			err: errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(
				errorspb.ErrorGeneralErrorNotFound("")),
			want: errorView{code: 401, reason: "GENERAL_ERROR_UNAUTHENTICATED", internal: []string{"GENERAL_ERROR_NOT_FOUND"}},
		},
		{
			name: "client-facing message is part of identity",
			err:  errorspb.ErrorGeneralErrorInvalidArgument("name already taken"),
			want: errorView{code: 400, reason: "GENERAL_ERROR_INVALID_ARGUMENT", message: "name already taken"},
		},
		{
			// A prose cause contributes no internal classification — it is
			// the factual root, carried whole.
			name: "prose cause stays whole as the root",
			err: errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(
				fmt.Errorf("user not found: %s", "u123")),
			want: errorView{code: 401, reason: "GENERAL_ERROR_UNAUTHENTICATED", root: "user not found: u123"},
		},
		{
			// A %w wrapper is walked through: the root is the raw error
			// beneath, not the wrapper text.
			name: "%w wrapper above the raw fact",
			err: errorspb.ErrorGeneralErrorInvalidArgument("").WithCause(
				fmt.Errorf("password rejected: %w", stderrors.New("bcrypt: password length exceeds 72 bytes"))),
			want: errorView{code: 400, reason: "GENERAL_ERROR_INVALID_ARGUMENT", root: "bcrypt: password length exceeds 72 bytes"},
		},
		{
			name: "infra fact stays whole",
			err:  errorspb.ErrorInfraErrorDbInternal("").WithCause(stderrors.New("ERROR: relation does not exist")),
			want: errorView{code: 500, reason: "INFRA_ERROR_DB_INTERNAL", root: "ERROR: relation does not exist"},
		},
		{
			// A non-kratos error keeps no identity here; sanitize masks it on
			// the wire while telemetry classifies it as unclassified/500.
			name: "non-kratos error has no identity",
			err:  stderrors.New("boom"),
			want: errorView{root: "boom"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := viewError(c.err)
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("viewError = %+v, want %+v", got, c.want)
			}
		})
	}
}

// captured is what the captureHandler saw: level plus attributes as a map.
type captured struct {
	level slog.Level
	attrs map[string]string
	keys  []string
}

type captureHandler struct{ got *captured }

func (h *captureHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	h.got = &captured{level: r.Level, attrs: map[string]string{}}
	r.Attrs(func(a slog.Attr) bool {
		h.got.attrs[a.Key] = a.Value.String()
		h.got.keys = append(h.got.keys, a.Key)
		return true
	})
	return nil
}

func (h *captureHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *captureHandler) WithGroup(string) slog.Handler      { return h }

// emit feeds a record shaped like the stock kratos logging middleware's
// failure line through a RenderErrors logger and returns what came out.
func emit(t *testing.T, err error) *captured {
	t.Helper()
	handler := &captureHandler{}
	logger := RenderErrors(slog.New(handler))
	se := kratoserrors.FromError(err)
	r := slog.NewRecord(time.Now(), slog.LevelError, "server request", 0)
	r.AddAttrs(
		slog.String("kind", "server"),
		slog.String("component", "connect"),
		slog.String("operation", "/cyber.system.v1.DeptService/UpdateDept"),
		slog.Int64("code", int64(se.GetCode())),
		slog.String("reason", se.GetReason()),
		slog.Float64("latency", 0.001),
	)
	if err != nil {
		r.AddAttrs(slog.Any("error", err))
		r.AddAttrs(slog.String("stack", stockRendering(err)))
	}
	_ = logger.Handler().Handle(t.Context(), r)
	return handler.got
}

// stockRendering reproduces what the middleware derives for its stack
// attribute: %+v of the error, identical to the error attribute for a
// kratos *Error.
func stockRendering(err error) string {
	return err.Error()
}

func TestRenderErrorsDomainRule(t *testing.T) {
	got := emit(t, errorspb.ErrorGeneralErrorInvalidArgument(""))

	if slices.Contains(got.keys, "error") {
		t.Errorf("domain rule has no fact; error attr should be omitted, got %v", got.keys)
	}
	if slices.Contains(got.keys, "stack") {
		t.Errorf("stack duplication should be dropped, got %v", got.keys)
	}
	if got.level != slog.LevelWarn {
		t.Errorf("4xx failure should be demoted to WARN, got %v", got.level)
	}
	if got.attrs["reason"] != "GENERAL_ERROR_INVALID_ARGUMENT" {
		t.Errorf("identity must keep its dedicated field, got reason=%q", got.attrs["reason"])
	}
}

func TestRenderErrorsInfraFact(t *testing.T) {
	entErr := stderrors.New(`ent: constraint violation: pq: duplicate key value violates unique constraint "dept_name_key" (sqlstate 23505)`)
	got := emit(t, errorspb.ErrorInfraErrorDbConstraint("").WithCause(entErr))

	if got.attrs["error"] != entErr.Error() {
		t.Errorf("error attr should carry the flattened root, got %q", got.attrs["error"])
	}
	if slices.Contains(got.keys, "stack") {
		t.Errorf("stack duplication should be dropped, got %v", got.keys)
	}
	if got.level != slog.LevelWarn {
		t.Errorf("409 failure should be demoted to WARN, got %v", got.level)
	}
	if slices.Contains(got.keys, "message") {
		t.Errorf("empty message should not surface an attribute, got %v", got.keys)
	}
}

func TestRenderErrorsServerFault(t *testing.T) {
	got := emit(t, errorspb.ErrorGeneralErrorInternal("").WithCause(stderrors.New("connection refused")))

	if got.level != slog.LevelError {
		t.Errorf("5xx failure keeps ERROR level, got %v", got.level)
	}
	if got.attrs["error"] != "connection refused" {
		t.Errorf("error attr should carry the root fact, got %q", got.attrs["error"])
	}
}

func TestRenderErrorsInternalClassification(t *testing.T) {
	// Fuzzy wire identity over a specific internal one: the log line must
	// expose the internal path (collection needs the specific cause) while
	// the wire keeps only the fuzzy reason.
	got := emit(t, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(
		errorspb.ErrorGeneralErrorNotFound("")))

	if got.attrs["internal"] != "GENERAL_ERROR_NOT_FOUND" {
		t.Errorf("internal classification must be queryable, got %q", got.attrs["internal"])
	}
	if got.attrs["reason"] != "GENERAL_ERROR_UNAUTHENTICATED" {
		t.Errorf("wire identity stays fuzzy, got %q", got.attrs["reason"])
	}
	if slices.Contains(got.keys, "error") {
		t.Errorf("classification-only chain has no factual root, got %v", got.keys)
	}
}

func TestRenderErrorsClientMessage(t *testing.T) {
	got := emit(t, errorspb.ErrorGeneralErrorInvalidArgument("name already taken"))

	if got.attrs["message"] != "name already taken" {
		t.Errorf("client-facing message must be queryable in logs, got %q", got.attrs["message"])
	}
}

func TestRenderErrorsNonKratosPassThrough(t *testing.T) {
	got := emit(t, stderrors.New("boom"))

	if got.attrs["error"] != "boom" || !slices.Contains(got.keys, "stack") {
		t.Errorf("non-kratos records pass through unchanged, got attrs=%v", got.attrs)
	}
	if got.level != slog.LevelError {
		t.Errorf("non-kratos records keep their level, got %v", got.level)
	}
}

func TestErrorViewFlat(t *testing.T) {
	cases := []struct {
		name string
		view errorView
		want string
	}{
		{"reason only", errorView{reason: "SYSTEM_DEPT_SELF_PARENT"}, "SYSTEM_DEPT_SELF_PARENT"},
		{"reason over root", errorView{reason: "INFRA_ERROR_DB_NOT_FOUND", root: "ent: dept not found"}, "INFRA_ERROR_DB_NOT_FOUND: ent: dept not found"},
		{
			"full house",
			errorView{reason: "GENERAL_ERROR_UNAUTHENTICATED", internal: []string{"INFRA_ERROR_DB_NOT_FOUND"}, root: "ent: user not found"},
			"GENERAL_ERROR_UNAUTHENTICATED [INFRA_ERROR_DB_NOT_FOUND]: ent: user not found",
		},
		{
			"client message included",
			errorView{reason: "GENERAL_ERROR_INVALID_ARGUMENT", message: "name already taken"},
			"GENERAL_ERROR_INVALID_ARGUMENT (name already taken)",
		},
		{"non-kratos keeps only the fact", errorView{root: "boom"}, "boom"},
		{"empty never renders blank", errorView{}, "unclassified error"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.view.flat(); got != c.want {
				t.Fatalf("flat = %q, want %q", got, c.want)
			}
		})
	}
}

// TestRenderFlatIdentityPassthrough pins the contract that makes the outbound
// re-rendering safe: every identity query resolves to the ORIGINAL chain, so
// the wire translators, sanitize, and viewError all behave exactly as they do
// on the unwrapped error.
func TestRenderFlatIdentityPassthrough(t *testing.T) {
	entErr := stderrors.New("ent: dept not found")
	orig := errorspb.ErrorInfraErrorDbConstraint("").WithCause(entErr)
	wrapped := renderFlat(orig)

	t.Run("Error is the flat view", func(t *testing.T) {
		want := "INFRA_ERROR_DB_CONSTRAINT: ent: dept not found"
		if wrapped.Error() != want {
			t.Fatalf("Error() = %q, want %q", wrapped.Error(), want)
		}
	})

	t.Run("kratos FromError still resolves the original identity", func(t *testing.T) {
		ke := kratoserrors.FromError(wrapped)
		if ke.Reason != "INFRA_ERROR_DB_CONSTRAINT" || ke.Code != 409 {
			t.Fatalf("FromError = reason %q code %d, want INFRA_ERROR_DB_CONSTRAINT/409", ke.Reason, ke.Code)
		}
	})

	t.Run("sanitize-style As still sees a kratos error", func(t *testing.T) {
		var ke *kratoserrors.Error
		if !stderrors.As(wrapped, &ke) {
			t.Fatal("errors.As must reach the original kratos error through the wrapper")
		}
	})

	t.Run("viewError walks through to the same projection", func(t *testing.T) {
		want := viewError(orig)
		if got := viewError(wrapped); !reflect.DeepEqual(got, want) {
			t.Fatalf("viewError(wrapped) = %+v, want %+v", got, want)
		}
	})

	t.Run("idempotent wrap", func(t *testing.T) {
		//nolint:errorlint // pointer identity is exactly the assertion: re-wrapping must not nest
		if again := renderFlat(wrapped); again != wrapped {
			t.Fatal("re-wrapping must return the same error, not nest")
		}
	})

	t.Run("wire translation is unaffected", func(t *testing.T) {
		// The real connect exit translator: same code, and the same wire
		// text (empty message because nothing client-facing was authored).
		wireOrig := connecttransport.ErrorToConnect(orig)
		wireWrapped := connecttransport.ErrorToConnect(wrapped)
		if got, want := connect.CodeOf(wireWrapped), connect.CodeOf(wireOrig); got != want {
			t.Fatalf("connect code changed: wrapped=%v orig=%v", got, want)
		}
		if wireWrapped.Error() != wireOrig.Error() {
			t.Fatalf("connect error text changed: wrapped=%q orig=%q", wireWrapped.Error(), wireOrig.Error())
		}
	})
}
