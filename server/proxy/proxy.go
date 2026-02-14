package proxy

import (
	"fmt"
	"net/http"

	"gateway/config"
	"gateway/server/common"
	"gateway/server/interfaces"
)

type limiterFacede struct {
	interfaces.Limiter
	metric interfaces.LimiterMetric
}

func (l *limiterFacede) allowRequest(r *http.Request, key string) (bool, error) {
	allowed, err := l.Allow(r.Context(), key)
	if err != nil {
		return false, err
	}

	l.metric.Inc(allowed, key)
	return allowed, nil
}

type HttpReverseProxy struct {
	router *router

	lim *limiterFacede

	metric interfaces.ProxyMetric
	log    interfaces.Logger
}

type Option func(*HttpReverseProxy)

func WithLimiter(l interfaces.Limiter, m interfaces.LimiterMetric) Option {
	return func(p *HttpReverseProxy) {
		p.lim = &limiterFacede{
			Limiter: l,
			metric:  m,
		}
	}
}

type Input struct {
	Settings config.ProxySettings
	Log      interfaces.Logger
	Metric   interfaces.ProxyMetric
	Options  []Option
}

func NewHttpReverseProxy(input Input) (*HttpReverseProxy, error) {
	p := &HttpReverseProxy{
		metric: input.Metric,
		log:    input.Log,
	}
	if input.Options != nil {
		for _, opt := range input.Options {
			opt(p)
		}
	}

	settings := input.Settings
	router, err := newRouter(
		settings.Routes, settings.Upstreams, settings.DefaultUpstream,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create host routes: %w", err)
	}

	p.router = router

	return p, nil
}

func (p *HttpReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	proxy := p.router.find(common.GetHost(r), common.NormalizePath(r.URL.Path))

	if proxy == nil {
		http.Error(w, "no upstream configured", http.StatusBadGateway)
		return
	}

	p.log.Debug(r.Context(), "proxy request", map[string]any{
		"host":     r.Host,
		"path":     r.URL.Path,
		"upstream": proxy.upstream,
	})

	if p.lim != nil {
		allow, err := p.lim.allowRequest(r, proxy.upstream)
		if err != nil {
			p.log.Error(
				r.Context(),
				"rate limiter error",
				map[string]any{"upstream": proxy.upstream, "error": err},
			)

			http.Error(w, "rate limiter failed", http.StatusInternalServerError)
			return
		}

		if !allow {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
		}
		return
	}

	p.metric.Inc(proxy.upstream)
	proxy.ServeHTTP(w, r)
}
