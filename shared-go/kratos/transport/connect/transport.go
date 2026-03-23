package connect

import (
	"net/http"

	"github.com/go-kratos/kratos/v2/transport"
)

// Kind is the connect transport kind.
const KindConnect transport.Kind = "connect"

var _ transport.Transporter = (*Transport)(nil)

// Transport is a connect transport.
type Transport struct {
	endpoint    string
	operation   string
	reqHeader   transport.Header
	replyHeader transport.Header
}

// Kind returns the transport kind.
func (tr *Transport) Kind() transport.Kind {
	return KindConnect
}

// Endpoint returns the transport endpoint.
func (tr *Transport) Endpoint() string {
	return tr.endpoint
}

// Operation returns the transport operation.
func (tr *Transport) Operation() string {
	return tr.operation
}

// RequestHeader returns the request header.
func (tr *Transport) RequestHeader() transport.Header {
	return tr.reqHeader
}

// ReplyHeader returns the reply header.
func (tr *Transport) ReplyHeader() transport.Header {
	return tr.replyHeader
}

// Header implements transport.Header interface.
type Header struct {
	http.Header
}

// Get returns the value associated with the passed key.
func (h Header) Get(key string) string {
	return h.Header.Get(key)
}

// Set sets the header entries associated with key to the single element value.
func (h Header) Set(key string, value string) {
	h.Header.Set(key, value)
}

// Add adds the key, value pair to the header.
func (h Header) Add(key string, value string) {
	h.Header.Add(key, value)
}

// Keys returns the keys of the header.
func (h Header) Keys() []string {
	keys := make([]string, 0, len(h.Header))
	for k := range h.Header {
		keys = append(keys, k)
	}
	return keys
}

// Values returns all values associated with the given key.
func (h Header) Values(key string) []string {
	return h.Header.Values(key)
}

// NewHeader creates a new Header from http.Header.
func NewHeader(h http.Header) Header {
	return Header{Header: h}
}
