package proxy

import (
	"fmt"
	"strings"

	"gateway/config"
	"gateway/server/common"
)

type host struct {
	paths        *common.SyncMap[string, *upstreamProxy]
	defaultProxy *upstreamProxy
}

func newHost(paths []config.Path, upstreams config.Upstreams, def *upstreamProxy) (h *host, err error) {
	m := common.NewSyncMap[string, *upstreamProxy]()
	var root string

	for _, p := range paths {
		path := common.NormalizePath(p.Path)
		if path != "/" {
			addr, ok := upstreams[p.Upstream]
			if !ok {
				return nil, fmt.Errorf("upstream %s not found", p.Upstream)
			}

			p, err := newUpstreamProxy(addr, path)
			if err != nil {
				return nil, err
			}
			m.Add(path, p)
		} else {
			root = p.Upstream
		}
	}

	if root != "" {
		if def != nil {
			return nil, fmt.Errorf("default already set")
		}

		addr, ok := upstreams[root]
		if !ok {
			return nil, fmt.Errorf("default upstream not found")
		}

		def, err = newUpstreamProxy(addr, "/")
		if err != nil {
			return nil, err
		}
	}

	return &host{m, def}, nil
}

type router struct {
	hosts        *common.SyncMap[string, *host]
	defaultProxy *upstreamProxy
}

func newRouter(routes []config.Route, upstreams config.Upstreams, defaultUpstream *string) (_ *router, err error) {
	m := common.NewSyncMap[string, *host]()
	for _, r := range routes {
		var def *upstreamProxy
		if r.Default != nil {
			addr, ok := upstreams[*r.Default]
			if !ok {
				return nil, fmt.Errorf("default upstream %q not found", *r.Default)
			}
			def, err = newUpstreamProxy(addr, "/")
			if err != nil {
				return nil, err
			}
		}

		h, err := newHost(r.Paths, upstreams, def)
		if err != nil {
			return nil, fmt.Errorf("cannot create proxy for host %s: %w", r.Host, err)
		}
		m.Add(r.Host, h)
	}

	var def *upstreamProxy
	if defaultUpstream != nil {
		addr, ok := upstreams[*defaultUpstream]
		if !ok {
			return nil, fmt.Errorf("default upstream %q not found", *defaultUpstream)
		}
		def, err = newUpstreamProxy(addr, "/")
		if err != nil {
			return nil, err
		}
	}

	return &router{m, def}, nil
}

// Поиск по наибольшему общему префиксу пути
func (r *router) find(hostname, path string) *upstreamProxy {
	path = common.NormalizePath(path)
	h, ok := r.hosts.Get(hostname)
	if !ok {
		return nil
	}

	var proxy *upstreamProxy
	var curPath strings.Builder
	for _, part := range strings.Split(path, "/")[1:] {
		curPath.WriteString(part)
		curPath.WriteString("/")
		if p, ok := h.paths.Get(curPath.String()); ok {
			proxy = p
		}
	}

	if proxy != nil {
		return proxy
	}

	if h.defaultProxy != nil {
		return h.defaultProxy
	}

	return r.defaultProxy
}
