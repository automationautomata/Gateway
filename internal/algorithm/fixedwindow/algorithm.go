package fixedwindow

import (
	"encoding/json"
	"gateway/internal/limiter"
	"time"
)

type Params struct {
	WindowStart time.Time
	Count       int
}

func (p Params) Marshal() ([]byte, error) { return json.Marshal(p) }

type fixedWindow struct {
	limit     int
	windowDur time.Duration
}

func NewFixedWindow(limit int, windowDur time.Duration) *fixedWindow {
	return &fixedWindow{
		limit:     limit,
		windowDur: windowDur,
	}
}

func (fw *fixedWindow) FirstState() *limiter.State {
	return &limiter.State{Params: Params{time.Now(), 0}}
}

func (fw *fixedWindow) Action(state *limiter.State) (bool, *limiter.State, error) {
	p, ok := state.Params.(*Params)
	if !ok {
		return false, nil, limiter.ErrInvalidState
	}

	count, windowStart := p.Count, p.WindowStart
	now := time.Now()

	if now.Sub(windowStart) >= fw.windowDur {
		windowStart = now.Truncate(fw.windowDur)
		count = 0
	}

	allow := false
	if count < fw.limit {
		count++
		allow = true
	}

	return allow, &limiter.State{
		Params: &Params{
			WindowStart: windowStart,
			Count:       count,
		},
	}, nil
}
