package bootstrap

import (
	"fmt"
	"gateway/config"
	"gateway/internal/algorithm"
	"gateway/internal/common"
	"gateway/internal/limiter"
	"gateway/server/handlers"
	mw "gateway/server/middlewares"
)

func checkProxyRules(cfg config.ReverseProxyRules) error {
	reserved := common.NewSet(metricEndpoint)

	for _, hostCfg := range cfg.Hosts {
		for endpoint := range hostCfg.Pathes {
			if reserved.Has(endpoint) {
				return fmt.Errorf(
					"proxy rule contain reserved endpoint: %s%s", hostCfg.Host, endpoint,
				)
			}
		}
	}
	return nil
}

func provideProxyHandler(cfg config.ReverseProxyConfig) (*handlers.HttpReverseProxy, error) {
	if cfg.LimiterConfig == nil {
		return handlers.NewHttpReverseProxy(cfg.Rules, nil)
	}

	fact, err := algorithm.ProvideAlgorithmFactory(*cfg.LimiterConfig)
	if err != nil {
		return nil, err
	}

	lim := limiter.NewLimiter(fact, nil)
	return handlers.NewHttpReverseProxy(cfg.Rules, nil, handlers.WithLimiter(lim, nil))
}

func provideRateLimitMiddleware(cfg config.EdgeLimiterConfig) (*mw.RateLimiter, error) {
	fact, err := algorithm.ProvideAlgorithmFactory(cfg.LimiterConfig)
	if err != nil {
		return nil, err
	}
	lim := limiter.NewLimiter(fact, nil)

	keyType := mw.Global
	if !(*cfg.IsGlobalLimiter) {
		keyType = mw.IP
	}
	return mw.NewRateLimiter(lim, keyType), nil
}
