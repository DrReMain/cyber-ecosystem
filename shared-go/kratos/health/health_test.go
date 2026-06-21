package health

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLivenessHandler_AlwaysOK(t *testing.T) {
	checker := NewChecker()
	req := httptest.NewRequest(http.MethodGet, HealthzPath, nil)

	// Liveness must stay 200 regardless of readiness: a draining process is
	// still alive and must not be restarted by the orchestrator.
	for _, ready := range []bool{false, true} {
		checker.SetReady(ready)
		rec := httptest.NewRecorder()
		checker.LivenessHandler().ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("ready=%v: want 200, got %d", ready, rec.Code)
		}
	}
}

func TestReadinessHandler_ReflectsFlag(t *testing.T) {
	checker := NewChecker()
	req := httptest.NewRequest(http.MethodGet, ReadyzPath, nil)

	checker.SetReady(false)
	rec := httptest.NewRecorder()
	checker.ReadinessHandler().ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("not ready: want 503, got %d", rec.Code)
	}

	checker.SetReady(true)
	rec = httptest.NewRecorder()
	checker.ReadinessHandler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("ready: want 200, got %d", rec.Code)
	}
}

func TestChecker_StartsNotReadyAndFlips(t *testing.T) {
	checker := NewChecker()
	if checker.Ready() {
		t.Fatal("a new checker must start not-ready")
	}
	checker.SetReady(true)
	if !checker.Ready() {
		t.Fatal("SetReady(true) was not reflected")
	}
	checker.SetReady(false)
	if checker.Ready() {
		t.Fatal("SetReady(false) was not reflected")
	}
}

func TestWithGracefulStopTimeout_OverridesDefault(t *testing.T) {
	cfg := mountConfig{drain: DefaultGracefulStopTimeout}
	WithGracefulStopTimeout(42 * time.Second).apply(&cfg)
	if cfg.drain != 42*time.Second {
		t.Fatalf("want 42s, got %v", cfg.drain)
	}
}
