package connect

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"

	connectrpc "connectrpc.com/connect"
	"connectrpc.com/grpchealth"

	"github.com/go-kratos/kratos/v2/encoding"
)

// healthChecker implements grpchealth.Checker and drives both the standard
// grpc.health.v1.Health protocol and the simple /healthz REST endpoint.
type healthChecker struct {
	serving  atomic.Bool
	services map[string]struct{}
}

func newHealthChecker() *healthChecker {
	return &healthChecker{
		services: make(map[string]struct{}),
	}
}

func (h *healthChecker) setServices(services map[string]struct{}) {
	h.services = services
}

func (h *healthChecker) resume() {
	h.serving.Store(true)
}

func (h *healthChecker) shutdown() {
	h.serving.Store(false)
}

// Check implements grpchealth.Checker.
// An empty service name checks the overall server health.
func (h *healthChecker) Check(_ context.Context, req *grpchealth.CheckRequest) (*grpchealth.CheckResponse, error) {
	if req.Service != "" {
		if _, ok := h.services[req.Service]; !ok {
			return nil, connectrpc.NewError(connectrpc.CodeNotFound,
				fmt.Errorf("unknown service: %s", req.Service))
		}
	}
	if h.serving.Load() {
		return &grpchealth.CheckResponse{Status: grpchealth.StatusServing}, nil
	}
	return &grpchealth.CheckResponse{Status: grpchealth.StatusNotServing}, nil
}

// healthzHandlerFunc returns a simple REST handler for load balancers and probes.
func (h *healthChecker) healthzHandlerFunc() http.HandlerFunc {
	type response struct {
		Status string `json:"status"`
	}
	return func(w http.ResponseWriter, _ *http.Request) {
		status := "SERVING"
		if !h.serving.Load() {
			status = "NOT_SERVING"
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		codec := encoding.GetCodec("json")
		data, _ := codec.Marshal(response{Status: status})
		w.Header().Set("Content-Type", "application/json")
		w.Write(data) //nolint:errcheck
	}
}
