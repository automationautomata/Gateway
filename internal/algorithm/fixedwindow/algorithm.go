package fixedwindow

import (
	"context"
	"encoding/json"
	"gateway/internal/limiter"
	"time"
)

type Params struct {
	WindowStart time.Time
	Count       int
}

func (p *Params) Marshal() ([]byte, error) { return json.Marshal(p) }

type fixedWindow struct {
	limit     int
	windowDur time.Duration
}

func newFixedWindow(limit int, windowDur time.Duration) *fixedWindow {
	return &fixedWindow{
		limit:     limit,
		windowDur: windowDur,
	}
}

func (fw *fixedWindow) Action(ctx context.Context, state *limiter.State) (bool, *limiter.State, error) {
	p, ok := state.Params.(*Params)
	if !ok {
		return false, nil, limiter.ErrIvalidState
	}
	count, windowStart := p.Count, p.WindowStart

	now := time.Now()
	diff := now.Sub(windowStart)

	if diff > fw.windowDur {
		windowStart = windowStart.Add(diff - diff%fw.windowDur)
		count = 0
	}

	allow := false
	if count < fw.limit {
		count++
		allow = true
	}
	return allow, &limiter.State{Params: &Params{windowStart, count}}, nil
}
