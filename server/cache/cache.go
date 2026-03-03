package cache

import (
	"errors"
	"gateway/server/interfaces"
	"gateway/server/pathstree"
	"gateway/server/urlutils"
	"net/http"
	"time"
)

type CacheMiddleware struct {
	paths  *pathstree.Tree[time.Duration]
	cache  interfaces.CacheStorage[*ResponseContent]
	metric interfaces.CacheMetric
	log    interfaces.Logger
}

func NewCacheMiddleware(
	paths map[string]time.Duration,
	metric interfaces.CacheMetric,
	cache interfaces.CacheStorage[*ResponseContent],
	log interfaces.Logger,
) *CacheMiddleware {
	tree := pathstree.New[time.Duration]()
	for p, ttl := range paths {
		tree.Add(p, ttl)
	}
	return &CacheMiddleware{tree, cache, metric, log}
}

func (c *CacheMiddleware) isCached(r *http.Request) (time.Duration, bool) {
	ttl, ok := c.paths.Find(r.URL.Path)
	if !ok || r.Method != http.MethodGet {
		return 0, false
	}
	return ttl, true
}

func (c *CacheMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ttl, ok := c.isCached(r)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			ok = c.serveCache(r, w)
			if ok {
				return
			}

			bufWriter := newBufferedResponseWriter(w)
			next.ServeHTTP(bufWriter, r)

			resp := bufWriter.toResponseContent()
			err := resp.copyTo(w)
			if err != nil {
				http.Error(w, "", http.StatusInternalServerError)
				c.logErr(r, "cannot copy response", err, false)
				return
			}

			err = c.cache.Set(r.Context(), r.URL.String(), resp, ttl)
			if err != nil {
				c.logErr(r, "cache set failed", err, true)
			}
		},
	)
}

func (c *CacheMiddleware) serveCache(r *http.Request, w http.ResponseWriter) bool {
	resp, err := c.cache.Get(r.Context(), r.URL.String())
	if err == nil {
		c.metric.Inc(urlutils.GetHost(r), r.URL.Path, r.URL.Query().Encode(), true)
		err = resp.copyTo(w)
		if err != nil {
			c.logErr(r, "cannot copy response", err, true)
			return false
		}
		return false
	}

	if !errors.Is(err, interfaces.ErrCacheNotFound) {
		c.logErr(r, "cache get failed", err, true)
		return false
	}
	c.metric.Inc(urlutils.GetHost(r), r.URL.Path, r.URL.Query().Encode(), false)
	return false
}

func (c *CacheMiddleware) logErr(r *http.Request, msg string, err error, isWarn bool) {
	fields := map[string]any{
		"url":   r.URL.String(),
		"error": err,
	}
	if isWarn {
		c.log.Warn(r.Context(), msg, fields)
		return
	}
	c.log.Error(r.Context(), msg, fields)
}
