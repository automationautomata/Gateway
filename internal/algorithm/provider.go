package algorithm

import (
	"gateway/config"
	"gateway/internal/algorithm/fixedwindow"
	"gateway/internal/algorithm/slidingwindow"
	"gateway/internal/algorithm/tokenbucket"
	"gateway/internal/limiter"
)

func ProvideAlgorithmFactory(cfg config.LimiterConfig) (*limiter.AlgorithmFactory, error) {
	var (
		alg        limiter.Algorithm
		firstState *limiter.State
	)

	switch cfg.LimiterType {
	case config.TokenBucketType:
		algConf := &config.TokenBucketConfig{}
		err := config.DecodeAlgorithmConfig(cfg.AlgorithmConfig, algConf)
		if err != nil {
			return nil, err
		}
		alg, firstState = tokenbucket.Provide(algConf)

	case config.FixedWindowType:
		algConf := &config.FixedWindowConfig{}
		err := config.DecodeAlgorithmConfig(cfg.AlgorithmConfig, algConf)
		if err != nil {
			return nil, err
		}
		alg, firstState = fixedwindow.Provide(algConf)

	case config.SlidingWindowLogType:
		algConf := &config.SlidingWindowLogConfig{}
		err := config.DecodeAlgorithmConfig(cfg.AlgorithmConfig, algConf)
		if err != nil {
			return nil, err
		}
		alg, firstState = slidingwindow.ProvideLogWindow(algConf)

	case config.SlidingWindowCounterType:
		algConf := &config.SlidingWindowCounterConfig{}
		err := config.DecodeAlgorithmConfig(cfg.AlgorithmConfig, algConf)
		if err != nil {
			return nil, err
		}

	}

	return limiter.NewFactory(string(cfg.LimiterType), alg, firstState), nil
}
