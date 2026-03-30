package connect

import (
	"context"
	"crypto/tls"
	stderrors "errors"
	"io"
	"net"
	"net/http"
	"time"

	connectrpc "connectrpc.com/connect"
	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/selector"
	"github.com/go-kratos/kratos/v2/selector/wrr"
	"github.com/go-kratos/kratos/v2/transport"
	"golang.org/x/net/http2"
)

func init() {
	if selector.GlobalSelector() == nil {
		selector.SetGlobalSelector(wrr.NewBuilder())
	}
}

// ClientOption is connect client option.
type ClientOption func(*clientOptions)

type clientOptions struct {
	endpoint     string
	timeout      time.Duration
	tlsConf      *tls.Config
	middleware   []middleware.Middleware
	streamMw     []middleware.Middleware
	transport    http.RoundTripper
	interceptors []connectrpc.Interceptor
	clientOpts   []connectrpc.ClientOption
	discovery    registry.Discovery
	nodeFilters  []selector.NodeFilter
	block        bool
	subsetSize   int
	h2c          bool
}

// WithEndpoint with client endpoint.
func WithEndpoint(endpoint string) ClientOption {
	return func(o *clientOptions) {
		o.endpoint = endpoint
	}
}

// WithTimeout with client timeout.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(o *clientOptions) {
		o.timeout = timeout
	}
}

// WithMiddleware with client middleware.
func WithMiddleware(m ...middleware.Middleware) ClientOption {
	return func(o *clientOptions) {
		o.middleware = m
	}
}

// WithStreamMiddleware with client stream middleware.
func WithStreamMiddleware(m ...middleware.Middleware) ClientOption {
	return func(o *clientOptions) {
		o.streamMw = m
	}
}

// WithTLSConfig with TLS config.
func WithTLSConfig(c *tls.Config) ClientOption {
	return func(o *clientOptions) {
		o.tlsConf = c
	}
}

// WithTransport with custom HTTP round tripper.
func WithTransport(rt http.RoundTripper) ClientOption {
	return func(o *clientOptions) {
		o.transport = rt
	}
}

// WithInterceptors with connect client interceptors.
func WithInterceptors(interceptors ...connectrpc.Interceptor) ClientOption {
	return func(o *clientOptions) {
		o.interceptors = append(o.interceptors, interceptors...)
	}
}

// WithClientOptions with connect client options.
func WithClientOptions(opts ...connectrpc.ClientOption) ClientOption {
	return func(o *clientOptions) {
		o.clientOpts = append(o.clientOpts, opts...)
	}
}

// WithDiscovery with client discovery.
func WithDiscovery(d registry.Discovery) ClientOption {
	return func(o *clientOptions) {
		o.discovery = d
	}
}

// WithNodeFilter with select filters.
func WithNodeFilter(filters ...selector.NodeFilter) ClientOption {
	return func(o *clientOptions) {
		o.nodeFilters = filters
	}
}

// WithBlock blocks until resolver receives initial service list.
func WithBlock() ClientOption {
	return func(o *clientOptions) {
		o.block = true
	}
}

// WithSubset with client discovery subset size.
// zero value means subset filter disabled.
func WithSubset(size int) ClientOption {
	return func(o *clientOptions) {
		o.subsetSize = size
	}
}

// WithH2C enables/disables HTTP/2 cleartext support for insecure client transport.
func WithH2C(enabled bool) ClientOption {
	return func(o *clientOptions) {
		o.h2c = enabled
	}
}

// Client is a connect transport client.
type Client struct {
	httpClient *http.Client
	endpoint   string
	clientOpts []connectrpc.ClientOption
	timeout    time.Duration
	middleware []middleware.Middleware
	resolver   *resolver
	selector   selector.Selector
}

// Dial creates a connect client.
func Dial(_ context.Context, opts ...ClientOption) (*Client, error) {
	return dial(false, opts...)
}

// DialInsecure creates a connect client over cleartext HTTP.
func DialInsecure(_ context.Context, opts ...ClientOption) (*Client, error) {
	return dial(true, opts...)
}

func dial(insecure bool, opts ...ClientOption) (*Client, error) {
	options := clientOptions{
		timeout:    2 * time.Second,
		transport:  nil,
		subsetSize: 25,
		h2c:        false,
	}
	for _, o := range opts {
		o(&options)
	}
	if options.endpoint == "" {
		return nil, stderrors.New("connect client endpoint is required")
	}
	isInsecure := insecure || options.tlsConf == nil
	target, err := parseTarget(options.endpoint, isInsecure)
	if err != nil {
		return nil, err
	}
	endpoint := target.baseURL()

	rt := options.transport
	if rt == nil {
		rt = defaultRoundTripper(isInsecure, options.tlsConf, options.h2c)
	}
	if tr, ok := rt.(*http.Transport); ok {
		cloned := tr.Clone()
		if options.tlsConf != nil {
			cloned.TLSClientConfig = options.tlsConf
		}
		rt = cloned
	}

	sel := selector.GlobalSelector().Build()
	var r *resolver
	if options.discovery != nil && target.Scheme == "discovery" {
		r, err = newResolver(context.Background(), options.discovery, target, sel, options.block, isInsecure, options.subsetSize)
		if err != nil {
			return nil, err
		}
	}

	wrapped := &clientRoundTripper{
		next:        rt,
		endpoint:    endpoint,
		timeout:     options.timeout,
		resolver:    r,
		selector:    sel,
		nodeFilters: options.nodeFilters,
		insecure:    isInsecure,
	}
	clientOpts := make([]connectrpc.ClientOption, 0, len(options.clientOpts)+2)
	clientOpts = append(clientOpts, options.clientOpts...)
	clientOpts = append(clientOpts, connectrpc.WithInterceptors(
		&unaryClientInterceptor{
			endpoint: endpoint,
			timeout:  options.timeout,
			mw:       options.middleware,
		},
		&streamClientInterceptor{
			endpoint: endpoint,
			mw:       options.streamMw,
		},
	))
	if len(options.interceptors) > 0 {
		clientOpts = append(clientOpts, connectrpc.WithInterceptors(options.interceptors...))
	}
	return &Client{
		httpClient: &http.Client{
			Timeout:   options.timeout,
			Transport: wrapped,
		},
		endpoint:   endpoint,
		clientOpts: clientOpts,
		timeout:    options.timeout,
		middleware: options.middleware,
		resolver:   r,
		selector:   sel,
	}, nil
}

// HTTPClient returns underlying HTTP client.
func (c *Client) HTTPClient() connectrpc.HTTPClient {
	return c.httpClient
}

// Endpoint returns normalized endpoint.
func (c *Client) Endpoint() string {
	return c.endpoint
}

// ClientOptions returns connect client options.
func (c *Client) ClientOptions() []connectrpc.ClientOption {
	opts := make([]connectrpc.ClientOption, 0, len(c.clientOpts))
	opts = append(opts, c.clientOpts...)
	return opts
}

// Close closes client resources.
func (c *Client) Close() error {
	if c.httpClient == nil {
		return nil
	}
	if c.resolver != nil {
		_ = c.resolver.Close()
	}
	if tr, ok := c.httpClient.Transport.(interface{ CloseIdleConnections() }); ok {
		tr.CloseIdleConnections()
	}
	return nil
}

type clientRoundTripper struct {
	next        http.RoundTripper
	endpoint    string
	timeout     time.Duration
	resolver    *resolver
	selector    selector.Selector
	nodeFilters []selector.NodeFilter
	insecure    bool
}

func (rt *clientRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	baseCtx := req.Context()
	ctx := baseCtx
	var cancel context.CancelFunc
	if rt.timeout > 0 {
		ctx, cancel = context.WithTimeout(baseCtx, rt.timeout)
	}

	r := req.Clone(ctx)
	var done selector.DoneFunc
	if rt.resolver != nil {
		node, doneFn, err := rt.selector.Select(ctx, selector.WithNodeFilter(rt.nodeFilters...))
		if err != nil {
			if cancel != nil {
				cancel()
			}
			return nil, kerrors.ServiceUnavailable("NODE_NOT_FOUND", err.Error())
		}
		done = doneFn
		r.URL.Host = node.Address()
		r.Host = node.Address()
		if rt.insecure {
			r.URL.Scheme = "http"
		} else {
			r.URL.Scheme = "https"
		}
	}
	resp, err := rt.next.RoundTrip(r)
	if done != nil {
		bytesReceived := false
		if resp != nil && resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
			bytesReceived = true
		}
		done(ctx, selector.DoneInfo{
			Err:           err,
			BytesSent:     true,
			BytesReceived: bytesReceived,
		})
	}
	if err != nil {
		if cancel != nil {
			cancel()
		}
		return nil, err
	}
	if cancel != nil {
		resp.Body = &cancelOnCloseReadCloser{ReadCloser: resp.Body, cancel: cancel}
	}
	return resp, nil
}

func defaultRoundTripper(insecure bool, tlsConf *tls.Config, enableH2C bool) http.RoundTripper {
	if insecure && enableH2C {
		return &http2.Transport{
			AllowHTTP: true,
			DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
				var d net.Dialer
				return d.DialContext(ctx, network, addr)
			},
		}
	}
	tr := &http.Transport{}
	if tlsConf != nil {
		tr.TLSClientConfig = tlsConf
	}
	return tr
}

type unaryClientInterceptor struct {
	endpoint string
	timeout  time.Duration
	mw       []middleware.Middleware
}

func (i *unaryClientInterceptor) WrapUnary(next connectrpc.UnaryFunc) connectrpc.UnaryFunc {
	return func(ctx context.Context, req connectrpc.AnyRequest) (connectrpc.AnyResponse, error) {
		tr := &Transport{
			endpoint:    i.endpoint,
			operation:   req.Spec().Procedure,
			reqHeader:   NewHeader(req.Header()),
			replyHeader: NewHeader(http.Header{}),
		}
		ctx = transport.NewClientContext(ctx, tr)

		if i.timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, i.timeout)
			defer cancel()
		}

		h := func(ctx context.Context, _ any) (any, error) {
			resp, err := next(ctx, req)
			if resp != nil {
				tr.replyHeader = NewHeader(resp.Header())
			}
			return resp, err
		}
		if len(i.mw) > 0 {
			h = middleware.Chain(i.mw...)(h)
		}
		respAny, err := h(ctx, req.Any())
		if err != nil {
			return nil, ErrorToConnect(err)
		}
		resp, ok := respAny.(connectrpc.AnyResponse)
		if !ok {
			return nil, connectrpc.NewError(connectrpc.CodeInternal, stderrors.New("invalid unary response type"))
		}
		return resp, nil
	}
}

func (i *unaryClientInterceptor) WrapStreamingClient(next connectrpc.StreamingClientFunc) connectrpc.StreamingClientFunc {
	return next
}

func (i *unaryClientInterceptor) WrapStreamingHandler(next connectrpc.StreamingHandlerFunc) connectrpc.StreamingHandlerFunc {
	return next
}

type streamClientInterceptor struct {
	endpoint string
	mw       []middleware.Middleware
}

func (i *streamClientInterceptor) WrapUnary(next connectrpc.UnaryFunc) connectrpc.UnaryFunc {
	return next
}

func (i *streamClientInterceptor) WrapStreamingHandler(next connectrpc.StreamingHandlerFunc) connectrpc.StreamingHandlerFunc {
	return next
}

func (i *streamClientInterceptor) WrapStreamingClient(next connectrpc.StreamingClientFunc) connectrpc.StreamingClientFunc {
	return func(ctx context.Context, spec connectrpc.Spec) connectrpc.StreamingClientConn {
		tr := &Transport{
			endpoint:    i.endpoint,
			operation:   spec.Procedure,
			reqHeader:   NewHeader(http.Header{}),
			replyHeader: NewHeader(http.Header{}),
		}
		ctx = transport.NewClientContext(ctx, tr)

		h := func(ctx context.Context, _ any) (any, error) {
			return next(ctx, spec), nil
		}
		if len(i.mw) > 0 {
			h = middleware.Chain(i.mw...)(h)
		}
		connAny, err := h(ctx, nil)
		if err != nil {
			return &errorStreamingClientConn{spec: spec, err: ErrorToConnect(err)}
		}
		conn, ok := connAny.(connectrpc.StreamingClientConn)
		if !ok {
			return &errorStreamingClientConn{spec: spec, err: stderrors.New("invalid streaming client conn type")}
		}
		return &streamingClientConnWrapper{StreamingClientConn: conn, tr: tr}
	}
}

type streamingClientConnWrapper struct {
	connectrpc.StreamingClientConn
	tr *Transport
}

func (w *streamingClientConnWrapper) RequestHeader() http.Header {
	h := w.StreamingClientConn.RequestHeader()
	w.tr.reqHeader = NewHeader(h)
	return h
}

func (w *streamingClientConnWrapper) ResponseHeader() http.Header {
	h := w.StreamingClientConn.ResponseHeader()
	w.tr.replyHeader = NewHeader(h)
	return h
}

func (w *streamingClientConnWrapper) ResponseTrailer() http.Header {
	h := w.StreamingClientConn.ResponseTrailer()
	w.tr.replyHeader = NewHeader(h)
	return h
}

type errorStreamingClientConn struct {
	spec connectrpc.Spec
	err  error
}

func (e *errorStreamingClientConn) Spec() connectrpc.Spec        { return e.spec }
func (e *errorStreamingClientConn) Peer() connectrpc.Peer        { return connectrpc.Peer{} }
func (e *errorStreamingClientConn) Send(any) error               { return e.err }
func (e *errorStreamingClientConn) RequestHeader() http.Header   { return make(http.Header) }
func (e *errorStreamingClientConn) CloseRequest() error          { return e.err }
func (e *errorStreamingClientConn) Receive(any) error            { return e.err }
func (e *errorStreamingClientConn) ResponseHeader() http.Header  { return make(http.Header) }
func (e *errorStreamingClientConn) ResponseTrailer() http.Header { return make(http.Header) }
func (e *errorStreamingClientConn) CloseResponse() error         { return e.err }

type cancelOnCloseReadCloser struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (c *cancelOnCloseReadCloser) Close() error {
	err := c.ReadCloser.Close()
	if c.cancel != nil {
		c.cancel()
	}
	return err
}
