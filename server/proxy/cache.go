package proxy

import (
	"bytes"
	"encoding/json"
	"errors"
	"gateway/server/common"
	"gateway/server/interfaces"
	"net/http"
	"net/url"
	"time"
)

var ErrRequestFailed = errors.New("request failed")

type ResponseContent struct {
	StatusCode int
	Headers    map[string][]string
	Data       []byte
}

func (r *ResponseContent) MarshalJSON() ([]byte, error) { return json.Marshal(r) }

func (r *ResponseContent) copyTo(w http.ResponseWriter) {
	for k, vv := range r.Headers {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.Write(r.Data)
}

type responseBodyWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (rw *responseBodyWriter) Write(b []byte) (int, error) {
	if _, err := rw.body.Write(b); err != nil {
		return 0, err
	}
	return rw.ResponseWriter.Write(b)
}

func (rw *responseBodyWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

type cacheMiddleware struct {
	requestsToCache map[string]time.Duration
	cache           interfaces.Cache[*ResponseContent]
	metric          interfaces.CacheMetric
}

func (c *cacheMiddleware) isCached(url *url.URL) bool {
	_, ok := c.requestsToCache[url.Path]
	return ok
}

func (c *cacheMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet || !c.isCached(r.URL) {
				next.ServeHTTP(w, r)
				return
			}

			resp, err := c.cache.Get(
				r.Context(), r.URL.String(), func() (*ResponseContent, time.Duration, error) {
					bodyWriter := &responseBodyWriter{w, 0, &bytes.Buffer{}}
					next.ServeHTTP(bodyWriter, r)
					if bodyWriter.statusCode >= 400 {
						return nil, 0, ErrRequestFailed
					}

					c.metric.Inc(common.GetHost(r), r.URL.Path, r.URL.RawQuery, false)
					resp := &ResponseContent{
						StatusCode: bodyWriter.statusCode,
						Headers:    bodyWriter.Header().Clone(),
						Data:       bodyWriter.body.Bytes(),
					}
					return resp, c.requestsToCache[r.URL.Path], nil
				},
			)

			if err != nil {
				if !errors.Is(err, ErrRequestFailed) {
					http.Error(w, "", http.StatusInternalServerError)
				}
				return
			}

			c.metric.Inc(common.GetHost(r), r.URL.Path, r.URL.RawQuery, true)

			resp.copyTo(w)
		},
	)
}
