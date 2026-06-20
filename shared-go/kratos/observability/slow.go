package observability

import (
	"context"
	"log/slog"
	"time"

	"entgo.io/ent/dialect"
	"github.com/redis/go-redis/v9"
)

// slowDB / slowCache are the slow-op thresholds populated by Init (observability.go)
// from cfg.SlowQuery. Zero disables. Kept package-level so the platform layer wires
// driver/client without threading thresholds through every constructor.
var (
	slowDB    time.Duration
	slowCache time.Duration
)

// slowLogQueryLimit caps the SQL string emitted in a slow-query warning so very
// large statements don't flood the log.
const slowLogQueryLimit = 512

// WrapSlowQueryDriver returns a dialect.Driver proxy that logs a WARN for any
// Exec/Query — including statements inside a transaction — exceeding the db slow
// threshold. When the threshold is zero (disabled) or inner is nil it returns
// inner unchanged (a no-op). logger may be nil (defaults to slog.Default). The
// logger's context carries the active span, so the warning correlates with the
// request span in SigNoz.
func WrapSlowQueryDriver(inner dialect.Driver, logger *slog.Logger) dialect.Driver {
	if slowDB == 0 || inner == nil {
		return inner
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &slowDriver{Driver: inner, logger: logger, threshold: slowDB}
}

// AttachSlowRedisHook adds a redis.Hook that logs a WARN for any command
// exceeding the cache slow threshold. No-op when the threshold is zero or the
// client is nil.
func AttachSlowRedisHook(client *redis.Client, logger *slog.Logger) {
	if slowCache == 0 || client == nil {
		return
	}
	if logger == nil {
		logger = slog.Default()
	}
	client.AddHook(&slowHook{logger: logger, threshold: slowCache})
}

// slowDriver wraps a dialect.Driver, timing Exec/Query against a threshold. Tx
// returns a wrapped dialect.Tx so statements inside a transaction are timed too;
// Close/Dialect delegate to the inner driver via the embed.
type slowDriver struct {
	dialect.Driver
	logger    *slog.Logger
	threshold time.Duration
}

func (d *slowDriver) Exec(ctx context.Context, query string, args, v any) error {
	start := time.Now()
	err := d.Driver.Exec(ctx, query, args, v)
	if dur := time.Since(start); dur >= d.threshold {
		d.logger.WarnContext(ctx, "slow db query",
			slog.String("query", trunc(query, slowLogQueryLimit)),
			slog.Duration("duration", dur),
			slog.Any("err", err),
		)
	}
	return err
}

func (d *slowDriver) Query(ctx context.Context, query string, args, v any) error {
	start := time.Now()
	err := d.Driver.Query(ctx, query, args, v)
	if dur := time.Since(start); dur >= d.threshold {
		d.logger.WarnContext(ctx, "slow db query",
			slog.String("query", trunc(query, slowLogQueryLimit)),
			slog.Duration("duration", dur),
			slog.Any("err", err),
		)
	}
	return err
}

func (d *slowDriver) Tx(ctx context.Context) (dialect.Tx, error) {
	tx, err := d.Driver.Tx(ctx)
	if err != nil {
		return nil, err
	}
	return &slowTx{Tx: tx, logger: d.logger, threshold: d.threshold}, nil
}

// slowTx wraps a dialect.Tx to time its Exec/Query against the threshold.
// Commit/Rollback delegate to the inner Tx via the embed.
type slowTx struct {
	dialect.Tx
	logger    *slog.Logger
	threshold time.Duration
}

func (t *slowTx) Exec(ctx context.Context, query string, args, v any) error {
	start := time.Now()
	err := t.Tx.Exec(ctx, query, args, v)
	if dur := time.Since(start); dur >= t.threshold {
		t.logger.WarnContext(ctx, "slow db query",
			slog.String("query", trunc(query, slowLogQueryLimit)),
			slog.Duration("duration", dur),
			slog.Bool("tx", true),
			slog.Any("err", err),
		)
	}
	return err
}

func (t *slowTx) Query(ctx context.Context, query string, args, v any) error {
	start := time.Now()
	err := t.Tx.Query(ctx, query, args, v)
	if dur := time.Since(start); dur >= t.threshold {
		t.logger.WarnContext(ctx, "slow db query",
			slog.String("query", trunc(query, slowLogQueryLimit)),
			slog.Duration("duration", dur),
			slog.Bool("tx", true),
			slog.Any("err", err),
		)
	}
	return err
}

// slowHook is a redis.Hook timing each command against the cache threshold.
// DialHook and ProcessPipelineHook pass through unchanged.
type slowHook struct {
	logger    *slog.Logger
	threshold time.Duration
}

func (h *slowHook) DialHook(next redis.DialHook) redis.DialHook { return next }

func (h *slowHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return next
}

func (h *slowHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmd)
		if dur := time.Since(start); dur >= h.threshold {
			h.logger.WarnContext(ctx, "slow redis op",
				slog.String("cmd", cmd.FullName()),
				slog.Duration("duration", dur),
				slog.Any("err", err),
			)
		}
		return err
	}
}

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
