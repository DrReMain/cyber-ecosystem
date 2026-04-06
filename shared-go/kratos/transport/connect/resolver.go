package connect

import (
	"context"
	stderrors "errors"
	"net/url"
	"strings"
	"time"

	"github.com/go-kratos/aegis/subset"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/selector"

	"cyber-ecosystem/shared-go/kratos/transport/connect/internal/endpoint"
)

type Target struct {
	Scheme    string
	Authority string
	Endpoint  string
}

func (t *Target) baseURL() string {
	switch t.Scheme {
	case "discovery":
		if t.Endpoint == "" {
			return "http://localhost"
		}
		return "http://" + t.Endpoint
	default:
		s := t.Scheme
		if s == "" {
			s = "http"
		}
		return strings.TrimRight((&url.URL{Scheme: s, Host: t.Authority}).String(), "/")
	}
}

func parseTarget(raw string, insecure bool) (*Target, error) {
	if !strings.Contains(raw, "://") {
		if insecure {
			raw = "http://" + raw
		} else {
			raw = "https://" + raw
		}
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	t := &Target{Scheme: u.Scheme, Authority: u.Host}
	if len(u.Path) > 1 {
		t.Endpoint = strings.TrimPrefix(u.Path, "/")
	}
	return t, nil
}

type resolver struct {
	rebalancer selector.Rebalancer

	target     *Target
	watcher    registry.Watcher
	subsetSize int
	insecure   bool
	cancel     context.CancelFunc
}

func newResolver(
	ctx context.Context,
	discovery registry.Discovery,
	target *Target,
	rebalancer selector.Rebalancer,
	block bool,
	insecure bool,
	subsetSize int,
) (*resolver, error) {
	rctx, cancel := context.WithCancel(ctx)
	watcher, err := discovery.Watch(rctx, target.Endpoint)
	if err != nil {
		cancel()
		return nil, err
	}
	r := &resolver{
		target:     target,
		watcher:    watcher,
		rebalancer: rebalancer,
		subsetSize: subsetSize,
		insecure:   insecure,
		cancel:     cancel,
	}

	if block {
		done := make(chan error, 1)
		go func() {
			for {
				services, wErr := watcher.Next()
				if wErr != nil {
					done <- wErr
					return
				}
				if r.update(services) {
					done <- nil
					return
				}
			}
		}()
		select {
		case waitErr := <-done:
			if waitErr != nil {
				r.cancel()
				_ = watcher.Stop()
				return nil, waitErr
			}
		case <-rctx.Done():
			r.cancel()
			_ = watcher.Stop()
			return nil, rctx.Err()
		}
	}

	go func() {
		for {
			services, wErr := watcher.Next()
			if wErr != nil {
				if stderrors.Is(wErr, context.Canceled) {
					return
				}
				log.Errorf("[connect resolver] watch %v error: %v", target, wErr)
				time.Sleep(time.Second)
				continue
			}
			r.update(services)
		}
	}()
	return r, nil
}

func (r *resolver) update(services []*registry.ServiceInstance) bool {
	filtered := make([]*registry.ServiceInstance, 0, len(services))
	connectScheme := endpoint.Scheme("connect", !r.insecure)
	httpScheme := endpoint.Scheme("http", !r.insecure)
	for _, ins := range services {
		ept, err := endpoint.ParseEndpoint(ins.Endpoints, connectScheme)
		if err != nil {
			log.Errorf("[connect resolver] parse endpoint %v error: %v", ins.Endpoints, err)
			continue
		}
		if ept == "" {
			ept, err = endpoint.ParseEndpoint(ins.Endpoints, httpScheme)
			if err != nil {
				log.Errorf("[connect resolver] parse endpoint %v error: %v", ins.Endpoints, err)
				continue
			}
		}
		if ept == "" {
			continue
		}
		filtered = append(filtered, ins)
	}
	if r.subsetSize != 0 {
		filtered = subset.Subset(r.target.Endpoint, filtered, r.subsetSize)
	}
	nodes := make([]selector.Node, 0, len(filtered))
	for _, ins := range filtered {
		ept, _ := endpoint.ParseEndpoint(ins.Endpoints, connectScheme)
		if ept == "" {
			ept, _ = endpoint.ParseEndpoint(ins.Endpoints, httpScheme)
		}
		if ept == "" {
			continue
		}
		nodes = append(nodes, selector.NewNode("connect", ept, ins))
	}
	if len(nodes) == 0 {
		log.Warnf("[connect resolver] zero endpoint found, service=%s", r.target.Endpoint)
		return false
	}
	r.rebalancer.Apply(nodes)
	return true
}

func (r *resolver) Close() error {
	if r.cancel != nil {
		r.cancel()
	}
	return r.watcher.Stop()
}
