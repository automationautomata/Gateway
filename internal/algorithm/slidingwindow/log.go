package slidingwindow

import (
	"context"
	"gateway/internal/limiter"
	"sort"
	"time"

	"github.com/pkg/errors"
)

type slidingWindowLog struct {
	windowDur time.Duration
	limit     int
}

func newSlidingWindowLog(limit int, windowDur time.Duration) *slidingWindowLog {
	return &slidingWindowLog{
		windowDur: windowDur, limit: limit,
	}
}

func (sw *slidingWindowLog) Action(ctx context.Context, state *limiter.State) (*limiter.State, error) {
	p, err := parseLogParams(state.Params)
	if err != nil {
		return nil, errors.Wrap(limiter.ErrIvalidState, err.Error())
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
	return &limiter.State{Allow: allow, Params: p.toMap()}, nil
}
