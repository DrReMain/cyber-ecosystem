package server

import (
	"context"
	"net/http"
	"net/http/pprof"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"cyber-ecosystem/apps/genesis/services/base/internal/conf"
)

type OpsServer struct {
	httpServer *http.Server
}

func NewOpsServer(c *conf.Ops) *OpsServer {
	if !c.Enabled {
		return nil
	}
	mux := http.NewServeMux()

	if c.Metrics != "" && len(c.Metrics) > 1 {
		mux.Handle(c.Metrics, promhttp.Handler())
	}

	if c.Pprof != nil && c.Pprof.Enabled {
		registerPprofEndpoints(mux, c.Pprof)
	}

	return &OpsServer{
		httpServer: &http.Server{
			Addr:    c.Addr,
			Handler: mux,
		},
	}
}

func registerPprofEndpoints(mux *http.ServeMux, c *conf.Ops_Pprof) {
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
