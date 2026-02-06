package proxy

import (
	"gateway/config"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
)

type rules[K any, V any] struct {
	m *sync.Map
}

func newRules[K any, V any]() *rules[K, V] {
	var m sync.Map
	return &rules[K, V]{&m}
}

func (r *rules[K, V]) add(k string, v V) {
	r.m.Store(k, v)
}

func (r *rules[K, V]) get(k K) (V, bool) {
	p, ok := r.m.Load(k)
	if !ok {
		return *new(V), false
	}
	return p.(V), ok
}

type proxy struct {
	backend string
	*httputil.ReverseProxy
}

func newProxy(backend string, prefix string) (*proxy, error) {
	target, err := url.Parse(backend)
	if err != nil {
		return nil, err
	}
	return &proxy{
		backend: prefix,
		ReverseProxy: &httputil.ReverseProxy{
			Rewrite: func(r *httputil.ProxyRequest) {
				r.SetURL(target)
				out, in := r.Out, r.In

				out.Host = in.Host
				out.URL.Path = strings.TrimPrefix(in.URL.Path, prefix)
				// r.Out.Header.Set("X-Forwarded-Host", r.In.Host)
			},
		},
	}, nil
}

type hostRule struct {
	defaultProxy *proxy
	pathRules    *rules[string, *proxy]
}

type hostRulesMap struct {
	*rules[string, *hostRule]
}

func newHostRulesMap(hosts []config.HostRules) (*hostRulesMap, error) {
	hostRules := hostRulesMap{newRules[string, *hostRule]()}

	var err error
	for _, hostCfg := range hosts {
		r := &hostRule{pathRules: newRules[string, *proxy]()}

		if hostCfg.Default != nil {
			r.defaultProxy, err = newProxy(*hostCfg.Default, "/")
			if err != nil {
				return nil, err
			}
		}

		for prefix, backend := range hostCfg.Pathes {
			proxy, err := newProxy(backend, prefix)
			if err != nil {
				return nil, err
			}
			r.pathRules.add(prefix, proxy)
		}

		hostRules.add(hostCfg.Host, r)
	}

	return &hostRules, nil
}
