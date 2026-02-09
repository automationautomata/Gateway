package middlewares

import (
	"fmt"
	"gateway/server/common"
	"gateway/server/interfaces"
	"net/http"
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
	log    interfaces.Logger

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
func NewRateLimiter(lim interfaces.Limiter, log interfaces.Logger, options ...RateLimiterOption) *RateLimiter {
	rl := &RateLimiter{metric: nil, lim: lim, keyType: IP, log: log}
	for _, opt := range options {
		opt(rl)
	}
	return rl
}

func (rl *RateLimiter) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var key string

			ip := common.GetIP(r)
			switch rl.keyType {
			case Global:
				key = globalKey
			case IP:
				key = ip
			}

			allow, err := rl.lim.Allow(r.Context(), key)
			if err != nil {
				msg := fmt.Sprintf(
					"rate limiter failed from %s to %s", ip, r.URL.String(),
				)
				rl.log.Error(r.Context(), msg, map[string]any{"error": err})
				return
			}
			rl.log.Debug(r.Context(), "rate limiter", map[string]any{
				"from": ip, "to": common.GetHost(r), "allowed": allow,
			})

			rl.metric.Inc(allow, key)
			if !allow {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		},
	)
}
