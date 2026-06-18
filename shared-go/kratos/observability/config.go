package observability

// Config is the service-agnostic observability configuration. A service maps
// its own generated conf onto this struct before calling Init.
type Config struct {
	Endpoint string // shared OTLP/HTTP collector (e.g. SigNoz host:4318); empty = no collector
	Insecure bool
	Trace    TraceConfig
	Metrics  MetricsConfig
	Log      LogConfig
}

type TraceConfig struct {
	Enabled       bool
	SamplingRatio float64 // 0~1; <=0 or >1 defaults to 1.0 (no sampling)
}

type MetricsConfig struct {
	Enabled bool
}

type LogConfig struct {
	Level   string            // debug/info/warn/error; default info
	Console bool              // stdout (text)
	OTLP    bool              // OTLP log sink (uses Endpoint)
	File    *FileOutputConfig // nil = no file sink
}

type FileOutputConfig struct {
	Path       string
	MaxSizeMB  int // megabytes before rotation
	MaxBackups int // rotated files to retain
	MaxAgeDays int // days to retain
	Compress   bool
}
