package client

import (
	"log/slog"

	"github.com/go-kratos/kratos/contrib/otel/v3/tracing"
	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/middleware/circuitbreaker"
	"github.com/go-kratos/kratos/v3/middleware/logging"
	"github.com/go-kratos/kratos/v3/middleware/metadata"
	"github.com/google/wire"

	"cyber-ecosystem/shared-go/kratos/middleware/sanitize"
)

func standardMiddleware(logger *slog.Logger) []middleware.Middleware {
	return []middleware.Middleware{
		tracing.Client(),
		circuitbreaker.Client(),
		metadata.Client(),
		sanitize.Client(),
		logging.Client(logger),
	}
}

var ProviderSet = wire.NewSet(
	NewSystemClient,
	NewSystemConnectClient,
)
