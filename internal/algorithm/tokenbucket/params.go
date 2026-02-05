package tokenbucket

import (
	"fmt"
	"time"
)

type params struct {
	tokens     float64
	lastUpdate time.Time
}

func (p params) toMap() map[string]any {
	return map[string]any{
		"tokens":      p.tokens,
		"last_refill": p.lastUpdate,
	}
}

func parseParams(raw map[string]any) (p params, err error) {
	tokensVal, ok := raw["tokens"]
	if !ok {
		return params{}, fmt.Errorf("tokens not in state")
	}

	p.tokens, ok = tokensVal.(float64)
	if !ok {
		return params{}, fmt.Errorf("tokens invalid type")
	}

	lastUpdateVal, ok := raw["last_refill"]
	if !ok {
		return params{}, fmt.Errorf("last_refill not in state")
	}

	p.lastUpdate, ok = lastUpdateVal.(time.Time)
	if !ok {
		return params{}, fmt.Errorf("last_refill invalid type")
	}

	return p, nil
}
