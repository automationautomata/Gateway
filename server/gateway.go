package server

import (
	"context"
	"fmt"
	"gateway/config"
	"gateway/server/common"
	"gateway/server/interfaces"
	"gateway/server/limiter"
	"net/http"
)

type Gateway struct {
	EdgeLimiter     *limiter.RateLimiter
	InternalLimiter *limiter.RateLimiter // может быть nil
	Router          *UpstreamRouter
	Log             interfaces.Logger
}

func (g *Gateway) Handler() http.Handler {
	return g.EdgeLimiter.Wrap(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			proxy := g.Router.find(common.GetHost(r), common.NormalizePath(r.URL.Path))

			if proxy == nil {
				http.Error(w, "no upstream configured", http.StatusBadGateway)
				g.Log.Debug(
					r.Context(),
					"proxy request failed",
					map[string]any{
						"host": common.GetHost(r), "path": r.URL.Path,
					},
				)

				return
			}

			g.Log.Debug(
				r.Context(),
				"proxy request",
				map[string]any{
					"host": common.GetHost(r), "path": r.URL.Path, "upstream": proxy.Upstream(),
				},
			)

			if g.InternalLimiter == nil {
				proxy.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), limiter.LimiterContextKey, proxy.Upstream())
			h := g.InternalLimiter.Wrap(proxy)
			h.ServeHTTP(w, r.WithContext(ctx))
		}),
	)
}

type LimiterOptions struct {
	Metric interfaces.LimiterMetric
	Lim    interfaces.Limiter
	Log    interfaces.Logger
}

type GatewayBuilder struct {
	edgeLimiter     *limiter.RateLimiter
	internalLimiter *limiter.RateLimiter
	router          *UpstreamRouter
	log             interfaces.Logger
	err             error
}

func NewGatewayBuilder() *GatewayBuilder {
	return &GatewayBuilder{}
}

func (b *GatewayBuilder) Router(settings config.RouterSettings, metric interfaces.ProxyMetric) *GatewayBuilder {
	router, err := NewUpstreamRouter(
		settings.Routes, settings.Upstreams, metric, settings.DefaultUpstream,
	)
	if err != nil {
		b.err = fmt.Errorf("cannot create router for proxy: %w", err)
	}

	b.router = router
	return b
}

func (b *GatewayBuilder) EdgeLimiter(opt LimiterOptions, global bool) *GatewayBuilder {
	edgeLimKey := limiter.IP
	if global {
		edgeLimKey = limiter.Global
	}

	b.edgeLimiter = limiter.NewRateLimiter(
		opt.Lim, opt.Log, limiter.WithKeyType(edgeLimKey), limiter.WithMetric(opt.Metric),
	)
	return b
}

func (b *GatewayBuilder) InternalLimiter(opt LimiterOptions) *GatewayBuilder {
	b.internalLimiter = limiter.NewRateLimiter(
		opt.Lim, opt.Log, limiter.WithKeyType(limiter.ContextValue), limiter.WithMetric(opt.Metric),
	)
	return b
}

func (b *GatewayBuilder) Logger(log interfaces.Logger) *GatewayBuilder {
	b.log = log
	return b
}

func (b *GatewayBuilder) Build() (*Gateway, error) {
	if b.err != nil {
		return nil, b.err
	}

	return &Gateway{
		Log:             b.log,
		Router:          b.router,
		EdgeLimiter:     b.edgeLimiter,
		InternalLimiter: b.internalLimiter,
	}, nil
}
