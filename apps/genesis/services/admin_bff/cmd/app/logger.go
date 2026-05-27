package main

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"

	zaplog "cyber-ecosystem/shared-go/kratos/logging/zap"

	"cyber-ecosystem/apps/genesis/services/admin_bff/internal/conf"
)

func newLogger(bc *conf.Bootstrap) (log.Logger, func(), error) {
	logger, cleanup, err := zaplog.NewLoggerFromConfig(convertLogConfig(bc.Log))
	if err != nil {
		return nil, nil, err
	}
	log.SetLogger(logger)
	logger = log.With(logger,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)
	return logger, cleanup, nil
}

func convertLogConfig(c *conf.Log) *zaplog.Config {
	if c == nil {
		return nil
	}

	cfg := &zaplog.Config{
		Level: c.Level,
	}

	if c.Console != nil {
		cfg.Console = &zaplog.ConsoleConfig{
			Enabled: c.Console.Enabled,
			Color:   c.Console.Color,
			Format:  c.Console.Format,
		}
	}

	if c.File != nil {
		cfg.File = &zaplog.FileConfig{
			Enabled:    c.File.Enabled,
			Path:       c.File.Path,
			MaxSize:    int(c.File.MaxSize),
			MaxBackups: int(c.File.MaxBackups),
			MaxAge:     int(c.File.MaxAge),
			Compress:   c.File.Compress,
		}
	}

	if c.OtlpLog != nil {
		cfg.OtlpLog = &zaplog.OtlpLogConfig{
			Enabled:        c.OtlpLog.Enabled,
			Endpoint:       c.OtlpLog.Endpoint,
			Insecure:       c.OtlpLog.Insecure,
			ServiceName:    Name,
			ServiceID:      id,
			ServiceVersion: Version,
		}
	}

	return cfg
}
