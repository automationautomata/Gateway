package server

import (
	"gateway/server/common"
	"gateway/server/proxy"
	"gateway/server/urlutils"
)

type Router struct {
	hosts         *common.SyncMap[string, *routes]
	globalDefault *proxy.ReverseProxyAdapter
}

func NewRouter() *Router {
	return &Router{
		hosts:         common.NewSyncMap[string, *routes](),
		globalDefault: nil,
	}
}

func (r *Router) Add(host, path string, proxy *proxy.ReverseProxyAdapter) {
	h, ok := r.hosts.Get(host)
	if !ok {
		h = newHostRouter()
		r.hosts.Add(host, h)
	}
	h.add(urlutils.NormalizePath(path), proxy)
}

func (r *Router) AddDefault(host string, proxy *proxy.ReverseProxyAdapter) {
	h, ok := r.hosts.Get(host)
	if !ok {
		h = newHostRouter()
		r.hosts.Add(host, h)
	}
	h.setDefault(proxy)
}

// Поиск по наибольшему общему префиксу пути
func (r *Router) Find(hostname, path string) (*proxy.ReverseProxyAdapter, bool) {
	hasDefault := r.globalDefault == nil

	h, ok := r.hosts.Get(hostname)
	if !ok {
		return r.globalDefault, hasDefault
	}

	p, ok := h.find(path)
	if ok {
		return p, true
	}
	return r.globalDefault, hasDefault
}

type routes struct {
	paths        *urlutils.PathTree[*proxy.ReverseProxyAdapter]
	defaultProxy *proxy.ReverseProxyAdapter
}

func newHostRouter() *routes {
	return &routes{
		paths:        urlutils.NewPathTree[*proxy.ReverseProxyAdapter](),
		defaultProxy: nil,
	}
}

func (h *routes) setDefault(def *proxy.ReverseProxyAdapter) {
	h.defaultProxy = def
}

func (h *routes) add(path string, proxy *proxy.ReverseProxyAdapter) {
	h.paths.Add(path, proxy)
}

func (h *routes) find(path string) (*proxy.ReverseProxyAdapter, bool) {
	p, ok := h.paths.LongestCommonPrefix(path)
	if !ok {
		return h.defaultProxy, h.defaultProxy == nil
	}
	return p, true
}
