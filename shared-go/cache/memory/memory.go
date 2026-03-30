package memory

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/shared-go/cache"
	"cyber-ecosystem/shared-go/cache/memory/otel"
)

// Config holds Memory client configuration.
type Config struct {
	EnableTracing      bool
	EnableLogging      bool
	Logger             log.Logger
	LogLevel           string
	SlowQuery          bool
	SlowQueryThreshold time.Duration
}

type entry struct {
	value     []byte
	expiresAt time.Time
}

func (e *entry) isExpired() bool {
	return !e.expiresAt.IsZero() && time.Now().After(e.expiresAt)
}

type tracer interface {
	TraceOperation(ctx context.Context, op string, fn func(context.Context) error) error
	TraceOperationWithArgs(ctx context.Context, op string, args *cache.OperationArgs, fn func(context.Context) error) error
}

type Memory struct {
	mu            sync.RWMutex
	data          map[string]*entry
	logger        log.Logger
	logLevel      log.Level
	slowQuery     bool
	slowThreshold time.Duration
	otel          tracer
}

func NewMemoryClient(cfg *Config) *cache.Cache {
	if cfg == nil {
		cfg = &Config{}
	}

	var logger log.Logger
	if cfg.EnableLogging {
		logger = cfg.Logger
	}

	m := &Memory{
		data:          make(map[string]*entry),
		logger:        logger,
		logLevel:      parseLogLevel(cfg.LogLevel),
		slowQuery:     cfg.SlowQuery,
		slowThreshold: cfg.SlowQueryThreshold,
	}

	if cfg.EnableTracing {
		m.otel = otel.NewTracingHook()
	}

	return &cache.Cache{
		Client:      m,
		KV:          &kv{m: m},
		Counter:     &counter{m: m},
		Session:     newSession(m),
		SortedSet:   newSortedSet(m),
		RateLimiter: newRateLimiter(m),
	}
}

func (m *Memory) Close() error {
	return nil
}

func (m *Memory) logOperation(ctx context.Context, op string, args ...any) {
	if m.logger == nil {
		return
	}

	logger := log.WithContext(ctx, m.logger)

	fields := []any{
		"msg", "Cache operation",
		"component", "cache",
		"backend", "memory",
		"operation", op,
	}

	var opErr error
	var opDuration time.Duration
	for i := 0; i < len(args)-1; i += 2 {
		if key, ok := args[i].(string); ok {
			if key == "error" {
				if e, ok := args[i+1].(error); ok && e != nil {
					opErr = e
				}
			} else if key == "duration" {
				if d, ok := args[i+1].(time.Duration); ok {
					opDuration = d
				}
			} else {
				fields = append(fields, key, args[i+1])
			}
		}
	}

	if opDuration <= 0 {
		opDuration = 1 * time.Nanosecond
	}
	fields = append(fields, "duration", opDuration.String())

	if opErr != nil {
		fields = append(fields, "error", opErr.Error())
		if isExpectedOperationError(opErr) {
			fields = append(fields, "expected_error", true)
			_ = logger.Log(m.logLevel, fields...)
			return
		}
		_ = logger.Log(log.LevelError, fields...)
	} else if m.slowQuery && m.slowThreshold > 0 && opDuration > m.slowThreshold {
		fields = append(fields, "slow_query", true, "threshold", m.slowThreshold.String())
		_ = logger.Log(log.LevelWarn, fields...)
	} else {
		_ = logger.Log(m.logLevel, fields...)
	}
}

func isExpectedOperationError(err error) bool {
	return errors.Is(err, cache.ErrCacheMiss) ||
		errors.Is(err, cache.ErrKeyNotFound) ||
		errors.Is(err, cache.ErrSessionNotFound) ||
		errors.Is(err, cache.ErrQuotaExceeded)
}

func (m *Memory) traceOperationWithArgs(ctx context.Context, op string, args *cache.OperationArgs, fn func(context.Context) error) error {
	if m.otel != nil {
		return m.otel.TraceOperationWithArgs(ctx, op, args, fn)
	}
	return fn(ctx)
}

type CloserFunc func() error

func (f CloserFunc) Close() error {
	return f()
}
