package handlers

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"gateway/config"
	"gateway/server/interfaces"
)

type proxy struct {
	backend string
	*httputil.ReverseProxy
}

func newProxy(rawURL string) (*proxy, error) {
	target, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	return &proxy{
		backend: rawURL,
		ReverseProxy: &httputil.ReverseProxy{
			Rewrite: func(r *httputil.ProxyRequest) {
				r.SetURL(target)
				r.Out.Host = r.In.Host
				// r.Out.Header.Set("X-Forwarded-Host", r.In.Host)
			},
		},
	}, nil
}

type hostRule struct {
	defaultProxy *proxy
	pathRules    map[string]*proxy
}

type proxyLimiter struct {
	interfaces.Limiter
	metric interfaces.LimiterMetric
}

type HttpReverseProxy struct {
	hostMapping  map[string]hostRule
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
	p = &HttpReverseProxy{}
	for _, opt := range options {
		opt(p)
	}

	p.defaultProxy, err = newProxy(rules.Default)
	if err != nil {
		return nil, fmt.Errorf("cannot create deafult reverse proxy: %w", err)
	}

	if rules.Hosts == nil {
		return p, nil
	}

	p.hostMapping, err = createHostMapping(rules.Hosts)
	if err != nil {
		return nil, fmt.Errorf("cannot create reverse proxy: %w", err)
	}
	return p, nil
}

func (hp *HttpReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p := hp.getProxy(r.Host, r.URL.Path)

	if hp.lim != nil {
		allow, err := hp.lim.Allow(r.Context(), p.backend)
		if err != nil {
			log.Printf("rate limiter failed to %s: %s", p.backend, err)
		}

		hp.lim.metric.Record(allow, p.backend)
		if !allow {
			return
		}
	}
	p.ServeHTTP(w, r)
	hp.metric.Record(p.backend)
}

func (p *HttpReverseProxy) getProxy(host, path string) *proxy {
	rule, ok := p.hostMapping[host]
	if !ok {
		return p.defaultProxy
	}

	var curPath strings.Builder
	for _, part := range strings.Split(path, "/") {
		curPath.WriteString("/")
		curPath.WriteString(part)
		if proxy, ok := rule.pathRules[curPath.String()]; ok {
			return proxy
		}
	}

	if rule.defaultProxy != nil {
		return rule.defaultProxy
	}

	return p.defaultProxy
}

func createHostMapping(hosts []config.HostRules) (mapping map[string]hostRule, err error) {
	mapping = make(map[string]hostRule)

	for _, hostCfg := range hosts {
		rules := make(map[string]*proxy)

		for prefix, backend := range hostCfg.Pathes {
			proxy, err := newProxy(backend)
			if err != nil {
				return nil, err
			}
			rules[prefix] = proxy
		}

		var defaultProxy *proxy
		if hostCfg.Default != nil {
			defaultProxy, err = newProxy(*hostCfg.Default)
			if err != nil {
				return nil, err
			}
		}

		mapping[hostCfg.Host] = hostRule{defaultProxy, rules}
	}

	return mapping, nil
}
