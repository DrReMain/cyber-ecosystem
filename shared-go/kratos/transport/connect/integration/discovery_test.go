// Tests in this file exercise the client-side service-discovery resolver
// (transport/connect/resolver.go) end-to-end against a mock registry. They
// validate the full discovery path: Dial("discovery:///svc") → resolver watches
// the registry → parses connect:// endpoints → feeds nodes to the wrr selector
// → clientRoundTripper rewrites request URL.Host to the selected node. Coverage:
//   - TestDiscoveryNodeSelection — with two registered nodes (nodeA, nodeB) and
//     WithBlock(), a Raw call succeeds and is served by one of them.
//   - TestDiscoveryFailover      — registry starts with only nodeA; after the
//     registry swaps to nodeB, a subsequent Raw call reaches nodeB.
//   - TestDiscoveryNodeFilter    — WithNodeFilter rejects nodeA by metadata, so
//     every call lands on nodeB.
//
// The mock registry mirrors how a real registry.Watcher behaves: the first
// Next() returns the current instances (so the resolver's blocking initial
// resolve completes), and later Next() calls block until setServices notifies.
package integration

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	connectrpc "connectrpc.com/connect"
	"github.com/go-kratos/kratos/v3/registry"
	"github.com/go-kratos/kratos/v3/selector"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc"

	"cyber-ecosystem/shared-go/kratos/transport/connect"

	mobilepb "cyber-ecosystem/gen/go/cyber/mobile/v1"
	mobilev1connect "cyber-ecosystem/gen/go/cyber/mobile/v1/v1connect"
)

// taggedService implements TransferServiceServer and stamps every Raw response
// with a fixed node tag (the node identifier the server was constructed with).
// This lets a discovery test assert WHICH registered node served a given call:
// the response Data is "<tag>:<echo of request data>".
type taggedService struct {
	mobilepb.UnimplementedMobileTransferServiceServer
	tag string
}

func (s taggedService) Raw(_ context.Context, req *mobilepb.RawRequest) (*httpbody.HttpBody, error) {
	ct := req.GetContentType()
	if ct == "" {
		ct = "application/octet-stream"
	}
	return &httpbody.HttpBody{
		ContentType: ct,
		Data:        []byte(s.tag + ":" + string(req.GetData())),
	}, nil
}

// Subscribe / Echo / Pipe are not used by the discovery tests but must be
// implemented to satisfy the server registration (the generated handler binds
// the whole service). They delegate to the base testService behavior.
func (s taggedService) Subscribe(req *mobilepb.SubscribeRequest, stream grpc.ServerStreamingServer[mobilepb.SubscribeResponse]) error {
	return (testService{}).Subscribe(req, stream)
}
func (s taggedService) Echo(stream grpc.ClientStreamingServer[mobilepb.EchoRequest, mobilepb.EchoResponse]) error {
	return (testService{}).Echo(stream)
}
func (s taggedService) Pipe(stream grpc.BidiStreamingServer[mobilepb.PipeRequest, mobilepb.PipeResponse]) error {
	return (testService{}).Pipe(stream)
}

// startTaggedServer brings up a loopback h2c Connect server whose transfer
// service stamps responses with the given node tag. It returns the server's
// connect:// endpoint URL (for registry registration) and a stop func.
func startTaggedServer(t *testing.T, tag string) (endpointURL string, stop func()) {
	t.Helper()
	srv := connect.NewServer(connect.Address("127.0.0.1:0"), connect.Timeout(0))
	mobilepb.RegisterMobileTransferServiceConnectServer(srv, taggedService{tag: tag})

	ep, err := srv.Endpoint()
	if err != nil {
		t.Fatalf("endpoint: %v", err)
	}
	ctx := context.Background()
	go func() { _ = srv.Start(ctx) }()
	waitReady(t, ep.Host)

	return ep.String(), func() { _ = srv.Stop(ctx) }
}

// instance builds a registry.ServiceInstance for a service name, whose single
// endpoint is the given connect:// URL (as returned by Server.Endpoint()).
// metadata carries an optional node tag used by the NodeFilter test.
func instance(name, endpointURL, id, tag string) *registry.ServiceInstance {
	return &registry.ServiceInstance{
		ID:        id,
		Name:      name,
		Version:   "v1",
		Endpoints: []string{endpointURL},
		Metadata:  map[string]string{"node": tag},
	}
}

// --- mock registry + watcher -------------------------------------------------

// mockRegistry implements registry.Discovery. Watch() returns a mockWatcher that
// first yields the current instance set immediately (so the resolver's blocking
// initial resolve completes), then blocks until setServices pushes a new set.
type mockRegistry struct {
	mu       sync.Mutex
	services map[string][]*registry.ServiceInstance
	// watchers tracks all live watchers per service name so setServices can
	// notify each of them.
	watchers map[string]map[*mockWatcher]struct{}
}

func newMockRegistry() *mockRegistry {
	return &mockRegistry{
		services: make(map[string][]*registry.ServiceInstance),
		watchers: make(map[string]map[*mockWatcher]struct{}),
	}
}

func (r *mockRegistry) GetService(_ context.Context, name string) ([]*registry.ServiceInstance, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]*registry.ServiceInstance, len(r.services[name]))
	copy(out, r.services[name])
	return out, nil
}

func (r *mockRegistry) Watch(ctx context.Context, name string) (registry.Watcher, error) {
	r.mu.Lock()
	// Snapshot the current instance set so the watcher's first Next() returns
	// it regardless of later mutations.
	cur := make([]*registry.ServiceInstance, len(r.services[name]))
	copy(cur, r.services[name])
	w := &mockWatcher{
		ctx:     ctx,
		notify:  make(chan struct{}, 1),
		done:    make(chan struct{}),
		initial: cur,
	}
	if r.watchers[name] == nil {
		r.watchers[name] = make(map[*mockWatcher]struct{})
	}
	r.watchers[name][w] = struct{}{}
	r.mu.Unlock()

	// If the context is canceled, wake a blocked Next() so it returns.
	go func() {
		<-ctx.Done()
		w.cancelOnce.Do(func() { close(w.done) })
	}()
	return w, nil
}

// setServices replaces the instance set for a service and notifies every live
// watcher of that service. Thread-safe; safe to call before/after Watch().
func (r *mockRegistry) setServices(name string, instances []*registry.ServiceInstance) {
	r.mu.Lock()
	snap := make([]*registry.ServiceInstance, len(instances))
	copy(snap, instances)
	r.services[name] = snap
	ws := make([]*mockWatcher, 0, len(r.watchers[name]))
	for w := range r.watchers[name] {
		ws = append(ws, w)
	}
	r.mu.Unlock()

	for _, w := range ws {
		w.push(snap)
	}
}

// mockWatcher implements registry.Watcher. The first Next() returns the snapshot
// captured at Watch() time; subsequent Next() calls block until push() delivers
// a new snapshot (or the context is canceled).
type mockWatcher struct {
	ctx        context.Context
	mu         sync.Mutex
	notify     chan struct{}
	initial    []*registry.ServiceInstance
	delivered  bool
	latest     []*registry.ServiceInstance
	cancelOnce sync.Once
	done       chan struct{}
}

// push hands a new instance snapshot to the watcher and wakes a blocked Next().
func (w *mockWatcher) push(instances []*registry.ServiceInstance) {
	snap := make([]*registry.ServiceInstance, len(instances))
	copy(snap, instances)
	w.mu.Lock()
	w.latest = snap
	w.mu.Unlock()
	// Non-blocking signal: if a Next() is already pending a notify, this is a
	// no-op and the pending Next() will pick up the newest snapshot.
	select {
	case w.notify <- struct{}{}:
	default:
	}
}

func (w *mockWatcher) Next() ([]*registry.ServiceInstance, error) {
	w.mu.Lock()
	if !w.delivered {
		w.delivered = true
		out := w.initial
		w.mu.Unlock()
		return out, nil
	}
	w.mu.Unlock()

	// Block until either a new snapshot is pushed or the context is canceled.
	select {
	case <-w.notify:
		w.mu.Lock()
		out := w.latest
		w.mu.Unlock()
		// Defensive copy so callers can't mutate our slice.
		res := make([]*registry.ServiceInstance, len(out))
		copy(res, out)
		return res, nil
	case <-w.done:
		return nil, w.ctx.Err()
	}
}

func (w *mockWatcher) Stop() error {
	w.cancelOnce.Do(func() { close(w.done) })
	return nil
}

// --- helpers shared by the three tests --------------------------------------

// dialDiscovery builds a connect client bound to "discovery:///svc" backed by
// the mock registry. WithBlock() makes Dial wait until the resolver sees the
// initial instance set, so the client is ready to call immediately on return.
// The mock registry MUST already hold at least one "svc" instance before Dial
// (WithBlock would otherwise hang until the registry pushes one).
func dialDiscovery(t *testing.T, reg *mockRegistry, extra ...connect.ClientOption) (mobilev1connect.MobileTransferServiceClient, func()) {
	t.Helper()
	ctx := context.Background()
	opts := []connect.ClientOption{
		connect.WithEndpoint("discovery:///svc"),
		connect.WithDiscovery(reg),
		connect.WithH2C(true),
		connect.WithBlock(),
	}
	opts = append(opts, extra...)
	cli, err := connect.Dial(ctx, opts...)
	if err != nil {
		t.Fatalf("dial discovery: %v", err)
	}
	client := mobilev1connect.NewMobileTransferServiceClient(cli.HTTPClient(), cli.BaseURL(), cli.ClientOptions()...)
	return client, func() { _ = cli.Close() }
}

// rawNodeTag issues a Raw call with the given payload and extracts the node tag
// the serving node stamped into the response ("<tag>:<payload>").
func rawNodeTag(t *testing.T, cli mobilev1connect.MobileTransferServiceClient, payload string) string {
	t.Helper()
	resp, err := cli.Raw(context.Background(), connectrpc.NewRequest(&mobilepb.RawRequest{
		ContentType: "text/plain",
		Data:        []byte(payload),
	}))
	if err != nil {
		t.Fatalf("Raw: %v", err)
	}
	wantPrefix := ":"
	s := string(resp.Msg.Data)
	if i := indexByte(s, ':'); i < 0 {
		t.Fatalf("response %q missing %q node-tag separator", s, wantPrefix)
	} else {
		return s[:i]
	}
	return ""
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

// --- tests ------------------------------------------------------------------

// TestDiscoveryNodeSelection proves the resolver → wrr selector →
// clientRoundTripper path works: with nodeA + nodeB registered, WithBlock()
// resolves, and a Raw call succeeds and is served by one of the two registered
// nodes (identified by the tag the node stamps into the response).
func TestDiscoveryNodeSelection(t *testing.T) {
	nodeAURL, stopA := startTaggedServer(t, "nodeA")
	defer stopA()
	nodeBURL, stopB := startTaggedServer(t, "nodeB")
	defer stopB()

	reg := newMockRegistry()
	reg.setServices("svc", []*registry.ServiceInstance{
		instance("svc", nodeAURL, "a", "nodeA"),
		instance("svc", nodeBURL, "b", "nodeB"),
	})

	cli, stop := dialDiscovery(t, reg)
	defer stop()

	tag := rawNodeTag(t, cli, "ping")
	if tag != "nodeA" && tag != "nodeB" {
		t.Fatalf("served by %q, want nodeA or nodeB", tag)
	}
}

// TestDiscoveryFailover proves the resolver picks up dynamic registry updates.
// The registry starts with only nodeA; after a Raw call is served by nodeA, the
// registry is updated to drop nodeA and add nodeB. The resolver's background
// watch loop receives the update (mockWatcher.Next() unblocks on setServices),
// calls rebalancer.Apply with the new [nodeB] node set, and the wrr balancer
// drops nodeA from its pickable set. Subsequent calls must therefore land on
// nodeB.
//
// Connection-reuse caveat: HTTP/2 multiplexes over a single connection per host,
// and the clientRoundTripper rewrites URL.Host per request AFTER connection
// selection, so a call could in principle reuse an idle connection to the OLD
// host. In practice the host rewrite routes the new request to nodeB's address,
// so we additionally retry a few times with a short backoff and assert that at
// least one call reaches nodeB.
func TestDiscoveryFailover(t *testing.T) {
	nodeAURL, stopA := startTaggedServer(t, "nodeA")
	defer stopA()
	nodeBURL, stopB := startTaggedServer(t, "nodeB")
	defer stopB()

	reg := newMockRegistry()
	reg.setServices("svc", []*registry.ServiceInstance{
		instance("svc", nodeAURL, "a", "nodeA"),
	})

	cli, stop := dialDiscovery(t, reg)
	defer stop()

	// Phase 1: only nodeA is registered, so the call MUST hit nodeA.
	if tag := rawNodeTag(t, cli, "first"); tag != "nodeA" {
		t.Fatalf("phase 1 served by %q, want nodeA", tag)
	}

	// Phase 2: swap the registry to nodeB only. The resolver's watch loop gets
	// the update via mockWatcher.push → Next() unblocks → Apply([nodeB]).
	reg.setServices("svc", []*registry.ServiceInstance{
		instance("svc", nodeBURL, "b", "nodeB"),
	})

	// Retry loop: the background watch + Apply is asynchronous w.r.t. this
	// goroutine, so give the selector a moment to converge on [nodeB].
	deadline := time.Now().Add(3 * time.Second)
	var lastTag string
	for time.Now().Before(deadline) {
		lastTag = rawNodeTag(t, cli, "failover")
		if lastTag == "nodeB" {
			return // success: a post-update call reached nodeB
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("failover to nodeB never observed; last serving node = %q", lastTag)
}

// TestDiscoveryNodeFilter proves WithNodeFilter constrains node selection. Both
// nodeA and nodeB are registered, but the client installs a NodeFilter that
// drops any node whose metadata["node"] == "nodeA". Every call must therefore
// land on nodeB.
func TestDiscoveryNodeFilter(t *testing.T) {
	nodeAURL, stopA := startTaggedServer(t, "nodeA")
	defer stopA()
	nodeBURL, stopB := startTaggedServer(t, "nodeB")
	defer stopB()

	reg := newMockRegistry()
	reg.setServices("svc", []*registry.ServiceInstance{
		instance("svc", nodeAURL, "a", "nodeA"),
		instance("svc", nodeBURL, "b", "nodeB"),
	})

	// dropNodeA filters out the node tagged "nodeA" by its instance metadata.
	// selector.NodeFilter runs inside Selector.Select after Apply, so only
	// surviving nodes are pickable by the wrr balancer.
	dropNodeA := func(_ context.Context, nodes []selector.Node) []selector.Node {
		out := nodes[:0]
		for _, n := range nodes {
			if n.Metadata()["node"] == "nodeA" {
				continue
			}
			out = append(out, n)
		}
		return out
	}

	cli, stop := dialDiscovery(t, reg, connect.WithNodeFilter(dropNodeA))
	defer stop()

	// A handful of calls — all must hit nodeB, never nodeA.
	for i := 0; i < 5; i++ {
		tag := rawNodeTag(t, cli, fmt.Sprintf("c%d", i))
		if tag != "nodeB" {
			t.Fatalf("call %d served by %q, want nodeB (nodeA filtered out)", i, tag)
		}
	}
}

// Compile-time assertion that taggedService satisfies the generated server
// interface the registration helper expects (RegisterTransferServiceConnectServer
// binds the full TransferServiceServer). The Unimplemented embed covers the
// rest; we only override Raw (+ delegate the stream methods), so this catches a
// future interface drift.
var _ mobilepb.MobileTransferServiceServer = taggedService{}
