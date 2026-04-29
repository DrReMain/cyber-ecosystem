package error_report

import (
	"context"
	"fmt"
	"runtime/debug"
	"strconv"

	sentry "github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"

	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"
)

// Server returns a middleware that reports errors to GlitchTip/Sentry
// and records them as OTel span events.
// It captures panics directly to preserve the original value and stack trace,
// then re-panics so recovery middleware can still handle the response.
func Server() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (reply any, err error) {
			defer func() {
				if r := recover(); r != nil {
					reportPanic(ctx, r)
					panic(r)
				}
			}()

			reply, err = handler(ctx, req)

			recordSpanStatusCode(ctx, err)

			if err == nil {
				return
			}

			recordSpanError(ctx, err)

			if isReportable(err) {
				reportToSentry(ctx, err)
			}

			return
		}
	}
}

func reportPanic(ctx context.Context, v any) {
	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub()
	}

	panicStack := string(debug.Stack())

	hub.WithScope(func(scope *sentry.Scope) {
		scope.SetTag("panic", "true")
		if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
			scope.SetTag("trace_id", span.SpanContext().TraceID().String())
		}

		event := &sentry.Event{
			Level: sentry.LevelFatal,
			Contexts: map[string]sentry.Context{
				"panic": {
					"stack": panicStack,
				},
			},
			Exception: []sentry.Exception{
				{
					Type:       panicTypeName(v),
					Value:      fmt.Sprintf("%v", v),
					Stacktrace: sentry.NewStacktrace(),
				},
			},
		}
		hub.CaptureEvent(event)
	})
	hub.Flush(sentry.DefaultFlushTimeout)
}

func panicTypeName(v any) string {
	if _, ok := v.(error); ok {
		return fmt.Sprintf("%T", v)
	}
	return "panic"
}

func isReportable(err error) bool {
	if se := errors.FromError(err); se != nil {
		return se.Code >= 500
	}
	return true
}

func recordSpanError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	reason := ""
	statusCode := 500
	if se := errors.FromError(err); se != nil {
		reason = se.Reason
		statusCode = int(se.Code)
	}

	span.SetStatus(codes.Error, reason)
	span.SetAttributes(
		attribute.String("error.reason", reason),
		attribute.Int("error.status_code", statusCode),
	)
	span.RecordError(err)
}

func reportToSentry(ctx context.Context, err error) {
	sentry.WithScope(func(scope *sentry.Scope) {
		if se := errors.FromError(err); se != nil {
			scope.SetTag("error.reason", se.Reason)
			scope.SetTag("error.status_code", strconv.Itoa(int(se.Code)))
			scope.SetContext("error", sentry.Context{
				"message": se.Message,
			})
			scope.SetFingerprint([]string{se.Reason})

			if cause := errors.Unwrap(se); cause != nil {
				scope.SetContext("error", sentry.Context{
					"message": se.Message,
					"cause":   cause.Error(),
				})
			}
		}

		if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
			scope.SetTag("trace_id", span.SpanContext().TraceID().String())
		}

		sentry.CaptureException(err)
	})
}

func recordSpanStatusCode(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	tr, ok := transport.FromServerContext(ctx)
	if !ok {
		return
	}

	var statusCode int
	if err != nil {
		if se := errors.FromError(err); se != nil {
			statusCode = int(se.Code)
		} else {
			statusCode = 500
		}
	} else {
		statusCode = 200
	}

	switch tr.Kind() {
	case transport.KindHTTP, connecttransport.KindConnect:
		span.SetAttributes(attribute.Int("http.status_code", statusCode))
	case transport.KindGRPC:
		grpcCode := statusCode
		if grpcCode >= 100 {
			grpcCode = mapHTTPToGRPC(grpcCode)
		}
		span.SetAttributes(attribute.Int("rpc.grpc.status_code", grpcCode))
	}
}

func mapHTTPToGRPC(httpCode int) int {
	switch {
	case httpCode == 400:
		return 3
	case httpCode == 401:
		return 16
	case httpCode == 403:
		return 7
	case httpCode == 404:
		return 5
	case httpCode == 409:
		return 6
	case httpCode == 429:
		return 8
	case httpCode == 501:
		return 12
	case httpCode == 503:
		return 14
	case httpCode == 504:
		return 4
	default:
		return 2
	}
}
