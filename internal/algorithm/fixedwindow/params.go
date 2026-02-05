package fixedwindow

import (
	"fmt"
	"time"
)

type params struct {
	windowStart time.Time
	count       int
}

func (p params) toMap() map[string]any {
	return map[string]any{
		"count":        p.count,
		"window_start": p.windowStart,
	}
}

func parseParams(raw map[string]any) (p params, err error) {
	countVal, ok := raw["count"]
	if !ok {
		return params{}, fmt.Errorf("count not in state")
	}

	p.count, ok = countVal.(int)
	if !ok {
		return params{}, fmt.Errorf("count not in state")
	}

	windowStartVal, ok := raw["window_start"]
	if !ok {
		return params{}, fmt.Errorf("windowStart not in state")
	}
	p.windowStart, ok = windowStartVal.(time.Time)

	return p, nil
}
