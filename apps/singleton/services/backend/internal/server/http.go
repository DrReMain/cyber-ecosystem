package server

import (
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/handlers"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/ratelimit"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	krahttp "github.com/go-kratos/kratos/v2/transport/http"

	"cyber-ecosystem/shared-go/kratos/middleware/auth"
	"cyber-ecosystem/shared-go/kratos/middleware/i18n"
	"cyber-ecosystem/shared-go/kratos/middleware/validate"

	"cyber-ecosystem/apps/singleton/services/backend/internal/conf"
	pkgauth "cyber-ecosystem/apps/singleton/services/backend/internal/pkg/auth"
	pkgdatascope "cyber-ecosystem/apps/singleton/services/backend/internal/pkg/datascope"
	pkgsecurity "cyber-ecosystem/apps/singleton/services/backend/internal/pkg/security"
	mw "cyber-ecosystem/apps/singleton/services/backend/internal/server/middleware"
	"cyber-ecosystem/apps/singleton/services/backend/internal/service"
)

func NewHTTPServer(
	c *conf.Server,
	ca *conf.Auth,
	cs *conf.Security,
	logger log.Logger,
	registrar []service.Registrar,
	tp *trace.TracerProvider,
	_metricRequests metric.Int64Counter,
	_metricSeconds metric.Float64Histogram,
	i18nBundle *i18n.Bundle,
	sessionValidator pkgsecurity.SessionValidator,
	authorizer pkgsecurity.Authorizer,
	conditionChecker pkgsecurity.ConditionChecker,
	scopeResolver pkgdatascope.ScopeResolveFunc,
) *krahttp.Server {
	var middlewares []middleware.Middleware
	middlewares = append(middlewares, i18n.Server(i18nBundle))
	middlewares = append(middlewares, recovery.Recovery())
	middlewares = append(middlewares, ratelimit.Server())
	middlewares = append(middlewares, metrics.Server(metrics.WithSeconds(_metricSeconds), metrics.WithRequests(_metricRequests)))
	if tp != nil {
		middlewares = append(middlewares, tracing.Server(tracing.WithTracerProvider(tp)))
	}
	middlewares = append(middlewares, metadata.Server())
	middlewares = append(middlewares, logging.Server(logger))
	middlewares = append(middlewares, selector.Server(
		jwt.Server(
			func(token *jwtv5.Token) (any, error) { return []byte(ca.Secret), nil },
			jwt.WithSigningMethod(pkgauth.SigningMethod),
			jwt.WithClaims(func() jwtv5.Claims { return &pkgauth.Identity{} }),
		),
		mw.SessionValidator(cs.GetSession(), logger, sessionValidator),
		mw.Authorizer(cs.GetAuthorization(), logger, authorizer),
		mw.ConditionChecker(cs.GetConditions(), logger, conditionChecker),
		mw.ScopeInjector(cs.GetDataScope(), scopeResolver),
	).Match(auth.NewWhiteListByPublicAccessInProtoMatcher()).Build())
	middlewares = append(middlewares, validate.ProtoValidate())

	var opts = []krahttp.ServerOption{
		krahttp.Middleware(middlewares...),
		krahttp.Filter(handlers.CORS(
			handlers.AllowedOrigins([]string{"*"}),
			handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}),
			handlers.AllowedHeaders([]string{"Content-Type", i18n.DefaultHeaderLang, "Authorization"}),
			handlers.MaxAge(86400),
		)),
	}
	if c.Http.Network != "" {
		opts = append(opts, krahttp.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, krahttp.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, krahttp.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := krahttp.NewServer(opts...)
	for _, r := range registrar {
		r.RegisterHTTP(srv)
	}
	return srv
}
