package bootstrap

import (
	"gateway/config"
	"gateway/internal/algorithm"
	"gateway/internal/limiter"
	"gateway/internal/metrics"
	"gateway/internal/storage"
	"gateway/server/handlers"
	"gateway/server/interfaces"
	mw "gateway/server/middlewares"
)

const (
	proxyMetricName   = "proxy"
	limiterMetricName = "limiter"
)

func provideProxyHandler(cfg config.ReverseProxyConfig) (*handlers.HttpReverseProxy, error) {
	proxyMetric := metrics.ProvideProxyMetric(proxyMetricName)

	if cfg.LimiterConfig == nil {
		return handlers.NewHttpReverseProxy(cfg.Rules, proxyMetric)
	}

	lim, err := provideLimiter(*cfg.LimiterConfig)
	if err != nil {
		return nil, err
	}

	limMetric := metrics.ProvideLimiterMetric(limiterMetricName)
	return handlers.NewHttpReverseProxy(cfg.Rules, proxyMetric, handlers.WithLimiter(lim, limMetric))
}

func provideRateLimitMiddleware(cfg config.EdgeLimiterConfig) (*mw.RateLimiter, error) {
	lim, err := provideLimiter(cfg.Limiter)
	if err != nil {
		return nil, err
	}

	keyType := mw.Global
	if !(*cfg.IsGlobalLimiter) {
		keyType = mw.IP
	}
	return mw.NewRateLimiter(lim, keyType), nil
}

func provideLimiter(cfg config.LimiterSettings) (interfaces.Limiter, error) {
	fact, err := algorithm.ProvideAlgorithmFacade(cfg.AlgorithmSettings)
	if err != nil {
		return nil, err
	}

	stor, err := storage.ProvideStorage(cfg.Storage)
	if err != nil {
		return nil, err
	}

	return limiter.ProvideLimiter(fact, stor), nil
}
