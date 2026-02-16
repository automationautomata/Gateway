package proxy

import (
	"fmt"
	"gateway/server/interfaces"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type ReverseProxy struct {
	*httputil.ReverseProxy
	upstream string
	prefix   string
	metric   interfaces.ProxyMetric
}

func (p *ReverseProxy) Upstream() string { return p.upstream }

func NewReverseProxy(upstream, prefix string, metric interfaces.ProxyMetric) (*ReverseProxy, error) {
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

	return &ReverseProxy{
		upstream: upstream,
		prefix:   prefix,
		metric:   metric,
		ReverseProxy: &httputil.ReverseProxy{
			Transport: transport,
			Rewrite: func(r *httputil.ProxyRequest) {
				r.SetURL(target)
				out, in := r.Out, r.In

				out.Host = in.Host
				out.URL.Path = strings.TrimPrefix(in.URL.Path, prefix)
				r.Out.Header.Set("X-Forwarded-Host", r.In.Host)
			},
		},
	}, nil
}

func (p *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.ReverseProxy.ServeHTTP(w, r)
	p.metric.Inc(fmt.Sprint(p.upstream, p.prefix))
}
