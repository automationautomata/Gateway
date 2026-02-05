package slidingwindow

import (
	"fmt"
	"time"
)

type counterParams struct {
	buckets      []int64
	bucketTimes  []time.Time
	currentIndex int
}

func (p counterParams) toMap() map[string]any {
	return map[string]any{
		"buckets":       p.buckets,
		"bucket_times":  p.bucketTimes,
		"current_index": p.currentIndex,
	}
}

func parseCounterParams(raw map[string]any) (p counterParams, err error) {
	bucketsVal, ok := raw["buckets"]
	if !ok {
		return p, fmt.Errorf("buckets not in state")
	}
	p.buckets, ok = bucketsVal.([]int64)
	if !ok {
		return p, fmt.Errorf("buckets invalid type")
	}

	timesVal, ok := raw["bucket_times"]
	if !ok {
		return p, fmt.Errorf("bucket_times not in state")
	}
	p.bucketTimes, ok = timesVal.([]time.Time)
	if !ok {
		return p, fmt.Errorf("bucket_times invalid type")
	}

	indexVal, ok := raw["current_index"]
	if !ok {
		return p, fmt.Errorf("current_index not in state")
	}
	p.currentIndex, ok = indexVal.(int)
	if !ok {
		return p, fmt.Errorf("current_index invalid type")
	}

	return p, nil
}

type logParams struct {
	logs []time.Time
}

func (p logParams) toMap() map[string]any {
	return map[string]any{
		"logs": p.logs,
	}
}

func parseLogParams(raw map[string]any) (p logParams, err error) {
	logsVal, ok := raw["logs"]
	if !ok {
		return logParams{}, fmt.Errorf("count not in state")
	}

	p.logs, ok = logsVal.([]time.Time)
	if !ok {
		return logParams{}, fmt.Errorf("count not in state")
	}

	return p, nil
}
