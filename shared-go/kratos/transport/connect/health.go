package connect

import (
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/go-kratos/kratos/v2/encoding"
)

// healthServer implements a simple health check server.
type healthServer struct {
	mu      sync.Mutex
	serving atomic.Bool
}

// newHealthServer creates a new health server.
func newHealthServer() *healthServer {
	h := &healthServer{}
	h.serving.Store(false)
	return h
}

// Resume marks the server as serving.
func (h *healthServer) Resume() {
	h.serving.Store(true)
}

// Shutdown marks the server as not serving.
func (h *healthServer) Shutdown() {
	h.serving.Store(false)
}

// HealthResponse is the response for health check.
type HealthResponse struct {
	Status string `json:"status"`
}

// HandlerFunc returns a handler function for health check endpoint.
// It uses the globally registered JSON codec for consistent serialization.
func (h *healthServer) HandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := "SERVING"
		if !h.serving.Load() {
			status = "NOT_SERVING"
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		resp := HealthResponse{Status: status}

		// Use the globally registered JSON codec
		codec := encoding.GetCodec("json")
		data, _ := codec.Marshal(resp)

		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}
}
