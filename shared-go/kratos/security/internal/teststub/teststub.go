// Package teststub provides transport fakes shared across the security
// package family's test suites: one server Transporter stub and one header
// stub, so middleware tests only spell out the behavior under test.
package teststub

import (
	"net/http"

	"github.com/go-kratos/kratos/v3/transport"
)

// Transport embeds a nil Transporter and overrides exactly Operation and
// RequestHeader; calling any other method panics, so silent stub drift is
// impossible. Build it with NewTransport or NewTransportWithHeader.
type Transport struct {
	transport.Transporter
	op     string
	header transport.Header
}

func NewTransport(op string) Transport {
	return Transport{op: op}
}

func NewTransportWithHeader(op string, h http.Header) Transport {
	return Transport{op: op, header: Header(h)}
}

func (t Transport) Operation() string { return t.op }

func (t Transport) RequestHeader() transport.Header { return t.header }

// Header adapts net/http's header to kratos's transport header. Keys stays
// unimplemented (nil): no middleware under test reads it.
type Header http.Header

func (h Header) Get(key string) string      { return http.Header(h).Get(key) }
func (h Header) Set(key, value string)      { http.Header(h).Set(key, value) }
func (h Header) Add(key, value string)      { http.Header(h).Add(key, value) }
func (h Header) Keys() []string             { return nil }
func (h Header) Values(key string) []string { return http.Header(h).Values(key) }
