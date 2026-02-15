package server

import (
	"fmt"
	"strings"

	"gateway/config"
	"gateway/server/common"
	"gateway/server/interfaces"
	"gateway/server/proxy"
)

type host struct {
	paths        *common.SyncMap[string, *proxy.ReverseProxy]
	defaultProxy *proxy.ReverseProxy
}

func newHost(
	paths []config.Path,
	upstreams config.Upstreams,
	metric interfaces.ProxyMetric,
	def *proxy.ReverseProxy,
) (h *host, err error) {
	m := common.NewSyncMap[string, *proxy.ReverseProxy]()

	for _, p := range paths {
		path := common.NormalizePath(p.Path)
		addr, ok := upstreams[p.Upstream]
		if !ok {
			return nil, fmt.Errorf("upstream %s not found", p.Upstream)
		}

		p, err := proxy.NewReverseProxy(addr, path, metric)
		if err != nil {
			return nil, err
		}
		m.Add(path, p)
	}

	return &host{m, def}, nil
}

type UpstreamRouter struct {
	hosts        *common.SyncMap[string, *host]
	defaultProxy *proxy.ReverseProxy
}

func NewUpstreamRouter(
	routes []config.Route,
	upstreams config.Upstreams,
	metric interfaces.ProxyMetric,
	defaultUpstream *string,
) (router *UpstreamRouter, err error) {
	hosts := common.NewSyncMap[string, *host]()

	for _, r := range routes {
		var def *proxy.ReverseProxy
		if r.Default != nil {
			addr, ok := upstreams[*r.Default]
			if !ok {
				return nil, fmt.Errorf("default upstream %q not found", *r.Default)
			}

			def, err = proxy.NewReverseProxy(addr, "/", metric)
			if err != nil {
				return nil, err
			}
		}

		h, err := newHost(r.Paths, upstreams, metric, def)
		if err != nil {
			return nil, fmt.Errorf("cannot create proxy for host %s: %w", r.Host, err)
		}
		hosts.Add(r.Host, h)
	}

	router = &UpstreamRouter{hosts: hosts}

	if defaultUpstream != nil {
		addr, ok := upstreams[*defaultUpstream]
		if !ok {
			return nil, fmt.Errorf("default upstream %q not found", *defaultUpstream)
		}

		router.defaultProxy, err = proxy.NewReverseProxy(addr, "/", metric)
		if err != nil {
			return nil, err
		}
	}

	return router, nil
}

// Поиск по наибольшему общему префиксу пути
func (r *UpstreamRouter) find(hostname, path string) *proxy.ReverseProxy {
	path = common.NormalizePath(path)

	h, ok := r.hosts.Get(hostname)
	if !ok {
		return r.defaultProxy
	}

	var (
		matched *proxy.ReverseProxy
		builder strings.Builder
	)

	builder.WriteString("/")

	if p, ok := h.paths.Get(builder.String()); ok {
		matched = p
	}
	for _, part := range strings.Split(path, "/")[1:] {
		builder.WriteString(part)
		if p, ok := h.paths.Get(builder.String()); ok {
			matched = p
		}
		builder.WriteString("/")
	}

	if matched != nil {
		return matched
	}

	if h.defaultProxy != nil {
		return h.defaultProxy
	}

	return r.defaultProxy
}
