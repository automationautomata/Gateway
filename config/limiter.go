package config

import (
	"fmt"
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

type StorageSettings struct {
	KeyTTL time.Duration `yaml:"ttl,omitempty"`
}

type LimiterSettings struct {
	Storage   *StorageSettings `yaml:"storage,omitempty"`
	Type      AlgorithmType    `yaml:"type"`
	Algorithm any              `yaml:"algorithm"`
}

func (l *LimiterSettings) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("limiter must be a mapping")
	}

	var n yaml.Node
	if err := node.Decode(&n); err != nil {
		return err
	}
	var raw struct {
		Storage   *StorageSettings `yaml:"storage,omitempty"`
		Type      AlgorithmType    `yaml:"type"`
		Algorithm yaml.Node        `yaml:"algorithm"`
	}
	if err := n.Decode(&raw); err != nil {
		return err
	}
	l.Type, l.Storage = raw.Type, raw.Storage

	switch l.Type {
	case FixedWindowAlgorithm:
		var cfg FixedWindowSettings
		if err := raw.Algorithm.Decode(&cfg); err != nil {
			return fmt.Errorf("failed to decode fixed_window algorithm: %w", err)
		}
		l.Algorithm = &cfg

	case SlidingWindowLogAlgorithm:
		var cfg SlidingWindowLogSettings
		if err := raw.Algorithm.Decode(&cfg); err != nil {
			return fmt.Errorf("failed to decode sliding_window_log algorithm: %w", err)
		}
		l.Algorithm = &cfg

	case SlidingWindowCounterAlgorithm:
		var cfg SlidingWindowCounterSettings
		if err := raw.Algorithm.Decode(&cfg); err != nil {
			return fmt.Errorf("failed to decode sliding_window_counter algorithm: %w", err)
		}
		l.Algorithm = &cfg

	case TokenBucketAlgorithm:
		var cfg TokenBucketSettings
		if err := raw.Algorithm.Decode(&cfg); err != nil {
			return fmt.Errorf("failed to decode token_bucket algorithm: %w", err)
		}
		l.Algorithm = &cfg

	default:
		return fmt.Errorf("unknown algorithm type: %s", l.Type)
	}

	return nil
}
