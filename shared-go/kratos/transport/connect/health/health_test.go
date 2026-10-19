package health_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/kratos/transport/connect/health"
)

func TestHealthzNotServingThenServing(t *testing.T) {
	srv := connect.NewServer()
	srv.Register("/pkg.Svc/Method", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	ctl, err := health.Register(srv)
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(srv)
	defer ts.Close()

	get := func() int {
		resp, err := http.Get(ts.URL + "/healthz")
		if err != nil {
			t.Fatal(err)
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		return resp.StatusCode
	}

	if code := get(); code != http.StatusServiceUnavailable {
		t.Fatalf("before resume: status = %d, want 503", code)
	}
	ctl.Resume()
	if code := get(); code != http.StatusOK {
		t.Fatalf("after resume: status = %d, want 200", code)
	}
	ctl.Shutdown()
	if code := get(); code != http.StatusServiceUnavailable {
		t.Fatalf("after shutdown: status = %d, want 503", code)
	}
}
