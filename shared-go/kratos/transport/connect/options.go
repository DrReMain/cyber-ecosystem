package connect

import (
	"context"
	"crypto/tls"
	"net"
	"net/url"
	"time"

	"connectrpc.com/connect"
	"github.com/go-kratos/kratos/v3/middleware"
)

// ServerOption is a connect server option.
type ServerOption func(*Server)

// Network with server network.
func Network(network string) ServerOption {
	return func(s *Server) {
		s.network = network
	}
}

// Address with server address.
func Address(addr string) ServerOption {
	return func(s *Server) {
		s.address = addr
	}
}

// Endpoint with server endpoint.
func Endpoint(endpoint *url.URL) ServerOption {
	return func(s *Server) {
		s.endpoint = endpoint
	}
}

// Timeout with server timeout.
func Timeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.timeout = timeout
	}
}

// Middleware with server middleware.
func Middleware(m ...middleware.Middleware) ServerOption {
	return func(s *Server) {
		s.middleware.Use(m...)
	}
}

// StreamMiddleware with server stream middleware.
func StreamMiddleware(m ...middleware.Middleware) ServerOption {
	return func(s *Server) {
		s.streamMiddleware.Use(m...)
	}
}

// TLSConfig with server TLS config.
func TLSConfig(c *tls.Config) ServerOption {
	return func(s *Server) {
		s.tlsConf = c
	}
}

// Listener with server listener.
func Listener(lis net.Listener) ServerOption {
	return func(s *Server) {
		s.lis = lis
	}
}

// ConnectOptions with connect handler options.
func ConnectOptions(opts ...connect.HandlerOption) ServerOption {
	return func(s *Server) {
		s.connectOpts = append(s.connectOpts, opts...)
	}
}

// Interceptors with connect interceptors.
func Interceptors(interceptors ...connect.Interceptor) ServerOption {
	return func(s *Server) {
		s.interceptors = append(s.interceptors, interceptors...)
	}
}

// DisableH2C disables HTTP/2 cleartext (h2c) support.
func DisableH2C() ServerOption {
	return func(s *Server) {
		s.enableH2C = false
	}
}

// Filter with HTTP middleware option.
func Filter(filters ...FilterFunc) ServerOption {
	return func(s *Server) {
		s.filters = append(s.filters, filters...)
	}
}

// ErrorEncoder customizes connect error encoding.
func ErrorEncoder(encoder func(context.Context, error) error) ServerOption {
	return func(s *Server) {
		if encoder == nil {
			return
		}
		s.errorEncoder = encoder
	}
}
