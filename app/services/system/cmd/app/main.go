package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/go-kratos/kratos/v3"
	"github.com/go-kratos/kratos/v3/config"
	"github.com/go-kratos/kratos/v3/config/env"
	"github.com/go-kratos/kratos/v3/config/file"
	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	"cyber-ecosystem/shared-go/kratos/health"
	"cyber-ecosystem/shared-go/kratos/jsoncodec"
	"cyber-ecosystem/shared-go/kratos/observability"
	"cyber-ecosystem/shared-go/kratos/transport/connect"

	"cyber-ecosystem/app/services/system/internal/conf"
)

var (
	Name     string = "system"
	Version  string = "dev"
	flagconf string
	id, _    = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")

	// Override kratos's default "json" codec (Go std json, which omits proto
	// zero-value fields via struct tags) so HTTP JSON responses include zero
	// values (count=0, status=0, ...). Connect uses connect.WithCodec separately.
	jsoncodec.Register()
}

func newApp(logger *slog.Logger, gs *grpc.Server, hs *http.Server, cs *connect.Server) *kratos.App {
	return kratos.New(append([]kratos.Option{
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(gs, hs, cs),
	}, health.Mount(hs)...)...)
}

func initObservability(o *conf.Observability) (func(), *slog.Logger, error) {
	if o == nil {
		return func() {}, slog.New(slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelInfo}),
		), nil
	}

	// Each nested message is nil in proto3 if its sub-block is omitted from
	// config — guard each so a partial observability block doesn't panic.
	cfg := observability.Config{Endpoint: o.Endpoint, Insecure: o.Insecure}
	if o.Trace != nil {
		cfg.Trace = observability.TraceConfig{Enabled: o.Trace.Enabled, SamplingRatio: o.Trace.SamplingRatio}
	}
	if o.Metrics != nil {
		cfg.Metrics = observability.MetricsConfig{Enabled: o.Metrics.Enabled}
	}
	if o.Log != nil {
		logCfg := observability.LogConfig{Level: o.Log.Level, Console: o.Log.Console, OTLP: o.Log.Otlp}
		if o.Log.File != nil {
			logCfg.File = &observability.FileOutputConfig{
				Path:       o.Log.File.Path,
				MaxSizeMB:  int(o.Log.File.MaxSizeMb),
				MaxBackups: int(o.Log.File.MaxBackups),
				MaxAgeDays: int(o.Log.File.MaxAgeDays),
				Compress:   o.Log.File.Compress,
			}
		}
		cfg.Log = logCfg
	}
	if o.SlowQuery != nil {
		if d := o.SlowQuery.GetDb(); d != nil {
			cfg.SlowQuery.DB = d.AsDuration()
		}
		if d := o.SlowQuery.GetCache(); d != nil {
			cfg.SlowQuery.Cache = d.AsDuration()
		}
	}
	return observability.Init(cfg, Name, Version, id)
}

func main() {
	flag.Parse()

	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
			env.NewSource("APP_"),
		),
	)
	defer func() { _ = c.Close() }()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	shutdownObs, logger, err := initObservability(bc.Observability)
	if err != nil {
		panic(err)
	}
	defer shutdownObs()

	app, cleanup, err := wireApp(bc.Server, bc.Data, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
