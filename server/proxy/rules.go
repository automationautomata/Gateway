package proxy

import (
	"fmt"
	"gateway/config"
	"gateway/server/common"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type proxy struct {
	backend string
	*httputil.ReverseProxy
}

func newProxy(backend string, path string) (*proxy, error) {
	target, err := url.Parse(backend)
	if err != nil {
		return nil, err
	}
	httputil.NewSingleHostReverseProxy(target)
	defaultTransport := (http.DefaultTransport.(*http.Transport))
	return &proxy{
		backend: backend,
		ReverseProxy: &httputil.ReverseProxy{
			Rewrite: func(r *httputil.ProxyRequest) {
				r.SetURL(target)
				out, in := r.Out, r.In

				out.Host = in.Host
				out.URL.Path = strings.TrimPrefix(in.URL.Path, path)
				r.Out.Header.Set("X-Forwarded-Host", r.In.Host)
			},
			Transport: &http.Transport{
				Proxy:                 nil,
				DialContext:           defaultTransport.DialContext,
				MaxIdleConns:          defaultTransport.MaxConnsPerHost,
				IdleConnTimeout:       defaultTransport.IdleConnTimeout,
				TLSHandshakeTimeout:   defaultTransport.TLSHandshakeTimeout,
				ExpectContinueTimeout: defaultTransport.ExpectContinueTimeout,
			},
		},
	}, nil
}

type hostRule struct {
	defaultProxy *proxy
	pathRules    *common.SyncMap[string, *proxy]
}

type hostRulesMap struct {
	*common.SyncMap[string, *hostRule]
}

func newHostRulesMap(hosts []config.HostRules) (hostRules *hostRulesMap, err error) {
	hostRules = &hostRulesMap{common.NewSyncMap[string, *hostRule]()}

	for _, hostCfg := range hosts {
		r := &hostRule{pathRules: common.NewSyncMap[string, *proxy]()}

		if hostCfg.Default != nil {
			r.defaultProxy, err = newProxy(*hostCfg.Default, "/")
			if err != nil {
				return nil, err
			}
		}

		for path, backend := range hostCfg.Pathes {
			path = normalizePath(path)
			proxy, err := newProxy(backend, path)
			if err != nil {
				return nil, err
			}
			r.pathRules.Add(path, proxy)
		}

		hostRules.Add(hostCfg.Host, r)
	}

	return hostRules, nil
}

func normalizePath(path string) string {
	if strings.HasSuffix(path, "/") {
		return path
	}
	return fmt.Sprint(path, "/")
}
