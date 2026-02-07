package slidingwindow

import (
	"encoding/json"
	"gateway/internal/limiter"
	"time"
)

type CounterParams struct {
	buckets      []int64
	bucketTimes  []time.Time
	currentIndex int
}

func (p CounterParams) Marshal() ([]byte, error) { return json.Marshal(p) }

type slidingWindowCounter struct {
	windowSize time.Duration
	bucketSize time.Duration
	bucketsNum int
	limit      int64
}

func NewSlidingWindowCounter(window time.Duration, bucketsNum int, limit int64) *slidingWindowCounter {
	return &slidingWindowCounter{
		windowSize: window,
		bucketSize: window / time.Duration(bucketsNum),
		bucketsNum: bucketsNum,
		limit:      limit,
	}
}

func (sw *slidingWindowCounter) FirstState() *limiter.State {
	return &limiter.State{
		Params: &CounterParams{
			buckets:      make([]int64, sw.bucketsNum),
			bucketTimes:  make([]time.Time, sw.bucketsNum),
			currentIndex: 0,
		},
	}
}

func (sw *slidingWindowCounter) Action(state *limiter.State) (bool, *limiter.State, error) {
	p, ok := state.Params.(*CounterParams)
	if !ok {
		return false, nil, limiter.ErrInvalidState
	}

	now := time.Now()
	currentBucketStart := now.Truncate(sw.bucketSize)

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

	cutoff := now.Add(-sw.windowSize)
	for i := 0; i < sw.bucketsNum; i++ {
		if !p.bucketTimes[i].IsZero() && p.bucketTimes[i].Before(cutoff) {
			p.buckets[i] = 0
			p.bucketTimes[i] = time.Time{}
		}
	}

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
		return false, &limiter.State{Params: p}, nil
	}

	p.buckets[targetIndex]++
	return true, &limiter.State{Params: p}, nil
}
