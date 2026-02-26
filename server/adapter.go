package server

import (
	"gateway/server/interfaces"
	"gateway/server/proxy"
	"net/http"
)

type ProxyAdapter struct {
	*proxy.ReverseProxy
	inner http.Handler
}

func NewProxyAdapter(p *proxy.ReverseProxy, mws ...interfaces.Middleware) *ProxyAdapter {
	return &ProxyAdapter{
		ReverseProxy: p,
		inner:        chain(p, mws),
	}
}

func (p *ProxyAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.inner.ServeHTTP(w, r)
}
