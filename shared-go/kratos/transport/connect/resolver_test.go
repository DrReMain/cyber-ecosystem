package connect

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/selector"
)

func TestParseTarget(t *testing.T) {
	t1, err := parseTarget("localhost:13000", true)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(t1, &Target{Scheme: "http", Authority: "localhost:13000"}) {
		t.Fatalf("target=%+v", t1)
	}

	t2, err := parseTarget("discovery:///template1", true)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(t2, &Target{Scheme: "discovery", Endpoint: "template1"}) {
		t.Fatalf("target=%+v", t2)
	}

	t3, err := parseTarget("https://127.0.0.1:13000", false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(t3, &Target{Scheme: "https", Authority: "127.0.0.1:13000"}) {
		t.Fatalf("target=%+v", t3)
	}

	t4, err := parseTarget("127.0.0.1:13000", false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(t4, &Target{Scheme: "https", Authority: "127.0.0.1:13000"}) {
		t.Fatalf("target=%+v", t4)
	}
}

type mockRebalancer struct {
	last []selector.Node
}

func (m *mockRebalancer) Apply(nodes []selector.Node) {
	m.last = nodes
}

type mockDiscoveryResolver struct {
	watcher registry.Watcher
	err     error
}

func (m *mockDiscoveryResolver) GetService(context.Context, string) ([]*registry.ServiceInstance, error) {
	return nil, nil
}

func (m *mockDiscoveryResolver) Watch(context.Context, string) (registry.Watcher, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.watcher, nil
}

type mockWatcherResolver struct {
	ctx       context.Context
	items     [][]*registry.ServiceInstance
	i         int
	nextErr   error
	stopErr   error
	sleepEach time.Duration
}

func (m *mockWatcherResolver) Next() ([]*registry.ServiceInstance, error) {
	select {
	case <-m.ctx.Done():
		return nil, m.ctx.Err()
	default:
	}
	if m.nextErr != nil {
		return nil, m.nextErr
	}
	if m.sleepEach > 0 {
		time.Sleep(m.sleepEach)
	}
	if m.i >= len(m.items) {
		return m.items[len(m.items)-1], nil
	}
	out := m.items[m.i]
	m.i++
	return out, nil
}

func (m *mockWatcherResolver) Stop() error {
	return m.stopErr
}

func TestResolver(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	target, err := parseTarget("discovery:///demo", true)
	if err != nil {
		t.Fatal(err)
	}

	validList := []*registry.ServiceInstance{
		{
			ID:        "1",
			Name:      "demo",
			Version:   "v1",
			Endpoints: []string{fmt.Sprintf("connect://127.0.0.1:18080?isSecure=%v", false)},
		},
	}

	rb := &mockRebalancer{}
	w := &mockWatcherResolver{
		ctx:   ctx,
		items: [][]*registry.ServiceInstance{validList},
	}
	r, err := newResolver(
		ctx,
		&mockDiscoveryResolver{watcher: w},
		target,
		rb,
		true,
		true,
		25,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(rb.last) == 0 {
		t.Fatal("rebalancer should receive nodes")
	}
	if err := r.Close(); err != nil {
		t.Fatalf("close error=%v", err)
	}
}

func TestResolverWatchError(t *testing.T) {
	ctx := context.Background()
	target, err := parseTarget("discovery:///demo", true)
	if err != nil {
		t.Fatal(err)
	}
	_, err = newResolver(
		ctx,
		&mockDiscoveryResolver{err: errors.New("watch error")},
		target,
		&mockRebalancer{},
		true,
		true,
		25,
	)
	if err == nil {
		t.Fatal("expected watch error")
	}
}

func TestResolverBlockNextError(t *testing.T) {
	ctx := context.Background()
	target, err := parseTarget("discovery:///demo", true)
	if err != nil {
		t.Fatal(err)
	}
	_, err = newResolver(
		ctx,
		&mockDiscoveryResolver{
			watcher: &mockWatcherResolver{
				ctx:     ctx,
				nextErr: errors.New("next error"),
			},
		},
		target,
		&mockRebalancer{},
		true,
		true,
		25,
	)
	if err == nil {
		t.Fatal("expected block next error")
	}
}

func TestResolverBlockContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	target, err := parseTarget("discovery:///demo", true)
	if err != nil {
		t.Fatal(err)
	}
	_, err = newResolver(
		ctx,
		&mockDiscoveryResolver{
			watcher: &mockWatcherResolver{
				ctx: ctx,
			},
		},
		target,
		&mockRebalancer{},
		true,
		true,
		25,
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

func TestResolverCloseStopError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	target, err := parseTarget("discovery:///demo", true)
	if err != nil {
		t.Fatal(err)
	}
	stopErr := errors.New("stop error")
	r, err := newResolver(
		ctx,
		&mockDiscoveryResolver{
			watcher: &mockWatcherResolver{
				ctx:       ctx,
				items:     [][]*registry.ServiceInstance{{}},
				sleepEach: 1 * time.Millisecond,
				stopErr:   stopErr,
			},
		},
		target,
		&mockRebalancer{},
		false,
		true,
		25,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !errors.Is(r.Close(), stopErr) {
		t.Fatalf("close err mismatch")
	}
}
