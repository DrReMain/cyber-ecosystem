package health

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v3"
	kratoshttp "github.com/go-kratos/kratos/v3/transport/http"
)

// HTTP paths where the probes are served. Exported so deployment configs (for
// example Kubernetes probe definitions) can reference the exact same values
// rather than re-hardcoding them.
const (
	HealthzPath = "/healthz"
	ReadyzPath  = "/readyz"
)

// DefaultGracefulStopTimeout is the drain window used when a service does not
// override it. It bounds how long in-flight requests are allowed to finish
// during shutdown. 15s is generous for fast request APIs; services with
// long-running handlers pass a larger value via WithGracefulStopTimeout.
const DefaultGracefulStopTimeout = 15 * time.Second

// Mount registers the liveness and readiness endpoints on the HTTP server and
// returns kratos options that drive readiness from the application lifecycle:
// the instance becomes ready once startup finishes and stops being ready before
// shutdown drains. Adding the returned options to kratos.New is the only step a
// service takes to opt into probes.
//
// Probes are hosted on the plain HTTP server rather than on gRPC or Connect so
// the probe contract stays decoupled from whichever transports a service runs.
func Mount(hs *kratoshttp.Server, opts ...MountOption) []kratos.Option {
	cfg := mountConfig{drain: DefaultGracefulStopTimeout}
	for _, opt := range opts {
		opt.apply(&cfg)
	}

	checker := NewChecker()
	hs.HandleFunc(HealthzPath, checker.LivenessHandler())
	hs.HandleFunc(ReadyzPath, checker.ReadinessHandler())

	return []kratos.Option{
		kratos.AfterStart(func(context.Context) error {
			checker.SetReady(true)
			return nil
		}),
		kratos.BeforeStop(func(context.Context) error {
			checker.SetReady(false)
			return nil
		}),
		kratos.StopTimeout(cfg.drain),
	}
}

// MountOption configures Mount. Construct one with the With* helpers.
type MountOption interface {
	apply(*mountConfig)
}

type mountConfig struct {
	drain time.Duration
}

// drainOption overrides the graceful-stop drain window.
type drainOption time.Duration

func (d drainOption) apply(cfg *mountConfig) { cfg.drain = time.Duration(d) }

// WithGracefulStopTimeout overrides the drain window passed to kratos
// StopTimeout. Reach for it when a service has handlers that may legitimately
// run longer than DefaultGracefulStopTimeout.
func WithGracefulStopTimeout(d time.Duration) MountOption {
	return drainOption(d)
}
