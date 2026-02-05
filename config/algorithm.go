package config

import (
	"time"

	"gopkg.in/yaml.v3"
)

type AlgorithmType string

const (
	FixedWindowType          AlgorithmType = "fixed_window"
	SlidingWindowCounterType AlgorithmType = "sliding_window_counter"
	SlidingWindowLogType     AlgorithmType = "sliding_window_log"
	TokenBucketType          AlgorithmType = "token_bucket"
)

type TokenBucketConfig struct {
	Capacity int     `yaml:"capacity"`
	Rate     float64 `yaml:"rate"`
}

type FixedWindowConfig struct {
	Limit          int           `yaml:"limit"`
	WindowDuration time.Duration `yaml:"window_duration"`
}

type SlidingWindowLogConfig struct {
	Limit          int           `yaml:"limit"`
	WindowDuration time.Duration `yaml:"window_duration"`
}

type SlidingWindowCounterConfig struct {
	Limit          int64         `yaml:"limit"`
	WindowDuration time.Duration `yaml:"window_duration"`
	BucketsNum     int           `yaml:"window_duration"`
}

type algorithmConfig yaml.Node

type LimiterConfig struct {
	LimiterType     AlgorithmType   `yaml:"limiter_type"`
	AlgorithmConfig algorithmConfig `yaml:"action"`
}

type AlgorithmConfigTypes interface {
	TokenBucketConfig |
		FixedWindowConfig |
		SlidingWindowLogConfig |
		SlidingWindowCounterConfig
}

func DecodeAlgorithmConfig[T AlgorithmConfigTypes](c algorithmConfig, target *T) error {
	node := yaml.Node(c)
	if err := node.Decode(target); err != nil {
		return err
	}
	return nil
}
