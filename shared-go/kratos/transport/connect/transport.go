package connect

import (
	"net/http"

	"github.com/go-kratos/kratos/v2/transport"
)

const KindConnect transport.Kind = "connect"

var _ transport.Transporter = (*Transport)(nil)

type Transport struct {
	endpoint    string
	operation   string
	reqHeader   transport.Header
	replyHeader transport.Header
}

func (tr *Transport) Kind() transport.Kind {
	return KindConnect
}

func (tr *Transport) Endpoint() string {
	return tr.endpoint
}

func (tr *Transport) Operation() string {
	return tr.operation
}

func (tr *Transport) RequestHeader() transport.Header {
	return tr.reqHeader
}

func (tr *Transport) ReplyHeader() transport.Header {
	return tr.replyHeader
}

type Header struct {
	http.Header
}

func (h Header) Get(key string) string {
	return h.Header.Get(key)
}

func (h Header) Set(key string, value string) {
	h.Header.Set(key, value)
}

func (h Header) Add(key string, value string) {
	h.Header.Add(key, value)
}

func (h Header) Keys() []string {
	keys := make([]string, 0, len(h.Header))
	for k := range h.Header {
		keys = append(keys, k)
	}
	return keys
}

func (h Header) Values(key string) []string {
	return h.Header.Values(key)
}

func NewHeader(h http.Header) Header {
	return Header{Header: h}
}
