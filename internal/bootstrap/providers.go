package bootstrap

import (
	"gateway/config"
	"gateway/internal/algorithm"
	"gateway/internal/limiter"
	"gateway/internal/metrics"
	"gateway/internal/storage"
	"gateway/server/handlers"
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

	fact, err := algorithm.ProvideAlgorithmFactory(cfg.LimiterConfig.AlgorithmSettings)
	if err != nil {
		return nil, err
	}

	stor, err := storage.ProvideStorage(cfg.LimiterConfig.Storage)
	if err != nil {
		return nil, err
	}

	lim := limiter.ProvideLimiter(fact, stor)
	limMetric := metrics.ProvideLimiterMetric(limiterMetricName)
	return handlers.NewHttpReverseProxy(cfg.Rules, proxyMetric, handlers.WithLimiter(lim, limMetric))
}

func provideRateLimitMiddleware(cfg config.EdgeLimiterConfig) (*mw.RateLimiter, error) {
	fact, err := algorithm.ProvideAlgorithmFactory(cfg.Limiter.AlgorithmSettings)
	if err != nil {
		return nil, err
	}

	stor, err := storage.ProvideStorage(cfg.Limiter.Storage)
	if err != nil {
		return nil, err
	}

	lim := limiter.ProvideLimiter(fact, stor)

	keyType := mw.Global
	if !(*cfg.IsGlobalLimiter) {
		keyType = mw.IP
	}
	return mw.NewRateLimiter(lim, keyType), nil
}
