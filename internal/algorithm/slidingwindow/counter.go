package slidingwindow

import (
	"encoding/json"
	"gateway/internal/limiter"
	"time"
)

type CounterParams struct {
	Buckets      []int64
	BucketTimes  []time.Time
	CurrentIndex int
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
			Buckets:      make([]int64, sw.bucketsNum),
			BucketTimes:  make([]time.Time, sw.bucketsNum),
			CurrentIndex: 0,
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
		if p.BucketTimes[i].Equal(currentBucketStart) {
			targetIndex = i
			break
		}
	}

	if targetIndex == -1 {
		p.CurrentIndex = (p.CurrentIndex + 1) % sw.bucketsNum
		targetIndex = p.CurrentIndex

		p.Buckets[targetIndex] = 0
		p.BucketTimes[targetIndex] = currentBucketStart
	}

	cutoff := now.Add(-sw.windowSize)
	for i := 0; i < sw.bucketsNum; i++ {
		if !p.BucketTimes[i].IsZero() && p.BucketTimes[i].Before(cutoff) {
			p.Buckets[i] = 0
			p.BucketTimes[i] = time.Time{}
		}
	}

	var total int64
	for i := 0; i < sw.bucketsNum; i++ {
		start := p.BucketTimes[i]
		if start.IsZero() {
			continue
		}
		end := start.Add(sw.bucketSize)
		if end.After(cutoff) {
			total += p.Buckets[i]
		}
	}

	if total >= sw.limit {
		return false, &limiter.State{Params: p}, nil
	}

	p.Buckets[targetIndex]++
	return true, &limiter.State{Params: p}, nil
}
