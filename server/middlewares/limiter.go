package middlewares

import (
	"gateway/server/interfaces"
	"log"
	"net"
	"net/http"
	"strings"
)

type KeyType string

const (
	Global KeyType = "global"
	IP     KeyType = "IP"

	globalKey = "global"
)

type RateLimiter struct {
	lim     interfaces.Limiter
	keyType KeyType
}

func NewRateLimiter(lim interfaces.Limiter, keyType KeyType) *RateLimiter {
	return &RateLimiter{lim, keyType}
}

func (mw *RateLimiter) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var key string

			switch mw.keyType {
			case Global:
				key = globalKey
			case IP:
				key = getClientIP(r)
			}

			allow, err := mw.lim.Allow(r.Context(), key)
			if err != nil {
				log.Printf(
					"rate limiter failed from %s to %s: %s",
					getClientIP(r), r.URL.String(), err,
				)
			}

			if !allow {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		},
	)
}

func getClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		return strings.Split(ip, ",")[0]
	}

	ip = r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
