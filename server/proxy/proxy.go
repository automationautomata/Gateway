package proxy

import (
	"fmt"
	"gateway/server/interfaces"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type ReverseProxyAdapter struct {
	*httputil.ReverseProxy
	upstream string
	prefix   string
	metric   interfaces.ProxyMetric
	inner    http.Handler
}

type Option func(*ReverseProxyAdapter)

func WithMiddlewares(mws ...interfaces.Middleware) Option {
	return func(p *ReverseProxyAdapter) {
		for i := len(mws) - 1; i >= 0; i-- {
			p.inner = mws[i].Wrap(p.inner)
		}
	}
}

func NewReverseProxyAdapter(upstream, prefix string, metric interfaces.ProxyMetric, opts ...Option) (*ReverseProxyAdapter, error) {
	target, err := url.Parse(upstream)
	if err != nil {
		return nil, err
	}

	var transport http.RoundTripper
	if def, ok := http.DefaultTransport.(*http.Transport); ok {
		defClone := def.Clone()
		defClone.Proxy = nil
		transport = defClone
	}
	p := &httputil.ReverseProxy{
		Transport: transport,
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(target)
			out, in := r.Out, r.In

			out.Host = in.Host
			out.URL.Path = strings.TrimPrefix(in.URL.Path, prefix)
			r.Out.Header.Set("X-Forwarded-Host", r.In.Host)
		},
	}
	adapter := &ReverseProxyAdapter{
		upstream:     upstream,
		prefix:       prefix,
		metric:       metric,
		ReverseProxy: p,
		inner:        p,
	}
	for _, opt := range opts {
		opt(adapter)
	}
	return adapter, nil
}

func (p *ReverseProxyAdapter) Upstream() string { return p.upstream }

func (p *ReverseProxyAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.inner.ServeHTTP(w, r)
	p.metric.Inc(fmt.Sprint(p.upstream, p.prefix))
}
