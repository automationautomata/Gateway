package fixedwindow

import (
	"context"
	"gateway/internal/limiter"
	"time"

	"github.com/pkg/errors"
)

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

func (fw *fixedWindow) Action(ctx context.Context, state *limiter.State) (*limiter.State, error) {
	p, err := parseParams(state.Params)
	if err != nil {
		return nil, errors.Wrap(limiter.ErrIvalidState, err.Error())
	}

	now := time.Now()
	diff := now.Sub(p.windowStart)

	if diff > fw.windowDur {
		p.windowStart = p.windowStart.Add(diff - diff%fw.windowDur)
		p.count = 0
	}

	allow := false
	if p.count < fw.limit {
		p.count++
		allow = true
	}
	return &limiter.State{Allow: allow, Params: p.toMap()}, nil
}
