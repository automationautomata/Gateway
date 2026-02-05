package slidingwindow

import (
	"gateway/config"
	"gateway/internal/limiter"
	"time"
)

func ProvideLogWindow(cfg *config.SlidingWindowLogSettings) (limiter.Algorithm, *limiter.State) {
	alg := newSlidingWindowLog(cfg.Limit, cfg.WindowDuration)
	p := logParams{[]time.Time{time.Now()}}

	firstState := &limiter.State{
		Allow:  true,
		Params: p.toMap(),
	}
	return alg, firstState
}

func ProvideCounterWindow(cfg *config.SlidingWindowCounterSettings) (limiter.Algorithm, *limiter.State) {
	alg := newSlidingWindowCounter(cfg.WindowDuration, cfg.BucketsNum, cfg.Limit)
	p := logParams{[]time.Time{time.Now()}}

	firstState := &limiter.State{
		Allow:  true,
		Params: p.toMap(),
	}
	return alg, firstState
}
