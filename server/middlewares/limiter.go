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
	metric interfaces.LimiterMetric
	lim    interfaces.Limiter

	// IP by default
	keyType KeyType
}

type RateLimiterOption func(*RateLimiter)

func WithMetric(metric interfaces.LimiterMetric) RateLimiterOption {
	return func(rl *RateLimiter) {
		rl.metric = metric
	}
}

func WithKeyType(keyType KeyType) RateLimiterOption {
	return func(rl *RateLimiter) {
		rl.keyType = keyType
	}
}

// by default: keyType is IP, no metrics
func NewRateLimiter(lim interfaces.Limiter, options ...RateLimiterOption) *RateLimiter {
	rl := &RateLimiter{metric: nil, lim: lim, keyType: IP}
	for _, opt := range options {
		opt(rl)
	}
	return rl
}

func (rl *RateLimiter) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var key string

			switch rl.keyType {
			case Global:
				key = globalKey
			case IP:
				key = getClientIP(r)
			}

			allow, err := rl.lim.Allow(r.Context(), key)
			if err != nil {
				log.Printf(
					"rate limiter failed from %s to %s: %s",
					getClientIP(r), r.URL.String(), err,
				)
			}

			rl.metric.Inc(allow, key)
			if allow { ///
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
