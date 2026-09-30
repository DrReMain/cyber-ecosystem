// Package health exposes process-level liveness and readiness endpoints for
// orchestrators (Kubernetes probes, load balancers) and drives readiness from
// the application lifecycle.
//
// Readiness here is deliberately distinct from the "serving" health that the
// gRPC and Connect servers expose by default. Serving health reports whether a
// transport is up; this readiness reports whether the process is prepared to
// handle requests — it has finished starting up and is not draining. Keeping
// the two apart leaves the probe contract independent of whichever transports
// a service happens to run.
package health

import (
	"net/http"
	"sync/atomic"
)

// Checker holds the process readiness flag and renders it as HTTP handlers.
// One instance is shared between the readiness endpoint (which reads the flag)
// and the lifecycle hooks (which write it), so it must not be copied once in
// use.
type Checker struct {
	ready atomic.Bool
}

// NewChecker returns a Checker that starts in the not-ready state. The caller
// flips it to ready once startup completes and back to not-ready when shutdown
// begins.
func NewChecker() *Checker {
	return &Checker{}
}

// SetReady updates the readiness state. Intended to be called from application
// lifecycle hooks.
func (c *Checker) SetReady(ready bool) {
	c.ready.Store(ready)
}

// Ready reports whether the process is currently willing to accept traffic.
func (c *Checker) Ready() bool {
	return c.ready.Load()
}

// LivenessHandler always answers 200 OK. A liveness probe asks "is the process
// alive; should I restart it?", so the answer must stay independent of the
// readiness flag (a draining process is still alive and must not be killed
// mid-shutdown) and independent of dependencies (a flapping dependency would
// otherwise trigger a restart storm that helps nobody).
func (c *Checker) LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

// ReadinessHandler answers 200 when ready and 503 otherwise. A readiness probe
// asks "should I route traffic here?", and 503 takes the instance out of
// rotation without restarting it — exactly the desired behaviour during startup
// (not ready yet) and shutdown (deliberately draining).
func (c *Checker) ReadinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		if c.ready.Load() {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}
