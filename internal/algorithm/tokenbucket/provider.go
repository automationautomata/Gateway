package tokenbucket

import (
	"gateway/config"
	"gateway/internal/limiter"
	"time"
)

func Provide(cfg *config.TokenBucketSettings) (limiter.Algorithm, *limiter.State) {
	alg := newTokenBucket(cfg.Capacity, cfg.Rate)
	p := &Params{float64(cfg.Capacity), time.Now()}

	return alg, &limiter.State{Params: p}
}
