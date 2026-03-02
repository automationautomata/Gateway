package cache

import (
	"context"
	"errors"
	"gateway/server/interfaces"
	"gateway/server/pathstree"
	"gateway/server/urlutils"
	"net/http"
	"time"
)

type requestFailedErr struct {
	resp *ResponseContent
}

func (requestFailedErr) Error() string { return "request failed" }

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

			hit := true
			resp, err := c.cache.Get(
				r.Context(),
				r.URL.String(),
				func(ctx context.Context) (*ResponseContent, time.Duration, error) {
					bufWriter := newBufferedBodyWriter(w)
					next.ServeHTTP(bufWriter, r.Clone(ctx))
					resp := bufWriter.toResponseContent()
					hit = false

					if bufWriter.statusCode >= 400 {
						return nil, 0, requestFailedErr{resp}
					}
					return resp, ttl, nil
				},
			)

			var reqErr requestFailedErr
			if errors.As(err, &reqErr) {
				resp = reqErr.resp
			} else if err != nil {
				http.Error(w, "", http.StatusInternalServerError)

				c.log.Error(
					r.Context(),
					"cache error",
					map[string]any{
						"host":  urlutils.GetHost(r),
						"path":  r.URL.Path,
						"query": r.URL.RawQuery,
						"error": err,
					},
				)
				return
			}

			c.metric.Inc(urlutils.GetHost(r), r.URL.Path, r.URL.RawQuery, hit)

			if err = resp.copyTo(w); err != nil {
				http.Error(w, "", http.StatusInternalServerError)
				return
			}
		},
	)
}
