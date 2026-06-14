// Package health provides opt-in gRPC health + /healthz for a Connect server.
package health

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"

	connectrpc "connectrpc.com/connect"
	"connectrpc.com/grpchealth"
	"github.com/go-kratos/kratos/v3/encoding"

	"cyber-ecosystem/shared-go/kratos/transport/connect"
)

type checker struct {
	serving  atomic.Bool
	services map[string]struct{}
}

// Controller lets callers drive serving state explicitly (e.g. in tests).
type Controller struct{ c *checker }

// Resume marks the server as SERVING.
func (ctl *Controller) Resume() { ctl.c.serving.Store(true) }

// Shutdown marks the server as NOT_SERVING.
func (ctl *Controller) Shutdown() { ctl.c.serving.Store(false) }

// Register wires the grpc.health.v1 Health service and a /healthz endpoint onto
// srv. The known-service list is derived from srv.RegisteredServices(). The
// serving state is driven by srv's lifecycle (OnStart/OnStop); the returned
// Controller also allows explicit control.
func Register(srv *connect.Server) (*Controller, error) {
	c := &checker{services: make(map[string]struct{})}
	for _, s := range srv.RegisteredServices() {
		c.services[s] = struct{}{}
	}
	c.services[grpchealth.HealthV1ServiceName] = struct{}{}
	grpcPath, grpcHandler := grpchealth.NewHandler(c)
	srv.Register(grpcPath, grpcHandler)
	srv.Register("/healthz", http.HandlerFunc(c.healthz))
	srv.OnStart(c.resume)
	srv.OnStop(c.shutdown)
	return &Controller{c: c}, nil
}

func (c *checker) resume()   { c.serving.Store(true) }
func (c *checker) shutdown() { c.serving.Store(false) }

// Check implements grpchealth.Checker.
func (c *checker) Check(_ context.Context, req *grpchealth.CheckRequest) (*grpchealth.CheckResponse, error) {
	if req.Service != "" {
		if _, ok := c.services[req.Service]; !ok {
			return nil, connectrpc.NewError(connectrpc.CodeNotFound, fmt.Errorf("unknown service: %s", req.Service))
		}
	}
	if c.serving.Load() {
		return &grpchealth.CheckResponse{Status: grpchealth.StatusServing}, nil
	}
	return &grpchealth.CheckResponse{Status: grpchealth.StatusNotServing}, nil
}

func (c *checker) healthz(w http.ResponseWriter, _ *http.Request) {
	type response struct {
		Status string `json:"status"`
	}
	serving := c.serving.Load()
	status := "SERVING"
	if !serving {
		status = "NOT_SERVING"
	}
	w.Header().Set("Content-Type", "application/json")
	if !serving {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	codec := encoding.GetCodec("json")
	if codec == nil {
		http.Error(w, `{"status":"UNKNOWN"}`, http.StatusInternalServerError)
		return
	}
	data, err := codec.Marshal(response{Status: status})
	if err != nil {
		http.Error(w, `{"status":"UNKNOWN"}`, http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(data)
}
