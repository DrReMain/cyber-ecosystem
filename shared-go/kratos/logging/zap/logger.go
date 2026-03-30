package zap

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/go-kratos/kratos/v2/log"
)

type Logger struct {
	log     *zap.Logger
	msgKey  string
	closers []func()
}

type Option func(*Logger)

func WithMessageKey(key string) Option {
	return func(l *Logger) {
		l.msgKey = key
	}
}

func NewLogger(zapLogger *zap.Logger, closers ...func()) *Logger {
	return &Logger{
		log:     zapLogger,
		msgKey:  log.DefaultMessageKey,
		closers: closers,
	}
}

func (l *Logger) Log(level log.Level, keyvals ...any) error {
	if zapcore.Level(level) < zapcore.DPanicLevel && !l.log.Core().Enabled(zapcore.Level(level)) {
		return nil
	}

	var (
		msg    = ""
		keylen = len(keyvals)
	)

	if keylen == 0 || keylen%2 != 0 {
		l.log.Warn(fmt.Sprintf("Keyvalues must appear in pairs: %v", keyvals))
		return nil
	}

	fields := make([]zap.Field, 0, (keylen/2)+1)
	for i := 0; i < keylen; i += 2 {
		key, ok := keyvals[i].(string)
		if !ok {
			key = fmt.Sprint(keyvals[i])
		}
		if key == l.msgKey {
			msg, _ = keyvals[i+1].(string)
			continue
		}
		fields = append(fields, zap.Any(key, keyvals[i+1]))
	}

	switch level {
	case log.LevelDebug:
		l.log.Debug(msg, fields...)
	case log.LevelInfo:
		l.log.Info(msg, fields...)
	case log.LevelWarn:
		l.log.Warn(msg, fields...)
	case log.LevelError:
		l.log.Error(msg, fields...)
	case log.LevelFatal:
		l.log.Fatal(msg, fields...)
	}

	return nil
}

func (l *Logger) Sync() error {
	if l.log == nil {
		return nil
	}
	return l.log.Sync()
}

func (l *Logger) Close() error {
	_ = l.Sync()
	for _, closer := range l.closers {
		closer()
	}
	return nil
}

func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{
		log:     l.log.With(fields...),
		msgKey:  l.msgKey,
		closers: l.closers,
	}
}

func (l *Logger) ZapLogger() *zap.Logger {
	return l.log
}
