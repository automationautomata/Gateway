package proxy

import (
	"fmt"
	"net/http"
	"strings"

	"gateway/config"
	"gateway/server/common"
	"gateway/server/interfaces"
)

type proxyLimiter struct {
	interfaces.Limiter
	metric interfaces.LimiterMetric
}

type HttpReverseProxy struct {
	hostMap      *hostRulesMap
	defaultProxy *proxy
	lim          *proxyLimiter
	metric       interfaces.ProxyMetric
	log          interfaces.Logger
}
type ProxyOption func(*HttpReverseProxy)

func WithLimiter(lim interfaces.Limiter, metric interfaces.LimiterMetric) ProxyOption {
	return func(hp *HttpReverseProxy) {
		hp.lim = &proxyLimiter{lim, metric}
	}
}

type HttpProxyInput struct {
	Rules       config.ReverseProxyRules
	Log         interfaces.Logger
	ProxyMetric interfaces.ProxyMetric
}

func NewHttpReverseProxy(input HttpProxyInput, options ...ProxyOption) (p *HttpReverseProxy, err error) {
	p = &HttpReverseProxy{metric: input.ProxyMetric, log: input.Log}
	for _, opt := range options {
		opt(p)
	}

	p.defaultProxy, err = newProxy(input.Rules.Default, "/")
	if err != nil {
		return nil, fmt.Errorf("cannot create deafult reverse proxy: %w", err)
	}

	if input.Rules.Hosts == nil {
		return p, nil
	}

	p.hostMap, err = newHostRulesMap(input.Rules.Hosts)
	if err != nil {
		return nil, fmt.Errorf("cannot create reverse proxy: %w", err)
	}
	return p, nil
}

func (hp *HttpReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := common.GetHost(r)
	p := hp.getProxy(host, r.URL.Path)

	hp.log.Debug(r.Context(), "proxy", map[string]any{
		"host": host, "path": r.URL.Path, "to": p.backend,
	})

	if hp.lim != nil {
		allow, err := hp.lim.Allow(r.Context(), p.backend)
		if err != nil {
			msg := fmt.Sprintf("rate limiter failed to %s", p.backend)
			hp.log.Error(r.Context(), msg, map[string]any{"error": err})
		}

		hp.lim.metric.Inc(allow, p.backend)
		if !allow {
			return
		}
	}

	hp.metric.Inc(p.backend)
	p.ServeHTTP(w, r)
}

func (hp *HttpReverseProxy) getProxy(host, path string) *proxy {
	if hp.hostMap == nil {
		return hp.defaultProxy
	}
	path = normalizePath(path)

	rule, ok := hp.hostMap.Get(host)
	if !ok {
		return hp.defaultProxy
	}

	var curPath strings.Builder
	for _, part := range strings.Split(path, "/") {
		curPath.WriteString(part)
		curPath.WriteString("/")
		if proxy, ok := rule.pathRules.Get(curPath.String()); ok {
			return proxy
		}
	}

	if rule.defaultProxy != nil {
		return rule.defaultProxy
	}

	return hp.defaultProxy
}
