package server

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"gateway/config"
	"gateway/server/cache"
	"gateway/server/common"
	"gateway/server/interfaces"
	"gateway/server/proxy"
	"gateway/server/urlutils"
)

type UpstreamRouter struct {
	hosts        *common.SyncMap[string, *host]
	defaultProxy *ProxyAdapter
}

type CacheOptions struct {
	Metric interfaces.CacheMetric
	Cache  interfaces.CacheStorage[*cache.ResponseContent]
	Log    interfaces.Logger
}

type ProxyOptions struct {
	Upstreams config.Upstreams
	Metric    interfaces.ProxyMetric
	Default   *config.UpstreamSettings
}

type RouterOptions struct {
	Routes []config.Route
	Proxy  ProxyOptions
	Cache  CacheOptions
}

func NewUpstreamRouter(opts RouterOptions) (router *UpstreamRouter, err error) {
	hosts := common.NewSyncMap[string, *host]()

	for _, r := range opts.Routes {
		var def *config.UpstreamSettings
		if r.Default != nil {
			def = r.Default.UpstreamSettings
		}
		h, err := newHost(hostOptions{r.Paths, opts.Proxy, opts.Cache, def})
		if err != nil {
			return nil, fmt.Errorf("cannot create proxy for host %s: %w", r.Host, err)
		}
		hosts.Add(r.Host, h)
	}
	router = &UpstreamRouter{hosts: hosts}

	if opts.Proxy.Default != nil {
		defSettings := opts.Proxy.Default
		proxy, err := newProxyWithCache(
			config.Path{Path: "/", UpstreamSettings: *defSettings}, opts.Proxy, opts.Cache,
		)
		if err != nil {
			return nil, err
		}
		router.defaultProxy = proxy
	}
	return router, nil
}

// Поиск по наибольшему общему префиксу пути
func (r *UpstreamRouter) find(hostname, path string) *ProxyAdapter {
	path = urlutils.NormalizePath(path)

	h, ok := r.hosts.Get(hostname)
	if !ok {
		return r.defaultProxy
	}

	var (
		matched *ProxyAdapter
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

type host struct {
	paths        *common.SyncMap[string, *ProxyAdapter]
	defaultProxy *ProxyAdapter
}

type hostOptions struct {
	paths       []config.Path
	proxy       ProxyOptions
	cache       CacheOptions
	defSettings *config.UpstreamSettings
}

func newHost(opts hostOptions) (h *host, err error) {
	m := common.NewSyncMap[string, *ProxyAdapter]()

	for _, p := range opts.paths {
		proxy, err := newProxyWithCache(p, opts.proxy, opts.cache)
		if err != nil {
			return nil, err
		}
		m.Add(urlutils.NormalizePath(p.Path), proxy)
	}

	h = &host{m, nil}

	if opts.defSettings != nil {
		proxy, err := newProxyWithCache(
			config.Path{Path: "/", UpstreamSettings: *opts.defSettings}, opts.proxy, opts.cache,
		)
		if err != nil {
			return nil, err
		}
		h.defaultProxy = proxy
	}

	return h, nil
}

func newProxyWithCache(
	path config.Path,
	proxyOpts ProxyOptions,
	cacheOpts CacheOptions,
) (*ProxyAdapter, error) {
	addr, ok := proxyOpts.Upstreams[path.Upstream]
	if !ok {
		return nil, fmt.Errorf("upstream %q not found", path.Upstream)
	}
	commonPrefix := urlutils.NormalizePath(path.Path)
	proxy, err := proxy.NewReverseProxy(addr, commonPrefix, proxyOpts.Metric)
	if err != nil {
		return nil, err
	}

	if path.Cache == nil {
		return NewProxyAdapter(proxy), nil
	}

	cachePaths := make(map[string]time.Duration)
	for p, ttl := range *path.Cache {
		fullPath, err := url.JoinPath(commonPrefix, urlutils.NormalizePath(p))
		if err != nil {
			return nil, err
		}
		cachePaths[fullPath] = ttl
	}
	cache := cache.NewCacheMiddleware(cachePaths, cacheOpts.Metric, cacheOpts.Cache, cacheOpts.Log)
	return NewProxyAdapter(proxy, cache), nil
}
