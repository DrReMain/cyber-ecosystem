package cache

import (
	"context"
)

// OperationArgs holds detailed arguments for tracing cache operations.
// This is shared by both memory and redis implementations.
type OperationArgs struct {
	Key    string
	Keys   []string
	Value  []byte
	Values map[string][]byte
	// Session related
	SessionID string
	// SortedSet related
	Member  string
	Members []Member
	Score   float64
	Delta   float64
	// Counter related
	DeltaInt int64
	// RateLimiter related
	Quota  int64
	Window int64 // milliseconds
}

// Tracer defines the interface for tracing cache operations.
type Tracer interface {
	// TraceOperation traces a cache operation with the given name.
	TraceOperation(ctx context.Context, operation string, fn func(context.Context) error) error

	// TraceOperationWithArgs traces a cache operation with detailed arguments.
	TraceOperationWithArgs(ctx context.Context, operation string, args *OperationArgs, fn func(context.Context) error) error

	// TracePipeline traces multiple operations as a pipeline.
	TracePipeline(ctx context.Context, operations []string, fn func(context.Context) error) error
}

// SpanNamer defines how to format span names for different cache types.
type SpanNamer interface {
	SpanName(operation string) string
}

// DefaultSpanNamer provides consistent span naming.
type DefaultSpanNamer struct {
	Prefix string
}

func (s *DefaultSpanNamer) SpanName(operation string) string {
	return s.Prefix + " " + operation
}
