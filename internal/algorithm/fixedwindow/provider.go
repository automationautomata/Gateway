package fixedwindow

import (
	"gateway/config"
	"gateway/internal/limiter"
	"time"
)

func Provide(cfg *config.FixedWindowConfig) (limiter.Algorithm, *limiter.State) {
	alg := newFixedWindow(cfg.Limit, cfg.WindowDuration)
	firstState := &limiter.State{
		Allow:  true,
		Params: (params{time.Now(), 0}).toMap(),
	}
	return alg, firstState
}
