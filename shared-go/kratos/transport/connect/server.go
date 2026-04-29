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
	"sync"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/grpchealth"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"

	"cyber-ecosystem/shared-go/kratos/transport/connect/internal/endpoint"
	"cyber-ecosystem/shared-go/kratos/transport/connect/internal/host"
	"cyber-ecosystem/shared-go/kratos/transport/connect/internal/matcher"
)

var errUnexpectedResponseType = errors.New("internal error: unexpected response type")

var (
	_ transport.Server     = (*Server)(nil)
	_ transport.Endpointer = (*Server)(nil)
)

type FilterFunc func(http.Handler) http.Handler

type Server struct {
	*http.Server

	mu                   sync.RWMutex
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
	health               *healthChecker
	disableReflection    bool
	reflectionServices   map[string]struct{}
	reflectionRegistered bool
	healthRegistered     bool
	enableH2C            bool
	errorEncoder         func(context.Context, error) error
}

type handlerEntry struct {
	path    string
	handler http.Handler
}

func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		baseCtx:            context.Background(),
		network:            "tcp",
		address:            ":0",
		timeout:            1 * time.Second,
		middleware:         matcher.New(),
		streamMiddleware:   matcher.New(),
		mux:                http.NewServeMux(),
		health:             newHealthChecker(),
		reflectionServices: make(map[string]struct{}),
		enableH2C:          true,
		errorEncoder:       func(_ context.Context, err error) error { return ErrorToConnect(err) },
	}
	srv.connectOpts = append(srv.connectOpts, connect.WithCodec(JSONCodec()))

	for _, o := range opts {
		o(srv)
	}

	handler := http.Handler(srv.mux)
	if len(srv.filters) > 0 {
		handler = filterChain(srv.filters...)(handler)
	}
	if srv.tlsConf == nil && srv.enableH2C {
		handler = h2c.NewHandler(handler, &http2.Server{})
	}

	srv.Server = &http.Server{
		Handler:   handler,
		TLSConfig: srv.tlsConf,
	}

	return srv
}

func (s *Server) Use(selector string, m ...middleware.Middleware) {
	s.middleware.Add(selector, m...)
}

func (s *Server) UseStream(selector string, m ...middleware.Middleware) {
	s.streamMiddleware.Add(selector, m...)
}

func (s *Server) Register(path string, handler http.Handler) {
	s.mux.Handle(path, handler)
	s.handlers = append(s.handlers, handlerEntry{path: path, handler: handler})
	if service, ok := inferServiceName(path); ok {
		s.reflectionServices[service] = struct{}{}
	}
}

func (s *Server) HandlerOptions() []connect.HandlerOption {
	opts := make([]connect.HandlerOption, 0, len(s.connectOpts)+2)
	opts = append(opts, s.connectOpts...)
	if len(s.interceptors) > 0 {
		opts = append(opts, connect.WithInterceptors(s.interceptors...))
	}
	opts = append(opts, connect.WithInterceptors(newKratosInterceptor(s)))
	return opts
}

func (s *Server) Endpoint() (*url.URL, error) {
	if err := s.listenAndEndpoint(); err != nil {
		return nil, s.err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.endpoint, nil
}

func (s *Server) Start(ctx context.Context) error {
	if err := s.listenAndEndpoint(); err != nil {
		return s.err
	}
	s.ensureHealthRegistered()
	s.ensureReflectionRegistered()

	s.baseCtx = ctx
	s.Server.BaseContext = func(net.Listener) context.Context {
		return ctx
	}

	log.Infof("[Connect] server listening on: %s", s.lis.Addr().String())

	s.health.resume()

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

func (s *Server) Stop(ctx context.Context) error {
	log.Info("[Connect] server stopping")

	s.health.shutdown()

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
	s.mu.Lock()
	defer s.mu.Unlock()

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

func (s *Server) listenerAddr() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.lis == nil {
		return ""
	}
	return s.lis.Addr().String()
}

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

// ensureHealthRegistered registers the grpc.health.v1.Health service and /healthz endpoint.
// It must be called before ensureReflectionRegistered so the health service appears in reflection.
func (s *Server) ensureHealthRegistered() {
	if s.healthRegistered {
		return
	}
	// Provide health checker with the set of registered services
	s.health.setServices(s.reflectionServices)

	// Standard gRPC health protocol (grpc.health.v1.Health/Check)
	grpcPath, grpcHandler := grpchealth.NewHandler(s.health)
	s.mux.Handle(grpcPath, grpcHandler)

	// Simple REST health endpoint for load balancers
	s.mux.HandleFunc("/healthz", s.health.healthzHandlerFunc())

	// Add health service to reflection list
	s.reflectionServices[grpchealth.HealthV1ServiceName] = struct{}{}

	s.healthRegistered = true
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

type kratosInterceptor struct {
	server *Server
}

func newKratosInterceptor(srv *Server) connect.Interceptor {
	return &kratosInterceptor{server: srv}
}

func (i *kratosInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		ctx, cancel := Merge(ctx, i.server.baseCtx)
		defer cancel()

		ep := ""
		if i.server.endpoint != nil {
			ep = i.server.endpoint.String()
		}

		tr := &Transport{
			endpoint:    ep,
			operation:   req.Spec().Procedure,
			reqHeader:   NewHeader(req.Header()),
			replyHeader: NewHeader(http.Header{}),
			httpMethod:  "POST",
			remoteAddr:  req.Peer().Addr,
		}
		ctx = transport.NewServerContext(ctx, tr)

		if i.server.timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, i.server.timeout)
			defer cancel()
		}

		h := func(ctx context.Context, _ any) (any, error) {
			return next(ctx, req)
		}
		if m := i.server.middleware.Match(tr.Operation()); len(m) > 0 {
			h = middleware.Chain(m...)(h)
		}

		resp, err := h(ctx, req.Any())
		if err != nil {
			encodedErr := i.server.errorEncoder(ctx, err)
			attachReplyHeadersToConnectError(tr, encodedErr)
			return nil, encodedErr
		}

		if cr, ok := resp.(connect.AnyResponse); ok {
			for _, k := range tr.replyHeader.Keys() {
				for _, v := range tr.replyHeader.Values(k) {
					cr.Header().Add(k, v)
				}
			}
			return cr, nil
		}
		return nil, connect.NewError(connect.CodeInternal, errUnexpectedResponseType)
	}
}

func (i *kratosInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		return next(ctx, spec)
	}
}

func (i *kratosInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		ctx, cancel := Merge(ctx, i.server.baseCtx)
		defer cancel()

		ep := ""
		if i.server.endpoint != nil {
			ep = i.server.endpoint.String()
		}

		tr := &Transport{
			endpoint:    ep,
			operation:   conn.Spec().Procedure,
			reqHeader:   NewHeader(conn.RequestHeader()),
			replyHeader: NewHeader(http.Header{}),
			httpMethod:  "POST",
			remoteAddr:  conn.Peer().Addr,
		}
		ctx = transport.NewServerContext(ctx, tr)

		h := func(ctx context.Context, _ any) (any, error) {
			return nil, next(ctx, conn)
		}
		if m := i.server.streamMiddleware.Match(tr.Operation()); len(m) > 0 {
			_, err := middleware.Chain(m...)(h)(ctx, conn)
			if err != nil {
				encodedErr := i.server.errorEncoder(ctx, err)
				attachReplyHeadersToConnectError(tr, encodedErr)
				return encodedErr
			}
			attachReplyHeadersToConn(tr, conn)
			return nil
		}
		if err := next(ctx, conn); err != nil {
			encodedErr := i.server.errorEncoder(ctx, err)
			attachReplyHeadersToConnectError(tr, encodedErr)
			return encodedErr
		}
		attachReplyHeadersToConn(tr, conn)
		return nil
	}
}

func attachReplyHeadersToConn(tr *Transport, conn connect.StreamingHandlerConn) {
	for _, key := range tr.replyHeader.Keys() {
		for _, value := range tr.replyHeader.Values(key) {
			conn.ResponseHeader().Add(key, value)
		}
	}
}

func attachReplyHeadersToConnectError(tr *Transport, err error) {
	ce, ok := err.(*connect.Error)
	if !ok || tr == nil {
		return
	}
	for _, key := range tr.replyHeader.Keys() {
		for _, value := range tr.replyHeader.Values(key) {
			ce.Meta().Add(key, value)
		}
	}
}
