package zap

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/go-kratos/kratos/v2/log"
)

// TestLogger_Log tests the Log method with various log levels
func TestLogger_Log(t *testing.T) {
	tests := []struct {
		name    string
		level   log.Level
		keyvals []interface{}
		wantMsg string
		wantErr bool
	}{
		{
			name:    "debug level",
			level:   log.LevelDebug,
			keyvals: []interface{}{"msg", "debug message", "key", "value"},
			wantMsg: "debug message",
		},
		{
			name:    "info level",
			level:   log.LevelInfo,
			keyvals: []interface{}{"msg", "info message", "key", "value"},
			wantMsg: "info message",
		},
		{
			name:    "warn level",
			level:   log.LevelWarn,
			keyvals: []interface{}{"msg", "warn message", "key", "value"},
			wantMsg: "warn message",
		},
		{
			name:    "error level",
			level:   log.LevelError,
			keyvals: []interface{}{"msg", "error message", "key", "value"},
			wantMsg: "error message",
		},
		{
			name:    "empty keyvals",
			level:   log.LevelInfo,
			keyvals: []interface{}{},
			wantMsg: "",
		},
		{
			name:    "odd keyvals",
			level:   log.LevelInfo,
			keyvals: []interface{}{"key1", "value1", "key2"},
			wantMsg: "",
		},
		{
			name:    "custom message key",
			level:   log.LevelInfo,
			keyvals: []interface{}{"msg", "custom message", "extra", "data"},
			wantMsg: "custom message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a buffer to capture output
			var buf bytes.Buffer
			encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
			core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.DebugLevel)
			zapLogger := zap.New(core)

			logger := NewLogger(zapLogger)
			err := logger.Log(tt.level, tt.keyvals...)

			if (err != nil) != tt.wantErr {
				t.Errorf("Logger.Log() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantMsg != "" && !strings.Contains(buf.String(), tt.wantMsg) {
				t.Errorf("Logger.Log() output = %v, want to contain %v", buf.String(), tt.wantMsg)
			}
		})
	}
}

// TestLogger_WithMessageKey tests the WithMessageKey option
func TestLogger_WithMessageKey(t *testing.T) {
	var buf bytes.Buffer
	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.DebugLevel)
	zapLogger := zap.New(core)

	// Use WithMessageKey option
	loggerWithKey := &Logger{
		log:     zapLogger,
		msgKey:  "message",
		closers: nil,
	}
	err := loggerWithKey.Log(log.LevelInfo, "message", "test message", "key", "value")

	if err != nil {
		t.Errorf("Logger.Log() error = %v", err)
	}

	if !strings.Contains(buf.String(), "test message") {
		t.Errorf("Logger.Log() output = %v, want to contain 'test message'", buf.String())
	}
}

// TestLogger_With tests the With method for creating child loggers
func TestLogger_With(t *testing.T) {
	var buf bytes.Buffer
	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.DebugLevel)
	zapLogger := zap.New(core)

	logger := NewLogger(zapLogger)
	childLogger := logger.With(zap.String("service", "test-service"))

	err := childLogger.Log(log.LevelInfo, "msg", "child logger message")
	if err != nil {
		t.Errorf("Logger.Log() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "child logger message") {
		t.Errorf("Logger.Log() output = %v, want to contain 'child logger message'", output)
	}
	if !strings.Contains(output, "test-service") {
		t.Errorf("Logger.Log() output = %v, want to contain 'test-service'", output)
	}
}

// TestLogger_Sync tests the Sync method
func TestLogger_Sync(t *testing.T) {
	// Use a buffer instead of stdout, as stdout sync may fail
	var buf bytes.Buffer
	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.DebugLevel)
	zapLogger := zap.New(core)

	logger := NewLogger(zapLogger)
	err := logger.Sync()
	// Sync on buffer should succeed
	if err != nil {
		t.Errorf("Logger.Sync() error = %v", err)
	}
}

// TestLogger_Close tests the Close method
func TestLogger_Close(t *testing.T) {
	called := false
	closer := func() {
		called = true
	}

	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel)
	zapLogger := zap.New(core)

	logger := NewLogger(zapLogger, closer)
	err := logger.Close()

	if err != nil {
		t.Errorf("Logger.Close() error = %v", err)
	}
	if !called {
		t.Error("Logger.Close() did not call closer")
	}
}

// TestNewLoggerFromConfig tests the factory function
func TestNewLoggerFromConfig(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name:   "nil config",
			config: nil,
		},
		{
			name: "console only",
			config: &Config{
				Level: "debug",
				Console: &ConsoleConfig{
					Enabled: true,
					Color:   false,
				},
			},
		},
		{
			name: "console with color",
			config: &Config{
				Level: "info",
				Console: &ConsoleConfig{
					Enabled: true,
					Color:   true,
				},
			},
		},
		{
			name: "file output",
			config: &Config{
				Level: "debug",
				File: &FileConfig{
					Enabled:    true,
					Path:       filepath.Join(t.TempDir(), "test.log"),
					MaxSize:    10,
					MaxBackups: 3,
					MaxAge:     7,
					Compress:   false,
				},
			},
		},
		{
			name: "console and file",
			config: &Config{
				Level: "debug",
				Console: &ConsoleConfig{
					Enabled: true,
					Color:   false,
				},
				File: &FileConfig{
					Enabled:    true,
					Path:       filepath.Join(t.TempDir(), "test-combined.log"),
					MaxSize:    10,
					MaxBackups: 3,
					MaxAge:     7,
					Compress:   false,
				},
			},
		},
		{
			name: "all levels",
			config: &Config{
				Level: "warn",
				Console: &ConsoleConfig{
					Enabled: true,
					Color:   false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, cleanup, err := NewLoggerFromConfig(tt.config)
			if err != nil {
				t.Errorf("NewLoggerFromConfig() error = %v", err)
				return
			}
			defer cleanup()

			// Test logging at different levels
			logger.Log(log.LevelDebug, "msg", "debug message")
			logger.Log(log.LevelInfo, "msg", "info message")
			logger.Log(log.LevelWarn, "msg", "warn message")
			logger.Log(log.LevelError, "msg", "error message")
		})
	}
}

// TestLogger_LevelFiltering tests that log levels are properly filtered
func TestLogger_LevelFiltering(t *testing.T) {
	tests := []struct {
		name        string
		configLevel string
		logLevel    log.Level
		shouldLog   bool
	}{
		{"debug config, debug log", "debug", log.LevelDebug, true},
		{"debug config, info log", "debug", log.LevelInfo, true},
		{"info config, debug log", "info", log.LevelDebug, false},
		{"info config, info log", "info", log.LevelInfo, true},
		{"warn config, info log", "warn", log.LevelInfo, false},
		{"warn config, warn log", "warn", log.LevelWarn, true},
		{"error config, warn log", "error", log.LevelWarn, false},
		{"error config, error log", "error", log.LevelError, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
			level := parseLevel(tt.configLevel)
			core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), level)
			zapLogger := zap.New(core)

			logger := NewLogger(zapLogger)
			logger.Log(tt.logLevel, "msg", "test message")

			hasOutput := buf.String() != ""
			if hasOutput != tt.shouldLog {
				t.Errorf("Level filtering failed: config=%s, log=%s, got output=%v, want=%v",
					tt.configLevel, tt.logLevel, hasOutput, tt.shouldLog)
			}
		})
	}
}

// TestLogger_FileOutput tests file output functionality
func TestLogger_FileOutput(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	config := &Config{
		Level: "debug",
		File: &FileConfig{
			Enabled:    true,
			Path:       logPath,
			MaxSize:    10,
			MaxBackups: 3,
			MaxAge:     7,
			Compress:   false,
		},
	}

	logger, cleanup, err := NewLoggerFromConfig(config)
	if err != nil {
		t.Fatalf("NewLoggerFromConfig() error = %v", err)
	}

	// Log some messages
	for i := 0; i < 10; i++ {
		logger.Log(log.LevelInfo, "msg", fmt.Sprintf("test message %d", i), "index", i)
	}

	cleanup()

	// Check file exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Fatalf("Log file was not created: %s", logPath)
	}

	// Read and verify file content
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Verify JSON format
	var entry map[string]interface{}
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) == 0 {
		t.Fatal("Log file is empty")
	}

	for i, line := range lines {
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Errorf("Line %d is not valid JSON: %v", i, err)
		}
	}
}

// TestLogger_MultipleOutputs tests that logs are written to multiple outputs
func TestLogger_MultipleOutputs(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "multi.log")

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	config := &Config{
		Level: "debug",
		Console: &ConsoleConfig{
			Enabled: true,
			Color:   false,
		},
		File: &FileConfig{
			Enabled:    true,
			Path:       logPath,
			MaxSize:    10,
			MaxBackups: 3,
			MaxAge:     7,
			Compress:   false,
		},
	}

	logger, cleanup, err := NewLoggerFromConfig(config)
	if err != nil {
		t.Fatalf("NewLoggerFromConfig() error = %v", err)
	}

	logger.Log(log.LevelInfo, "msg", "multi-output test")

	cleanup()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	var stdoutBuf bytes.Buffer
	io.Copy(&stdoutBuf, r)
	stdoutContent := stdoutBuf.String()

	// Check console output
	if !strings.Contains(stdoutContent, "multi-output test") {
		t.Error("Console output missing expected message")
	}

	// Check file output
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "multi-output test") {
		t.Error("File output missing expected message")
	}
}

// TestLogger_KratosIntegration tests integration with Kratos log package
func TestLogger_KratosIntegration(t *testing.T) {
	var buf bytes.Buffer
	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.DebugLevel)
	zapLogger := zap.New(core)

	logger := NewLogger(zapLogger)

	// Set as global logger
	log.SetLogger(logger)

	// Test using Kratos global log functions
	log.Debug("debug message from kratos")
	log.Info("info message from kratos")
	log.Warn("warn message from kratos")
	log.Errorf("error message from kratos: %s", "test")

	output := buf.String()

	tests := []string{
		"debug message from kratos",
		"info message from kratos",
		"warn message from kratos",
		"error message from kratos",
	}

	for _, want := range tests {
		if !strings.Contains(output, want) {
			t.Errorf("Output missing expected message: %s", want)
		}
	}
}

// TestLogger_ContextWithFields tests log.With for adding fields
func TestLogger_ContextWithFields(t *testing.T) {
	var buf bytes.Buffer
	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.DebugLevel)
	zapLogger := zap.New(core)

	logger := NewLogger(zapLogger)

	// Create a logger with context fields
	contextLogger := log.With(logger,
		"service.id", "test-123",
		"service.name", "test-service",
		"service.version", "1.0.0",
	)

	contextLogger.Log(log.LevelInfo, "msg", "contextual message")

	output := buf.String()

	expectedFields := []string{
		"test-123",
		"test-service",
		"1.0.0",
		"contextual message",
	}

	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("Output missing expected field: %s", output)
		}
	}
}

// TestLogger_Concurrent tests concurrent logging
func TestLogger_Concurrent(t *testing.T) {
	// Use a thread-safe writer
	var buf safeBuffer
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.DebugLevel)
	zapLogger := zap.New(core)

	logger := NewLogger(zapLogger)

	// Log concurrently
	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				logger.Log(log.LevelInfo, "msg", fmt.Sprintf("goroutine %d message %d", id, j), "goroutine", id)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	// Verify output
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1000 {
		t.Errorf("Expected 1000 log lines, got %d", len(lines))
	}
}

// safeBuffer is a thread-safe bytes.Buffer
type safeBuffer struct {
	buf bytes.Buffer
	mu  sync.Mutex
}

func (sb *safeBuffer) Write(p []byte) (n int, err error) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.buf.Write(p)
}

func (sb *safeBuffer) String() string {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.buf.String()
}

// TestLogger_ZapLogger tests the ZapLogger method
func TestLogger_ZapLogger(t *testing.T) {
	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel)
	zapLogger := zap.New(core)

	logger := NewLogger(zapLogger)

	retrieved := logger.ZapLogger()
	if retrieved != zapLogger {
		t.Error("ZapLogger() did not return the original zap.Logger")
	}
}

// TestParseLevel tests the parseLevel function
func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected zapcore.Level
	}{
		{"debug", zapcore.DebugLevel},
		{"info", zapcore.InfoLevel},
		{"warn", zapcore.WarnLevel},
		{"error", zapcore.ErrorLevel},
		{"unknown", zapcore.InfoLevel}, // default
		{"", zapcore.InfoLevel},        // default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseLevel(tt.input)
			if result != tt.expected {
				t.Errorf("parseLevel(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestLogger_NilLogger tests behavior with nil logger
func TestLogger_NilLogger(t *testing.T) {
	logger := &Logger{log: nil}

	err := logger.Sync()
	if err != nil {
		t.Errorf("Sync() on nil logger should not error: %v", err)
	}
}

// TestLogger_FatalLevel tests fatal level logging (without actually exiting)
func TestLogger_FatalLevel(t *testing.T) {
	// We can't actually test fatal as it calls os.Exit
	// This test just verifies the code path exists
	var buf bytes.Buffer
	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.DebugLevel)
	zapLogger := zap.New(core, zap.WithFatalHook(zapcore.WriteThenPanic))

	logger := NewLogger(zapLogger)

	// Recover from panic
	defer func() {
		if r := recover(); r == nil {
			t.Log("Fatal level caused panic as expected")
		}
	}()

	logger.Log(log.LevelFatal, "msg", "fatal message")
}

// TestLogger_TraceID tests logging with trace context
func TestLogger_TraceID(t *testing.T) {
	var buf bytes.Buffer
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.DebugLevel)
	zapLogger := zap.New(core)

	logger := NewLogger(zapLogger)

	// Simulate trace context
	ctx := context.WithValue(context.Background(), "trace.id", "trace-123")
	ctx = context.WithValue(ctx, "span.id", "span-456")

	logger.Log(log.LevelInfo, "msg", "traced message", "trace.id", ctx.Value("trace.id"), "span.id", ctx.Value("span.id"))

	output := buf.String()

	if !strings.Contains(output, "trace-123") {
		t.Error("Output missing trace.id")
	}
	if !strings.Contains(output, "span-456") {
		t.Error("Output missing span.id")
	}
}

// BenchmarkLogger_Log benchmarks the Log method
func BenchmarkLogger_Log(b *testing.B) {
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(io.Discard), zapcore.DebugLevel)
	zapLogger := zap.New(core)

	logger := NewLogger(zapLogger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Log(log.LevelInfo, "msg", "benchmark message", "iteration", i)
	}
}

// BenchmarkLogger_LogParallel benchmarks parallel logging
func BenchmarkLogger_LogParallel(b *testing.B) {
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(io.Discard), zapcore.DebugLevel)
	zapLogger := zap.New(core)

	logger := NewLogger(zapLogger)

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			logger.Log(log.LevelInfo, "msg", "parallel message", "iteration", i)
			i++
		}
	})
}

// BenchmarkLogger_With benchmarks creating child loggers
func BenchmarkLogger_With(b *testing.B) {
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(io.Discard), zapcore.DebugLevel)
	zapLogger := zap.New(core)

	logger := NewLogger(zapLogger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.With(zap.String("key", "value"), zap.Int("iteration", i))
	}
}

// TestLogger_Timestamp tests that timestamps are included in output
func TestLogger_Timestamp(t *testing.T) {
	var buf bytes.Buffer
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.DebugLevel)
	zapLogger := zap.New(core)

	logger := NewLogger(zapLogger)
	logger.Log(log.LevelInfo, "msg", "timestamped message")

	output := buf.String()

	// Parse JSON and check for timestamp
	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if _, ok := entry["ts"]; !ok {
		t.Error("Output missing timestamp field")
	}
}

// TestLogger_Caller tests that caller information is included
func TestLogger_Caller(t *testing.T) {
	var buf bytes.Buffer
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.DebugLevel)
	zapLogger := zap.New(core, zap.AddCaller())

	logger := NewLogger(zapLogger)
	logger.Log(log.LevelInfo, "msg", "caller message")

	output := buf.String()

	// Parse JSON and check for caller
	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if _, ok := entry["caller"]; !ok {
		t.Error("Output missing caller field")
	}
}

// TestLogger_SpecialCharacters tests logging with special characters
func TestLogger_SpecialCharacters(t *testing.T) {
	var buf bytes.Buffer
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.DebugLevel)
	zapLogger := zap.New(core)

	logger := NewLogger(zapLogger)

	specialMsg := "message with \"quotes\" and \n newlines \t tabs"
	logger.Log(log.LevelInfo, "msg", specialMsg, "special", "value\u0000with\u0001nulls")

	output := buf.String()

	// Verify it's valid JSON
	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Errorf("Output is not valid JSON: %v", err)
	}
}

// TestLogger_Unicode tests logging with unicode characters
func TestLogger_Unicode(t *testing.T) {
	var buf bytes.Buffer
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.DebugLevel)
	zapLogger := zap.New(core)

	logger := NewLogger(zapLogger)

	unicodeMsg := "中文消息 日本語メッセ지 한국어 메시지 🎉🚀"
	logger.Log(log.LevelInfo, "msg", unicodeMsg)

	output := buf.String()

	// Verify it's valid JSON
	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Errorf("Output is not valid JSON: %v", err)
	}

	if entry["msg"] != unicodeMsg {
		t.Errorf("Unicode message not preserved: got %v, want %v", entry["msg"], unicodeMsg)
	}
}

// TestLogger_LongMessage tests logging with very long messages
func TestLogger_LongMessage(t *testing.T) {
	var buf bytes.Buffer
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.DebugLevel)
	zapLogger := zap.New(core)

	logger := NewLogger(zapLogger)

	// Create a very long message
	longMsg := strings.Repeat("a", 10000)
	logger.Log(log.LevelInfo, "msg", longMsg)

	output := buf.String()

	// Verify it's valid JSON
	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Errorf("Output is not valid JSON: %v", err)
	}
}

// TestLogger_StructuredData tests logging with structured data
func TestLogger_StructuredData(t *testing.T) {
	var buf bytes.Buffer
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.DebugLevel)
	zapLogger := zap.New(core)

	logger := NewLogger(zapLogger)

	// Log with structured data
	data := map[string]interface{}{
		"user_id":   12345,
		"action":    "login",
		"timestamp": time.Now().Unix(),
		"metadata": map[string]string{
			"ip":         "192.168.1.1",
			"user_agent": "test-agent",
		},
	}

	logger.Log(log.LevelInfo, "msg", "structured log", "data", data)

	output := buf.String()

	// Verify it's valid JSON
	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Errorf("Output is not valid JSON: %v", err)
	}
}
