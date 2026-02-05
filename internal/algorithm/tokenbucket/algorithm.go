package tokenbucket

import (
	"context"
	"gateway/internal/limiter"
	"time"

	"github.com/pkg/errors"
)

type tokenBucket struct {
	capacity int
	rate     float64
}

func newTokenBucket(capacity int, rate float64) *tokenBucket {
	return &tokenBucket{
		capacity: capacity,
		rate:     rate,
	}
}

func (tb *tokenBucket) Action(ctx context.Context, state *limiter.State) (*limiter.State, error) {
	p, err := parseParams(state.Params)
	if err != nil {
		return nil, errors.Wrap(limiter.ErrIvalidState, err.Error())
	}

	now := time.Now()
	elapsed := now.Sub(p.lastUpdate).Seconds()

	p.tokens += elapsed * tb.rate
	if p.tokens > float64(tb.capacity) {
		p.tokens = float64(tb.capacity)
	}
	p.lastUpdate = now

	allow := false
	if p.tokens >= 1 {
		p.tokens -= 1
		allow = true
	}

	return &limiter.State{Allow: allow, Params: p.toMap()}, nil
}
