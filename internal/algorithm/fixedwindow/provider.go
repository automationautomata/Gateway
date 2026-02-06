package fixedwindow

import (
	"gateway/config"
	"gateway/internal/limiter"
	"time"
)

func Provide(cfg *config.FixedWindowSettings) (limiter.Algorithm, *limiter.State) {
	alg := newFixedWindow(cfg.Limit, cfg.WindowDuration)
	firstState := &limiter.State{
		Params: &Params{time.Now(), 0},
	}
	return alg, firstState
}
