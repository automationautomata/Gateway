package limiter

import (
	"fmt"
	"gateway/server/common"
	"gateway/server/interfaces"
	"net/http"
)

type KeyType string

type ContextKey string

const (
	globalKey = "global"

	Global       KeyType = "global"
	IP           KeyType = "IP"
	ContextValue KeyType = "context"

	LimiterContextKey ContextKey = "limiter"
)

type RateLimiter struct {
	lim interfaces.Limiter
	log interfaces.Logger

	// IP - по умолчанию
	keyType KeyType

	// nil - по умолчанию
	metric interfaces.LimiterMetric
}

type Option func(*RateLimiter)

func WithMetric(metric interfaces.LimiterMetric) Option {
	return func(rl *RateLimiter) {
		rl.metric = metric
	}
}

func WithKeyType(keyType KeyType) Option {
	return func(rl *RateLimiter) {
		rl.keyType = keyType
	}
}

// По умолчанию: keyType = IP, metric - nil
func NewRateLimiter(lim interfaces.Limiter, log interfaces.Logger, options ...Option) *RateLimiter {
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
			case ContextValue:
				key = r.Context().Value(LimiterContextKey).(string)
			}

			allow, err := rl.lim.Allow(r.Context(), key)
			if err != nil {
				rl.log.Error(
					r.Context(),
					fmt.Sprintf("rate limiter failed from %s to %s", ip, r.URL.String()),
					map[string]any{"error": err},
				)
				return
			}
			rl.log.Debug(
				r.Context(),
				"handle request",
				map[string]any{"from": ip, "to": common.GetHost(r), "allowed": allow},
			)

			rl.metric.Inc(allow, key)
			if !allow {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		},
	)
}
