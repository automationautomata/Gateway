package slidingwindow

import (
	"encoding/json"
	"gateway/internal/limiter"
	"sort"
	"time"
)

type LogParams struct {
	Logs []time.Time
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
	ind := sort.Search(len(p.Logs), func(i int) bool {
		return p.Logs[i].After(windowEnd)
	})

	if ind > 0 {
		copy(p.Logs, p.Logs[ind:])
		p.Logs = p.Logs[:len(p.Logs)-ind]
	}

	allow := false
	if len(p.Logs) < sw.limit {
		p.Logs = append(p.Logs, now)
		allow = true
	}
	return allow, &limiter.State{Params: p}, nil
}
