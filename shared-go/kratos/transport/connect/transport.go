package connect

import (
	"net/http"

	"github.com/go-kratos/kratos/v3/transport"
)

// KindConnect is the Connect transport kind.
const KindConnect transport.Kind = "connect"

var _ transport.Transporter = (*Transport)(nil)

// Transport is a Connect transport context value.
type Transport struct {
	endpoint    string
	operation   string
	reqHeader   transport.Header
	replyHeader transport.Header
	httpMethod  string
	remoteAddr  string
}

func (tr *Transport) Kind() transport.Kind { return KindConnect }
func (tr *Transport) Endpoint() string     { return tr.endpoint }
func (tr *Transport) Operation() string    { return tr.operation }
func (tr *Transport) RequestHeader() transport.Header {
	if tr.reqHeader == nil {
		tr.reqHeader = headerCarrier(http.Header{})
	}
	return tr.reqHeader
}
func (tr *Transport) ReplyHeader() transport.Header {
	if tr.replyHeader == nil {
		tr.replyHeader = headerCarrier(http.Header{})
	}
	return tr.replyHeader
}
func (tr *Transport) HTTPMethod() string { return tr.httpMethod }
func (tr *Transport) RemoteAddr() string { return tr.remoteAddr }

// headerCarrier adapts http.Header to transport.Header (unexported, mirrors grpc/http).
type headerCarrier http.Header

func (h headerCarrier) Get(key string) string { return http.Header(h).Get(key) }
func (h headerCarrier) Set(key, value string) { http.Header(h).Set(key, value) }
func (h headerCarrier) Add(key, value string) { http.Header(h).Add(key, value) }
func (h headerCarrier) Keys() []string {
	out := make([]string, 0, len(h))
	for k := range http.Header(h) {
		out = append(out, k)
	}
	return out
}
func (h headerCarrier) Values(key string) []string { return http.Header(h).Values(key) }

// newHeader is the internal constructor used by interceptor/client.
func newHeader(h http.Header) transport.Header { return headerCarrier(h) }
