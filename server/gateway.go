package server

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gateway/config"
	"gateway/server/cache"
	"gateway/server/interfaces"
	"gateway/server/limiter"
	"gateway/server/proxy"
	"gateway/server/urlutils"
)

type Gateway struct {
	EdgeLimiter     *limiter.RateLimiter
	InternalLimiter *limiter.RateLimiter // может быть nil
	Router          *Router
	Log             interfaces.Logger
}

func (g *Gateway) Handler() http.Handler {
	return g.EdgeLimiter.Wrap(http.HandlerFunc(g.serve))
}

func (g *Gateway) serve(w http.ResponseWriter, r *http.Request) {
	host, path := urlutils.GetHost(r), urlutils.NormalizePath(r.URL.Path)

	proxyAdapter, found := g.Router.Find(host, path)
	if !found {
		http.Error(w, "no upstream configured", http.StatusBadGateway)
		g.Log.Debug(
			r.Context(),
			"proxy request failed - no upstream",
			map[string]any{
				"host": urlutils.GetHost(r),
				"path": r.URL.Path,
			},
		)
		return
	}

	g.Log.Debug(
		r.Context(),
		"proxy request",
		map[string]any{
			"host":     urlutils.GetHost(r),
			"path":     r.URL.Path,
			"upstream": proxyAdapter.Upstream(),
		},
	)

	if g.InternalLimiter == nil {
		proxyAdapter.ServeHTTP(w, r)
		return
	}

	r = r.WithContext(context.WithValue(r.Context(), limiter.LimiterContextKey, proxyAdapter.Upstream()))
	h := g.InternalLimiter.Wrap(proxyAdapter)
	h.ServeHTTP(w, r)
}

type GatewayBuilder struct {
	router          *Router
	edgeLimiter     *limiter.RateLimiter
	internalLimiter *limiter.RateLimiter
	logger          interfaces.Logger
	err             error
}

type LimiterOptions struct {
	Metric  interfaces.LimiterMetric
	Limiter interfaces.Limiter
	Log     interfaces.Logger
}

type CacheOptions struct {
	Metric interfaces.CacheMetric
	Store  interfaces.CacheStorage[*cache.ResponseContent]
	Log    interfaces.Logger
}

type ProxyOptions struct {
	Metric  interfaces.ProxyMetric
	Default *config.UpstreamSettings
}

type RouterOptions struct {
	Settings config.RouterSettings
	Proxy    ProxyOptions
	Cache    *CacheOptions
}

func NewGatewayBuilder() *GatewayBuilder {
	return &GatewayBuilder{}
}

func (b *GatewayBuilder) Router(opts RouterOptions) *GatewayBuilder {
	if b.err != nil {
		return b
	}

	r := NewRouter()

	makeAdapter := func(upstream, prefix string, cache *config.Caches) (*proxy.ReverseProxyAdapter, error) {
		return b.createProxyAdapter(upstream, prefix, cache, opts.Proxy, opts.Cache)
	}

	settings := opts.Settings
	for _, route := range settings.Routes {
		host := route.Host

		if route.Default != nil {
			up, err := resolveUpstream(route.Default.UpstreamAlias, settings.UpstreamsAliases)
			if err != nil {
				b.err = err
				return b
			}
			adapter, err := makeAdapter(up, "", route.Default.Cache)
			if err != nil {
				b.err = fmt.Errorf("cannot create default proxy for host %s: %w", host, err)
				return b
			}
			r.AddDefault(host, adapter)
		}

		for _, path := range route.Paths {
			up, err := resolveUpstream(path.UpstreamAlias, settings.UpstreamsAliases)
			if err != nil {
				b.err = err
				return b
			}

			adapter, err := makeAdapter(up, path.Path, path.Cache)
			if err != nil {
				b.err = fmt.Errorf("cannot create proxy for route %s %s: %w", host, path.Path, err)
				return b
			}
			r.Add(host, path.Path, adapter)
		}
	}

	if opts.Proxy.Default != nil {
		def := opts.Proxy.Default
		up, err := resolveUpstream(def.UpstreamAlias, settings.UpstreamsAliases)
		if err != nil {
			b.err = err
			return b
		}

		adapter, err := makeAdapter(up, "", def.Cache)
		if err != nil {
			b.err = fmt.Errorf("cannot create global default proxy: %w", err)
			return b
		}
		r.globalDefault = adapter
	}

	b.router = r
	return b
}

func resolveUpstream(name string, upstreams map[string]string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("empty upstream name")
	}
	if upstreams == nil {
		return name, nil
	}
	if url, ok := upstreams[name]; ok {
		return url, nil
	}

	if strings.Contains(name, "://") {
		return name, nil
	}
	return "", fmt.Errorf("upstream alias %q not found", name)
}

func (b *GatewayBuilder) createProxyAdapter(
	upstream string,
	prefix string,
	cacheMap *config.Caches,
	proxyOpts ProxyOptions,
	cacheOpts *CacheOptions,
) (*proxy.ReverseProxyAdapter, error) {
	n := urlutils.NormalizePath(prefix)
	if cacheMap == nil || len(*cacheMap) == 0 {
		return proxy.NewReverseProxyAdapter(upstream, n, proxyOpts.Metric)
	}

	if cacheOpts == nil {
		return nil, fmt.Errorf("cacheMap provided without CacheOptions")
	}

	cachePaths := make(map[string]time.Duration, len(*cacheMap))
	for path, ttl := range *cacheMap {
		fullPath, err := url.JoinPath(n, urlutils.NormalizePath(path))
		if err != nil {
			return nil, err
		}
		cachePaths[fullPath] = ttl
	}

	mw := cache.NewCacheMiddleware(
		cachePaths,
		cacheOpts.Metric,
		cacheOpts.Store,
		cacheOpts.Log,
	)

	return proxy.NewReverseProxyAdapter(
		upstream,
		n,
		proxyOpts.Metric,
		proxy.WithMiddlewares(mw),
	)
}

func (b *GatewayBuilder) EdgeLimiter(opts LimiterOptions, global bool) *GatewayBuilder {
	if b.err != nil {
		return b
	}

	keyType := limiter.IP
	if global {
		keyType = limiter.Global
	}

	b.edgeLimiter = limiter.NewRateLimiter(
		opts.Limiter,
		opts.Log,
		limiter.WithKeyType(keyType),
		limiter.WithMetric(opts.Metric),
	)
	return b
}

func (b *GatewayBuilder) InternalLimiter(opts LimiterOptions) *GatewayBuilder {
	if b.err != nil {
		return b
	}

	b.internalLimiter = limiter.NewRateLimiter(
		opts.Limiter,
		opts.Log,
		limiter.WithKeyType(limiter.ContextValue),
		limiter.WithMetric(opts.Metric),
	)
	return b
}

func (b *GatewayBuilder) Logger(log interfaces.Logger) *GatewayBuilder {
	b.logger = log
	return b
}

func (b *GatewayBuilder) Build() (*Gateway, error) {
	if b.err != nil {
		return nil, b.err
	}
	if b.router == nil {
		return nil, fmt.Errorf("router must be configured with Router")
	}
	if b.edgeLimiter == nil {
		return nil, fmt.Errorf("edge limiter must be configured")
	}
	if b.logger == nil {
		return nil, fmt.Errorf("logger must be configured")
	}

	return &Gateway{
		Router:          b.router,
		EdgeLimiter:     b.edgeLimiter,
		InternalLimiter: b.internalLimiter,
		Log:             b.logger,
	}, nil
}
