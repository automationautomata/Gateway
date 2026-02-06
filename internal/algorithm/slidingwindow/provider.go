package slidingwindow

import (
	"gateway/config"
	"gateway/internal/limiter"
	"time"
)

func ProvideLogWindow(cfg *config.SlidingWindowLogSettings) (limiter.Algorithm, *limiter.State) {
	alg := newSlidingWindowLog(cfg.Limit, cfg.WindowDuration)

	p := &LogParams{[]time.Time{time.Now()}}
	return alg, &limiter.State{Params: p}
}

func ProvideCounterWindow(cfg *config.SlidingWindowCounterSettings) (limiter.Algorithm, *limiter.State) {
	alg := newSlidingWindowCounter(cfg.WindowDuration, cfg.BucketsNum, cfg.Limit)

	p := &CounterParams{
		buckets:      make([]int64, cfg.BucketsNum),
		bucketTimes:  make([]time.Time, cfg.BucketsNum),
		currentIndex: 0,
	}
	return alg, &limiter.State{Params: p}
}
