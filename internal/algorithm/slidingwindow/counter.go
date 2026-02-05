package slidingwindow

import (
	"context"
	"gateway/internal/limiter"
	"time"

	"github.com/pkg/errors"
)

type slidingWindow struct {
	windowSize time.Duration
	bucketSize time.Duration
	bucketsNum int
	limit      int64
}

func newSlidingWindow(window time.Duration, bucketsNum int, limit int64) *slidingWindow {
	return &slidingWindow{
		windowSize: window,
		bucketSize: window / time.Duration(bucketsNum),
		bucketsNum: bucketsNum,
		limit:      limit,
	}
}

func (sw *slidingWindow) Action(ctx context.Context, state *limiter.State) (*limiter.State, error) {
	p, err := parseCounterParams(state.Params)
	if err != nil {
		return nil, errors.Wrap(limiter.ErrIvalidState, err.Error())
	}

	now := time.Now()
	currentBucketStart := now.Truncate(sw.bucketSize)

	// find or rotate bucket
	targetIndex := -1
	for i := 0; i < sw.bucketsNum; i++ {
		if p.bucketTimes[i].Equal(currentBucketStart) {
			targetIndex = i
			break
		}
	}

	if targetIndex == -1 {
		p.currentIndex = (p.currentIndex + 1) % sw.bucketsNum
		targetIndex = p.currentIndex

		p.buckets[targetIndex] = 0
		p.bucketTimes[targetIndex] = currentBucketStart
	}

	// clear expired buckets
	cutoff := now.Add(-sw.windowSize)
	for i := 0; i < sw.bucketsNum; i++ {
		if !p.bucketTimes[i].IsZero() && p.bucketTimes[i].Before(cutoff) {
			p.buckets[i] = 0
			p.bucketTimes[i] = time.Time{}
		}
	}

	// sum valid buckets
	var total int64
	for i := 0; i < sw.bucketsNum; i++ {
		start := p.bucketTimes[i]
		if start.IsZero() {
			continue
		}
		end := start.Add(sw.bucketSize)
		if end.After(cutoff) {
			total += p.buckets[i]
		}
	}

	if total >= sw.limit {
		return &limiter.State{
			Allow:  false,
			Params: p.toMap(),
		}, nil
	}

	p.buckets[targetIndex]++

	return &limiter.State{
		Allow:  true,
		Params: p.toMap(),
	}, nil
}
