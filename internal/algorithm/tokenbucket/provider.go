package tokenbucket

import (
	"gateway/config"
	"gateway/internal/limiter"
	"time"
)

func Provide(cfg *config.TokenBucketConfig) (limiter.Algorithm, *limiter.State) {
	alg := newTokenBucket(cfg.Capacity, cfg.Rate)
	p := params{
		tokens:     float64(cfg.Capacity),
		lastUpdate: time.Now(),
	}

	firstState := &limiter.State{
		Allow:  true,
		Params: p.toMap(),
	}
	return alg, firstState
}
