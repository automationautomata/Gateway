package tokenbucket

import (
	"context"
	"encoding/json"
	"gateway/internal/limiter"
	"time"
)

type Params struct {
	Tokens     float64
	LastUpdate time.Time
}

func (s *Params) Marshal() ([]byte, error) { return json.Marshal(s) }

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

func (tb *tokenBucket) Action(ctx context.Context, state *limiter.State) (bool, *limiter.State, error) {
	p, ok := state.Params.(*Params)
	if !ok {
		return false, nil, limiter.ErrIvalidState
	}

	now := time.Now()
	elapsed := now.Sub(p.LastUpdate).Seconds()

	p.Tokens += elapsed * tb.rate
	if p.Tokens > float64(tb.capacity) {
		p.Tokens = float64(tb.capacity)
	}
	p.LastUpdate = now

	allow := false
	if p.Tokens >= 1 {
		p.Tokens -= 1
		allow = true
	}

	return allow, &limiter.State{Params: p}, nil
}
