package redis

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/shared-go/cache"
)

// ClientWrapper wraps a redis client to add logging support.
type ClientWrapper struct {
	client        *redis.Client
	logger        log.Logger
	level         log.Level
	slowQuery     bool
	slowThreshold time.Duration
}

// NewKratosClientWrapper creates a new client wrapper with kratos logging support
func NewKratosClientWrapper(client *redis.Client, logger log.Logger, level string, slowQuery bool, slowThreshold time.Duration) *ClientWrapper {
	return &ClientWrapper{
		client:        client,
		logger:        logger,
		level:         cache.ParseLogLevel(level),
		slowQuery:     slowQuery,
		slowThreshold: slowThreshold,
	}
}

// DialHook implements redis.Hook
func (h *ClientWrapper) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

// ProcessHook implements redis.Hook for command logging
func (h *ClientWrapper) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmd)
		duration := time.Since(start)

		h.logCommand(ctx, cmd, duration, err)

		return err
	}
}

// ProcessPipelineHook implements redis.Hook
func (h *ClientWrapper) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		return next(ctx, cmds)
	}
}

func (h *ClientWrapper) logCommand(ctx context.Context, cmd redis.Cmder, duration time.Duration, err error) {
	// Suppress go-redis client capability probes — not business operations
	if isExpectedRedisClientError(err) {
		return
	}

	logger := log.WithContext(ctx, h.logger)

	fields := []any{
		"msg", "Cache operation",
		"component", "cache",
		"backend", "redis",
		"operation", cmd.Name(),
	}

	if err != nil {
		fields = append(fields, "error", err.Error())
		if errors.Is(err, redis.Nil) {
			fields = append(fields, "expected_error", true)
		}
	} else if h.slowQuery && duration > h.slowThreshold {
		fields = append(fields, "slow_query", true, "threshold", h.slowThreshold.String())
	}

	fields = append(fields, "latency", duration.Seconds())

	if err != nil {
		if errors.Is(err, redis.Nil) {
			_ = logger.Log(h.level, fields...)
			return
		}
		_ = logger.Log(log.LevelError, fields...)
	} else if h.slowQuery && duration > h.slowThreshold {
		_ = logger.Log(log.LevelWarn, fields...)
	} else {
		_ = logger.Log(h.level, fields...)
	}
}

func isExpectedRedisClientError(err error) bool {
	if err == nil {
		return false
	}
	// go-redis may probe client capabilities with optional subcommands.
	return strings.Contains(err.Error(), "unknown subcommand 'maint_notifications'")
}
