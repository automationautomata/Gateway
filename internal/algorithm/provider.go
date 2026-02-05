package algorithm

import (
	"gateway/config"
	"gateway/internal/algorithm/fixedwindow"
	"gateway/internal/algorithm/slidingwindow"
	"gateway/internal/algorithm/tokenbucket"
	"gateway/internal/limiter"
)

func ProvideAlgorithmFactory(cfg config.AlgorithmSettings) (*limiter.AlgorithmFactory, error) {
	var (
		alg        limiter.Algorithm
		firstState *limiter.State
	)

	switch cfg.LimiterType {
	case config.TokenBucketAlgorithm:
		algConf := &config.TokenBucketSettings{}
		err := config.DecodeAlgorithmSettings(cfg.Algorithm, algConf)
		if err != nil {
			return nil, err
		}
		alg, firstState = tokenbucket.Provide(algConf)

	case config.FixedWindowAlgorithm:
		algConf := &config.FixedWindowSettings{}
		err := config.DecodeAlgorithmSettings(cfg.Algorithm, algConf)
		if err != nil {
			return nil, err
		}
		alg, firstState = fixedwindow.Provide(algConf)

	case config.SlidingWindowLogAlgorithm:
		algConf := &config.SlidingWindowLogSettings{}
		err := config.DecodeAlgorithmSettings(cfg.Algorithm, algConf)
		if err != nil {
			return nil, err
		}
		alg, firstState = slidingwindow.ProvideLogWindow(algConf)

	case config.SlidingWindowCounterAlgorithm:
		algConf := &config.SlidingWindowCounterSettings{}
		err := config.DecodeAlgorithmSettings(cfg.Algorithm, algConf)
		if err != nil {
			return nil, err
		}

	}

	return limiter.NewFactory(string(cfg.LimiterType), alg, firstState), nil
}
