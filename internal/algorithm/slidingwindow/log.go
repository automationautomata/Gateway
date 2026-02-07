package slidingwindow

import (
	"encoding/json"
	"gateway/internal/limiter"
	"sort"
	"time"
)

type LogParams struct {
	logs []time.Time
}

func (p LogParams) Marshal() ([]byte, error) { return json.Marshal(p) }

type slidingWindowLog struct {
	windowDur time.Duration
	limit     int
}

func newSlidingWindowLog(limit int, windowDur time.Duration) *slidingWindowLog {
	return &slidingWindowLog{
		windowDur: windowDur, limit: limit,
	}
}

func (fw *slidingWindowLog) FirstState() *limiter.State {
	return &limiter.State{
		Params: &LogParams{[]time.Time{}},
	}
}

func (sw *slidingWindowLog) Action(state *limiter.State) (bool, *limiter.State, error) {
	p, ok := state.Params.(*LogParams)
	if !ok {
		return false, nil, limiter.ErrInvalidState
	}

	now := time.Now()
	windowEnd := now.Add(-sw.windowDur)
	ind := sort.Search(len(p.logs), func(i int) bool {
		return p.logs[i].After(windowEnd)
	})

	if ind > 0 {
		copy(p.logs, p.logs[ind:])
		p.logs = p.logs[:len(p.logs)-ind]
	}

	allow := false
	if len(p.logs) < sw.limit {
		p.logs = append(p.logs, now)
		allow = true
	}
	return allow, &limiter.State{Params: p}, nil
}
