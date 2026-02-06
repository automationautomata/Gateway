package algorithm

import (
	"gateway/config"
	"gateway/internal/algorithm/fixedwindow"
	"gateway/internal/algorithm/slidingwindow"
	"gateway/internal/algorithm/tokenbucket"
	"gateway/internal/limiter"
)

func ProvideAlgorithmFacade(cfg config.AlgorithmSettings) (*limiter.AlgorithmFacade, error) {
	var (
		alg        limiter.Algorithm
		firstState *limiter.State
		unmarsh    limiter.Unmarshaler[limiter.State]
	)

	switch cfg.Type {
	case config.TokenBucketAlgorithm:
		algConf := &config.TokenBucketSettings{}
		err := config.DecodeAlgorithmSettings(cfg.Algorithm, algConf)
		if err != nil {
			return nil, err
		}
		alg, firstState = tokenbucket.Provide(algConf)
		unmarsh = &unmarshaler[*tokenbucket.Params]{}

	case config.FixedWindowAlgorithm:
		algConf := &config.FixedWindowSettings{}
		err := config.DecodeAlgorithmSettings(cfg.Algorithm, algConf)
		if err != nil {
			return nil, err
		}
		alg, firstState = fixedwindow.Provide(algConf)
		unmarsh = &unmarshaler[*fixedwindow.Params]{}

	case config.SlidingWindowLogAlgorithm:
		algConf := &config.SlidingWindowLogSettings{}
		err := config.DecodeAlgorithmSettings(cfg.Algorithm, algConf)
		if err != nil {
			return nil, err
		}
		alg, firstState = slidingwindow.ProvideLogWindow(algConf)
		unmarsh = &unmarshaler[*slidingwindow.LogParams]{}

	case config.SlidingWindowCounterAlgorithm:
		algConf := &config.SlidingWindowCounterSettings{}
		err := config.DecodeAlgorithmSettings(cfg.Algorithm, algConf)
		if err != nil {
			return nil, err
		}
		alg, firstState = slidingwindow.ProvideCounterWindow(algConf)
		unmarsh = &unmarshaler[*slidingwindow.CounterParams]{}

	}

	facade := limiter.NewFacade(string(cfg.Type), alg, firstState, unmarsh)
	return facade, nil
}
