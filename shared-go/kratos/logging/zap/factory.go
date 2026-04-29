package zap

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/go-kratos/kratos/v2/log"
)

type Config struct {
	Level   string
	Console *ConsoleConfig
	File    *FileConfig
	OtlpLog *OtlpLogConfig
}

type ConsoleConfig struct {
	Enabled bool
	Color   bool
	Format  string // "console" or "json", default "console"
}

type FileConfig struct {
	Enabled    bool
	Path       string
	MaxSize    int // MB
	MaxBackups int
	MaxAge     int // days
	Compress   bool
}

func NewLoggerFromConfig(cfg *Config) (log.Logger, func(), error) {
	if cfg == nil {
		return log.NewStdLogger(os.Stdout), func() {}, nil
	}

	cores := make([]zapcore.Core, 0)
	closers := make([]func(), 0)

	level := parseLevel(cfg.Level)

	if cfg.Console != nil && cfg.Console.Enabled {
		core := buildConsoleCore(cfg.Console, level)
		cores = append(cores, core)
	}

	if cfg.File != nil && cfg.File.Enabled {
		core, closer := buildFileCore(cfg.File, level)
		cores = append(cores, core)
		if closer != nil {
			closers = append(closers, closer)
		}
	}

	if cfg.OtlpLog != nil && cfg.OtlpLog.Enabled {
		core, closer, err := buildOtlpLogCore(cfg.OtlpLog, level)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create OTLP log core: %w", err)
		}
		cores = append(cores, core)
		if closer != nil {
			closers = append(closers, closer)
		}
	}

	if len(cores) == 0 {
		core := buildConsoleCore(&ConsoleConfig{Enabled: true, Color: true}, level)
		cores = append(cores, core)
	}

	tee := zapcore.NewTee(cores...)

	// Create zap logger with caller skip for proper caller information
	zapLogger := zap.New(tee, zap.AddCaller(), zap.AddCallerSkip(3))

	logger := NewLogger(zapLogger, closers...)

	cleanup := func() {
		_ = logger.Close()
	}
	return logger, cleanup, nil
}

func parseLevel(s string) zapcore.Level {
	switch s {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

func buildConsoleCore(cfg *ConsoleConfig, level zapcore.Level) zapcore.Core {
	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = buildJSONEncoder()
	} else {
		encoder = buildConsoleEncoder(cfg.Color)
	}
	writer := zapcore.AddSync(os.Stdout)
	return zapcore.NewCore(encoder, writer, level)
}

func buildFileCore(cfg *FileConfig, level zapcore.Level) (zapcore.Core, func()) {
	writer := &lumberjack.Logger{
		Filename:   cfg.Path,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	}
	encoder := buildJSONEncoder()
	return zapcore.NewCore(encoder, zapcore.AddSync(writer), level), func() { writer.Close() }
}

func buildOtlpLogCore(cfg *OtlpLogConfig, level zapcore.Level) (zapcore.Core, func(), error) {
	writer, err := NewOtlpLogWriter(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("create OTLP log writer: %w", err)
	}
	encoder := buildJSONEncoder()
	return zapcore.NewCore(encoder, zapcore.AddSync(writer), level), func() { writer.Close() }, nil
}

func buildConsoleEncoder(color bool) zapcore.Encoder {
	cfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if color {
		cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	return zapcore.NewConsoleEncoder(cfg)
}

func buildJSONEncoder() zapcore.Encoder {
	cfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	return zapcore.NewJSONEncoder(cfg)
}
