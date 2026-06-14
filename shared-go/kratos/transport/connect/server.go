package connect

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/go-kratos/kratos/v3/log"
	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/transport"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"cyber-ecosystem/shared-go/kratos/transport/connect/internal/endpoint"
	"cyber-ecosystem/shared-go/kratos/transport/connect/internal/host"
	"cyber-ecosystem/shared-go/kratos/transport/connect/internal/matcher"
)

var (
	_ transport.Server     = (*Server)(nil)
	_ transport.Endpointer = (*Server)(nil)
)

// FilterFunc is an HTTP middleware function.
type FilterFunc func(http.Handler) http.Handler

// Server is a Connect transport server.
type Server struct {
	*http.Server

	mu               sync.RWMutex
	baseCtx          context.Context
	tlsConf          *tls.Config
	lis              net.Listener
	err              error
	network          string
	address          string
	endpoint         *url.URL
	timeout          time.Duration
	middleware       matcher.Matcher
	streamMiddleware matcher.Matcher
	connectOpts      []connect.HandlerOption
	interceptors     []connect.Interceptor
	filters          []FilterFunc
	mux              *http.ServeMux
	handlers         []handlerEntry
	enableH2C        bool
	errorEncoder     func(context.Context, error) error
	onStart          []func()
	onStop           []func()
}

type handlerEntry struct {
	path    string
	handler http.Handler
}

// NewServer creates a Connect transport server.
func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		baseCtx:          context.Background(),
		network:          "tcp",
		address:          ":0",
		timeout:          1 * time.Second,
		middleware:       matcher.New(),
		streamMiddleware: matcher.New(),
		mux:              http.NewServeMux(),
		enableH2C:        true,
		errorEncoder:     func(_ context.Context, err error) error { return ErrorToConnect(err) },
	}

	// F2: do NOT force a single codec here. Leaving connectOpts empty lets
	// connect-go's HandlerOptions apply its NATIVE codec set (proto + json),
	// so the server accepts both application/json and application/proto
	// clients, selected by Content-Type.
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

// Use registers unary middleware for the server.
func (s *Server) Use(selector string, m ...middleware.Middleware) {
	s.middleware.Add(selector, m...)
}

// UseStream registers streaming middleware for the server.
func (s *Server) UseStream(selector string, m ...middleware.Middleware) {
	s.streamMiddleware.Add(selector, m...)
}

// streamMatcher returns the streaming-middleware matcher (used by adapter Handle*).
func (s *Server) streamMatcher() matcher.Matcher { return s.streamMiddleware }

// Register registers a handler at the given path.
func (s *Server) Register(path string, handler http.Handler) {
	s.mux.Handle(path, handler)
	s.handlers = append(s.handlers, handlerEntry{path: path, handler: handler})
}

// HandlerOptions returns connect.HandlerOption list for service registration.
func (s *Server) HandlerOptions() []connect.HandlerOption {
	opts := []connect.HandlerOption{
		connect.WithInterceptors(newKratosInterceptor(s)), // outermost: injects Transport first
	}
	opts = append(opts, s.connectOpts...)
	if len(s.interceptors) > 0 {
		opts = append(opts, connect.WithInterceptors(s.interceptors...))
	}
	return opts
}

// Endpoint returns the server endpoint.
func (s *Server) Endpoint() (*url.URL, error) {
	if err := s.listenAndEndpoint(); err != nil {
		return nil, s.err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.endpoint, nil
}

// RegisteredServices returns the distinct fully-qualified service names of all
// registered handlers (paths of the form "/pkg.Svc/Method" → "pkg.Svc").
func (s *Server) RegisteredServices() []string {
	seen := make(map[string]struct{})
	out := make([]string, 0)
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, h := range s.handlers {
		trimmed := strings.Trim(h.path, "/")
		parts := strings.SplitN(trimmed, "/", 2)
		if len(parts) == 0 || parts[0] == "" || !strings.Contains(parts[0], ".") {
			continue
		}
		if _, ok := seen[parts[0]]; ok {
			continue
		}
		seen[parts[0]] = struct{}{}
		out = append(out, parts[0])
	}
	return out
}

// OnStart registers a function invoked when the server starts (after listen, before Serve).
func (s *Server) OnStart(fn func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onStart = append(s.onStart, fn)
}

// OnStop registers a function invoked when the server stops (before Shutdown).
func (s *Server) OnStop(fn func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onStop = append(s.onStop, fn)
}

// ServeHTTP dispatches to the underlying handler chain.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Handler.ServeHTTP(w, r)
}

// Start starts the Connect server.
func (s *Server) Start(ctx context.Context) error {
	if err := s.listenAndEndpoint(); err != nil {
		return s.err
	}

	s.baseCtx = ctx
	s.Server.BaseContext = func(net.Listener) context.Context {
		return ctx
	}

	log.Info("[Connect] server listening", "addr", s.lis.Addr().String())

	s.mu.RLock()
	hooks := append([]func(){}, s.onStart...)
	s.mu.RUnlock()
	for _, fn := range hooks {
		fn()
	}

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

// Stop stops the Connect server.
func (s *Server) Stop(ctx context.Context) error {
	log.Info("[Connect] server stopping")

	s.mu.RLock()
	hooks := append([]func(){}, s.onStop...)
	s.mu.RUnlock()
	for _, fn := range hooks {
		fn()
	}

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

func filterChain(filters ...FilterFunc) FilterFunc {
	return func(final http.Handler) http.Handler {
		for i := len(filters) - 1; i >= 0; i-- {
			final = filters[i](final)
		}
		return final
	}
}
