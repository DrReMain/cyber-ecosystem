package main

import (
	"flag"
	"os"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/encoding/json"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	"cyber-ecosystem/shared-go/kratos/transport/connect"

	"cyber-ecosystem/apps/genesis/services/mobile_bff/internal/conf"
	"cyber-ecosystem/apps/genesis/services/mobile_bff/internal/server"
)

var (
	Name     string = "genesis_service_mobile_bff"
	Version  string
	flagConf string
	id, _    = os.Hostname()
)

func init() {
	flag.StringVar(&flagConf, "conf", "../../configs", "config path, eg: -conf config.yaml")

	json.MarshalOptions = protojson.MarshalOptions{
		EmitUnpopulated: true,
		UseProtoNames:   false,
	}
	json.UnmarshalOptions = protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
}

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server, cs *connect.Server, os *server.OpsServer) *kratos.App {
	var srv []transport.Server
	if hs != nil {
		srv = append(srv, hs)
	}
	if cs != nil {
		srv = append(srv, cs)
	}
	if os != nil {
		srv = append(srv, os)
	}
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(srv...),
	)
}

func main() {
	flag.Parse()
	c := config.New(
		config.WithSource(
			file.NewSource(flagConf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	logger, loggerCleanup, err := newLogger(&bc)
	if err != nil {
		panic(err)
	}

	meterProvider, metricRequests, metricSeconds, metricsCleanup := setupMetrics()

	tp, traceCleanup, err := newTracerProvider(&bc)
	if err != nil {
		panic(err)
	}

	sentryCleanup, err := initSentry(bc.ErrorReport)
	if err != nil {
		panic(err)
	}

	app, cleanup, err := wireApp(
		bc.Server,
		bc.Log,
		bc.Data,
		bc.Ops,
		logger,
		tp,
		meterProvider,
		metricRequests,
		metricSeconds,
	)
	if err != nil {
		panic(err)
	}

	defer func() {
		cleanup()
		loggerCleanup()
		metricsCleanup()
		traceCleanup()
		sentryCleanup()
	}()

	if err := app.Run(); err != nil {
		panic(err)
	}
}
