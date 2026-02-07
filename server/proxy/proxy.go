package proxy

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"gateway/config"
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
}

type ProxyOption func(*HttpReverseProxy)

func WithLimiter(lim interfaces.Limiter, metric interfaces.LimiterMetric) ProxyOption {
	return func(hp *HttpReverseProxy) {
		hp.lim = &proxyLimiter{lim, metric}
	}
}

func NewHttpReverseProxy(
	rules config.ReverseProxyRules, proxyMetric interfaces.ProxyMetric, options ...ProxyOption,
) (p *HttpReverseProxy, err error) {
	p = &HttpReverseProxy{metric: proxyMetric}
	for _, opt := range options {
		opt(p)
	}

	p.defaultProxy, err = newProxy(rules.Default, "/")
	if err != nil {
		return nil, fmt.Errorf("cannot create deafult reverse proxy: %w", err)
	}

	if rules.Hosts == nil {
		return p, nil
	}

	p.hostMap, err = newHostRulesMap(rules.Hosts)
	if err != nil {
		return nil, fmt.Errorf("cannot create reverse proxy: %w", err)
	}
	return p, nil
}

func (hp *HttpReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p := hp.getProxy(r.URL.Hostname(), r.URL.Path)

	if hp.lim != nil {
		allow, err := hp.lim.Allow(r.Context(), p.backend)
		if err != nil {
			log.Printf("rate limiter failed to %s: %s", p.backend, err)
		}

		hp.lim.metric.Inc(allow, p.backend)
		if !allow { ///
			return
		}
	}
	hp.metric.Inc(p.backend)
	p.ServeHTTP(w, r)
}

func (p *HttpReverseProxy) getProxy(host, path string) *proxy {
	rule, ok := p.hostMap.get(host)
	if !ok {
		return p.defaultProxy
	}

	var curPath strings.Builder
	for _, part := range strings.Split(path, "/") {
		curPath.WriteString("/")
		curPath.WriteString(part)
		if proxy, ok := rule.pathRules.get(curPath.String()); ok {
			return proxy
		}
	}

	if rule.defaultProxy != nil {
		return rule.defaultProxy
	}

	return p.defaultProxy
}
