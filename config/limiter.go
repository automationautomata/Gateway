package config

import (
	"time"

	"gopkg.in/yaml.v3"
)

type AlgorithmType string

const (
	FixedWindowAlgorithm          AlgorithmType = "fixed_window"
	SlidingWindowCounterAlgorithm AlgorithmType = "sliding_window_counter"
	SlidingWindowLogAlgorithm     AlgorithmType = "sliding_window_log"
	TokenBucketAlgorithm          AlgorithmType = "token_bucket"
)

type TokenBucketSettings struct {
	Capacity int     `yaml:"capacity"`
	Rate     float64 `yaml:"rate"`
}

type FixedWindowSettings struct {
	Limit          int           `yaml:"limit"`
	WindowDuration time.Duration `yaml:"window_duration"`
}

type SlidingWindowLogSettings struct {
	Limit          int           `yaml:"limit"`
	WindowDuration time.Duration `yaml:"window_duration"`
}

type SlidingWindowCounterSettings struct {
	Limit          int64         `yaml:"limit"`
	WindowDuration time.Duration `yaml:"window_duration"`
	BucketsNum     int           `yaml:"buckets_number"`
}

type RawAlgorithmSettings yaml.Node

type AlgorithmSettings struct {
	LimiterType AlgorithmType        `yaml:"limiter_type"`
	Algorithm   RawAlgorithmSettings `yaml:"algorithm"`
}

type StorageSettings struct {
	URL string         `yaml:"url"`
	TTL *time.Duration `yaml:"ttl,omitempty"`
}

type LimiterSettings struct {
	Storage StorageSettings `yaml:"storage"`
	AlgorithmSettings
}

type AlgorithmSettingsTypes interface {
	TokenBucketSettings |
		FixedWindowSettings |
		SlidingWindowLogSettings |
		SlidingWindowCounterSettings
}

func DecodeAlgorithmSettings[T AlgorithmSettingsTypes](c RawAlgorithmSettings, target *T) error {
	node := yaml.Node(c)
	if err := node.Decode(target); err != nil {
		return err
	}
	return nil
}
