package connect

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"

	"github.com/DrReMain/cyber-ecosystem/shared-go/kratos/transport/connect/internal/endpoint"
	"github.com/DrReMain/cyber-ecosystem/shared-go/kratos/transport/connect/internal/host"
	"github.com/DrReMain/cyber-ecosystem/shared-go/kratos/transport/connect/internal/matcher"
)

var errUnexpectedResponseType = errors.New("internal error: unexpected response type")

var (
	_ transport.Server     = (*Server)(nil)
	_ transport.Endpointer = (*Server)(nil)
)

// FilterFunc is an HTTP middleware function.
type FilterFunc func(http.Handler) http.Handler

// Server is a connect server wrapper.
type Server struct {
	*http.Server

	baseCtx              context.Context
	tlsConf              *tls.Config
	lis                  net.Listener
	err                  error
	network              string
	address              string
	endpoint             *url.URL
	timeout              time.Duration
	middleware           matcher.Matcher
	streamMiddleware     matcher.Matcher
	connectOpts          []connect.HandlerOption
	interceptors         []connect.Interceptor
	filters              []FilterFunc
	mux                  *http.ServeMux
	handlers             []handlerEntry
	health               *healthServer
	disableReflection    bool
	reflectionServices   map[string]struct{}
	reflectionRegistered bool
	enableH2C            bool
}

type handlerEntry struct {
	path    string
	handler http.Handler
}

// NewServer creates a connect server by options.
func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		baseCtx:            context.Background(),
		network:            "tcp",
		address:            ":0",
		timeout:            1 * time.Second,
		middleware:         matcher.New(),
		streamMiddleware:   matcher.New(),
		mux:                http.NewServeMux(),
		health:             newHealthServer(),
		reflectionServices: make(map[string]struct{}),
		enableH2C:          true,
	}
	srv.connectOpts = append(srv.connectOpts, connect.WithCodec(JSONCodec()))

	for _, o := range opts {
		o(srv)
	}

	// Create handler chain with filters
	handler := http.Handler(srv.mux)
	if len(srv.filters) > 0 {
		handler = filterChain(srv.filters...)(handler)
	}
	if srv.tlsConf == nil && srv.enableH2C {
		// Reflection and gRPC protocol over cleartext require HTTP/2 (h2c).
		handler = h2c.NewHandler(handler, &http2.Server{})
	}

	srv.Server = &http.Server{
		Handler:   handler,
		TLSConfig: srv.tlsConf,
	}

	// Register health check endpoint
	srv.mux.HandleFunc("/healthz", srv.health.HandlerFunc())

	return srv
}

// Use uses a service middleware with selector.
func (s *Server) Use(selector string, m ...middleware.Middleware) {
	s.middleware.Add(selector, m...)
}

// UseStream uses a stream middleware with selector.
func (s *Server) UseStream(selector string, m ...middleware.Middleware) {
	s.streamMiddleware.Add(selector, m...)
}

// Register registers a connect handler.
func (s *Server) Register(path string, handler http.Handler) {
	s.mux.Handle(path, handler)
	s.handlers = append(s.handlers, handlerEntry{path: path, handler: handler})
	if service, ok := inferServiceName(path); ok {
		s.reflectionServices[service] = struct{}{}
	}
}

// HandlerOptions returns connect.HandlerOption slice with interceptors for creating handlers.
// This should be used when creating connect handlers to ensure middleware is applied.
func (s *Server) HandlerOptions() []connect.HandlerOption {
	opts := make([]connect.HandlerOption, 0, len(s.connectOpts)+2)
	opts = append(opts, s.connectOpts...)
	if len(s.interceptors) > 0 {
		opts = append(opts, connect.WithInterceptors(s.interceptors...))
	}
	// Add kratos interceptor for middleware support
	opts = append(opts, connect.WithInterceptors(newKratosInterceptor(s)))
	return opts
}

// Endpoint returns the server endpoint.
func (s *Server) Endpoint() (*url.URL, error) {
	if err := s.listenAndEndpoint(); err != nil {
		return nil, s.err
	}
	return s.endpoint, nil
}

// Start starts the connect server.
func (s *Server) Start(ctx context.Context) error {
	if err := s.listenAndEndpoint(); err != nil {
		return s.err
	}
	s.ensureReflectionRegistered()

	s.baseCtx = ctx
	s.Server.BaseContext = func(net.Listener) context.Context {
		return ctx
	}

	log.Infof("[Connect] server listening on: %s", s.lis.Addr().String())

	// Mark health as serving
	s.health.Resume()

	var err error
	if s.tlsConf != nil {
		err = s.Server.ServeTLS(s.lis, "", "")
	} else {
		err = s.Server.Serve(s.lis)
	}

	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Stop stops the connect server.
func (s *Server) Stop(ctx context.Context) error {
	log.Info("[Connect] server stopping")

	// Mark health as not serving
	s.health.Shutdown()

	err := s.Server.Shutdown(ctx)
	if err != nil {
		if ctx.Err() != nil {
			log.Warn("[Connect] server couldn't stop gracefully in time, doing force stop")
			err = s.Server.Close()
		}
	}
	return err
}

func (s *Server) listenAndEndpoint() error {
	if s.lis == nil {
		lis, err := net.Listen(s.network, s.address)
		if err != nil {
			s.err = err
			return err
		}
		s.lis = lis
	}
	if s.endpoint == nil {
		addr, err := host.Extract(s.address, s.lis)
		if err != nil {
			s.err = err
			return err
		}
		s.endpoint = endpoint.NewEndpoint(endpoint.Scheme("connect", s.tlsConf != nil), addr)
	}
	return s.err
}

// filterChain returns a FilterFunc that chains multiple filters.
func filterChain(filters ...FilterFunc) FilterFunc {
	return func(final http.Handler) http.Handler {
		for i := len(filters) - 1; i >= 0; i-- {
			final = filters[i](final)
		}
		return final
	}
}

func inferServiceName(path string) (string, bool) {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return "", false
	}
	parts := strings.Split(trimmed, "/")
	if len(parts) == 0 {
		return "", false
	}
	service := parts[0]
	if service == "" || !strings.Contains(service, ".") {
		return "", false
	}
	return service, true
}

func (s *Server) ensureReflectionRegistered() {
	if s.disableReflection || s.reflectionRegistered {
		return
	}
	services := make([]string, 0, len(s.reflectionServices))
	for service := range s.reflectionServices {
		services = append(services, service)
	}
	if len(services) == 0 {
		return
	}
	sort.Strings(services)
	for _, entry := range ReflectionHandlers(services...) {
		if s.hasHandler(entry.Path) {
			continue
		}
		s.mux.Handle(entry.Path, entry.Handler)
		s.handlers = append(s.handlers, handlerEntry{path: entry.Path, handler: entry.Handler})
	}
	s.reflectionRegistered = true
}

func (s *Server) hasHandler(path string) bool {
	for _, entry := range s.handlers {
		if entry.path == path {
			return true
		}
	}
	return false
}

// kratosInterceptor adapts Kratos middleware to Connect interceptor.
type kratosInterceptor struct {
	server *Server
}

// newKratosInterceptor creates a new Kratos middleware adapter.
func newKratosInterceptor(srv *Server) connect.Interceptor {
	return &kratosInterceptor{server: srv}
}

// WrapUnary wraps a unary handler with Kratos middleware.
func (i *kratosInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		// Merge context
		ctx, cancel := Merge(ctx, i.server.baseCtx)
		defer cancel()

		endpoint := ""
		if i.server.endpoint != nil {
			endpoint = i.server.endpoint.String()
		}

		// Create Transport
		tr := &Transport{
			endpoint:    endpoint,
			operation:   req.Spec().Procedure,
			reqHeader:   NewHeader(req.Header()),
			replyHeader: NewHeader(http.Header{}),
		}
		ctx = transport.NewServerContext(ctx, tr)

		// Set timeout
		if i.server.timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, i.server.timeout)
			defer cancel()
		}

		// Apply Kratos middleware
		h := func(ctx context.Context, _ any) (any, error) {
			return next(ctx, req)
		}

		if m := i.server.middleware.Match(tr.Operation()); len(m) > 0 {
			h = middleware.Chain(m...)(h)
		}

		// Execute
		resp, err := h(ctx, req.Any())
		if err != nil {
			return nil, ErrorToConnect(err)
		}

		// Handle response headers
		if resp != nil {
			if cr, ok := resp.(connect.AnyResponse); ok {
				for _, k := range tr.replyHeader.Keys() {
					for _, v := range tr.replyHeader.Values(k) {
						cr.Header().Add(k, v)
					}
				}
			}
		}

		// Type assert back to AnyResponse
		if cr, ok := resp.(connect.AnyResponse); ok {
			return cr, nil
		}
		return nil, connect.NewError(connect.CodeInternal, errUnexpectedResponseType)
	}
}

// WrapStreamingClient wraps a streaming client with Kratos middleware.
// This is used for client-side streaming, not server handlers.
func (i *kratosInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		return next(ctx, spec)
	}
}

// WrapStreamingHandler wraps a streaming handler with Kratos middleware.
func (i *kratosInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		// Merge context
		ctx, cancel := Merge(ctx, i.server.baseCtx)
		defer cancel()

		endpoint := ""
		if i.server.endpoint != nil {
			endpoint = i.server.endpoint.String()
		}

		// Create Transport
		tr := &Transport{
			endpoint:    endpoint,
			operation:   conn.Spec().Procedure,
			reqHeader:   NewHeader(conn.RequestHeader()),
			replyHeader: NewHeader(http.Header{}),
		}
		ctx = transport.NewServerContext(ctx, tr)

		// Apply Kratos middleware
		h := func(ctx context.Context, _ any) (any, error) {
			return nil, next(ctx, conn)
		}

		if m := i.server.streamMiddleware.Match(tr.Operation()); len(m) > 0 {
			_, err := middleware.Chain(m...)(h)(ctx, conn)
			return err
		}

		return next(ctx, conn)
	}
}
