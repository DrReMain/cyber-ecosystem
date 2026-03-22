package entlog

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/go-kratos/kratos/v2/log"
)

var whitespaceRegex = regexp.MustCompile(`\s+`)
var dqmRegex = regexp.MustCompile(`"`)

func compactQuery(query string) string {
	result := strings.TrimSpace(whitespaceRegex.ReplaceAllString(query, " "))
	result = strings.TrimSpace(dqmRegex.ReplaceAllString(result, "'"))
	return result
}

// Logger implements ent.Log function signature
// This is a simple logger for Ent that logs SQL queries without trace information.
// For proper distributed tracing, use otelsql to wrap the database driver,
// which will automatically create spans for each SQL operation.
type Logger struct {
	logger             log.Logger
	level              log.Level
	slowQuery          bool
	slowQueryThreshold time.Duration
}

func NewLogger(logger log.Logger, level string, slowQuery bool, slowQueryThreshold time.Duration) *Logger {
	return &Logger{
		logger:             logger,
		level:              parseLogLevel(level),
		slowQuery:          slowQuery,
		slowQueryThreshold: slowQueryThreshold,
	}
}

// Log implements ent.Log function signature (without context)
func (l *Logger) Log(args ...interface{}) {
	if len(args) == 0 {
		return
	}

	msg := fmt.Sprint(args...)
	_ = l.logger.Log(l.level, "msg", msg, "component", "ent")
}

// DriverWrapper wraps an ent dialect.Driver to add logging support.
// It extracts trace information from context (created by otelsql) and includes it in logs.
type DriverWrapper struct {
	drv                *entsql.Driver
	logger             log.Logger
	level              log.Level
	slowQuery          bool
	slowQueryThreshold time.Duration
}

// NewDriverWrapper creates a new driver wrapper with logging support
func NewDriverWrapper(drv *entsql.Driver, logger log.Logger, level string, slowQuery bool, slowQueryThreshold time.Duration) *DriverWrapper {
	return &DriverWrapper{
		drv:                drv,
		logger:             logger,
		level:              parseLogLevel(level),
		slowQuery:          slowQuery,
		slowQueryThreshold: slowQueryThreshold,
	}
}

// Exec implements dialect.Driver.Exec
func (d *DriverWrapper) Exec(ctx context.Context, query string, args, v any) error {
	start := time.Now()
	err := d.drv.Exec(ctx, query, args, v)
	duration := time.Since(start)

	d.logQuery(ctx, query, args, duration, err)

	return err
}

// Query implements dialect.Driver.Query
func (d *DriverWrapper) Query(ctx context.Context, query string, args, v any) error {
	start := time.Now()
	err := d.drv.Query(ctx, query, args, v)
	duration := time.Since(start)

	d.logQuery(ctx, query, args, duration, err)

	return err
}

// Tx implements dialect.Driver.Tx
func (d *DriverWrapper) Tx(ctx context.Context) (dialect.Tx, error) {
	return d.drv.Tx(ctx)
}

// Close implements dialect.Driver.Close
func (d *DriverWrapper) Close() error {
	return d.drv.Close()
}

// Dialect implements dialect.Driver.Dialect
func (d *DriverWrapper) Dialect() string {
	return d.drv.Dialect()
}

// logQuery logs the SQL query with trace information extracted from context
// otelsql creates a span for each SQL operation, we extract the span info here
func (d *DriverWrapper) logQuery(ctx context.Context, query string, args any, duration time.Duration, err error) {
	// Use log.WithContext to get a logger that includes trace information from context
	// This works with Kratos tracing middleware which sets trace.id and span.id as Valuer functions
	logger := log.WithContext(ctx, d.logger)

	fields := []any{
		"msg", "SQL query",
		"component", "ent",
		"query", compactQuery(query), // Compact query to single line
		"args", fmt.Sprintf("%v", args),
		"duration", duration.String(),
	}

	if err != nil {
		fields = append(fields, "error", err.Error())
		_ = logger.Log(log.LevelError, fields...)
	} else if d.slowQuery && duration > d.slowQueryThreshold {
		fields = append(fields, "slow_query", true, "threshold", d.slowQueryThreshold.String())
		_ = logger.Log(log.LevelWarn, fields...)
	} else {
		_ = logger.Log(d.level, fields...)
	}
}

func parseLogLevel(s string) log.Level {
	switch s {
	case "debug":
		return log.LevelDebug
	case "info":
		return log.LevelInfo
	case "warn":
		return log.LevelWarn
	case "error":
		return log.LevelError
	default:
		return log.LevelInfo
	}
}
