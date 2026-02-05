package slidingwindow

import (
	"gateway/config"
	"gateway/internal/limiter"
	"time"
)

func ProvideLogWindow(cfg *config.SlidingWindowLogConfig) (limiter.Algorithm, *limiter.State) {
	alg := newSlidingWindowLog(cfg.Limit, cfg.WindowDuration)
	p := logParams{[]time.Time{time.Now()}}

	firstState := &limiter.State{
		Allow:  true,
		Params: p.toMap(),
	}
	return alg, firstState
}
