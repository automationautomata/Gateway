package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type upstreamProxy struct {
	upstream string
	prefix   string

	*httputil.ReverseProxy
}

func newUpstreamProxy(upstream, prefix string) (*upstreamProxy, error) {
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
	return &upstreamProxy{
		upstream: upstream,
		prefix:   prefix,
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
