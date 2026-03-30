package server

import (
	"context"
	"net/http"
	"net/http/pprof"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/apps/app_1/services/service_1/internal/conf"
)

type OpsServer struct {
	httpServer *http.Server
	logger     log.Logger
}

func NewOpsServer(c *conf.Ops, logger log.Logger) *OpsServer {
	if !c.Enabled {
		return nil
	}
	mux := http.NewServeMux()

	if c.Metrics != "" && len(c.Metrics) > 1 {
		mux.Handle(c.Metrics, promhttp.Handler())
	}

	if c.Pprof != nil && c.Pprof.Enabled {
		registerPprofEndpoints(mux, c.Pprof, logger)
	}

	return &OpsServer{
		httpServer: &http.Server{
			Addr:    c.Addr,
			Handler: mux,
		},
		logger: logger,
	}
}

func registerPprofEndpoints(mux *http.ServeMux, c *conf.Ops_Pprof, logger log.Logger) {
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	if c.CpuEnabled {
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	}
	if c.HeapEnabled {
		mux.HandleFunc("/debug/pprof/heap", pprof.Handler("heap").ServeHTTP)
	}
	if c.GoroutineEnabled {
		mux.HandleFunc("/debug/pprof/goroutine", pprof.Handler("goroutine").ServeHTTP)
	}
	if c.MutexEnabled {
		mux.HandleFunc("/debug/pprof/mutex", pprof.Handler("mutex").ServeHTTP)
	}
	if c.ThreadEnabled {
		mux.HandleFunc("/debug/pprof/threadcreate", pprof.Handler("threadcreate").ServeHTTP)
	}
	if c.TraceEnabled {
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}
}

func (s *OpsServer) Start(ctx context.Context) error {
	return s.httpServer.ListenAndServe()
}

func (s *OpsServer) Stop(ctx context.Context) error {
	return s.httpServer.Close()
}
