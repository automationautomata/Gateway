package bootstrap

import (
	"gateway/config"
	"gateway/internal/algorithm"
	"gateway/internal/algorithm/fixedwindow"
	"gateway/internal/algorithm/slidingwindow"
	"gateway/internal/algorithm/tokenbucket"
	"gateway/internal/limiter"
	"gateway/internal/metrics"
	"gateway/internal/storage"
	"gateway/server/interfaces"
	mw "gateway/server/middlewares"
	"gateway/server/proxy"

	"github.com/redis/go-redis/v9"
)

const (
	proxyMetricName        = "proxy"
	proxyLimiterMetricName = "limiter"
	edgeLimiterMetricName  = "rate_limit_iddlewar"
)

func provideProxyHandler(cfg config.ReverseProxyConfig, rdb *redis.Client) (*proxy.HttpReverseProxy, error) {
	proxyMetric := metrics.NewProxyMetric(proxyMetricName)
	proxyMetric.StartCount()
	if cfg.LimiterConfig == nil {
		return proxy.NewHttpReverseProxy(cfg.Rules, proxyMetric)
	}

	lim, err := provideLimiter(*cfg.LimiterConfig, rdb)
	if err != nil {
		return nil, err
	}

	limMetric := metrics.NewLimiterMetric(proxyLimiterMetricName)
	limMetric.StartCount()
	return proxy.NewHttpReverseProxy(
		cfg.Rules, proxyMetric, proxy.WithLimiter(lim, limMetric),
	)
}

func provideRateLimitMiddleware(cfg config.EdgeLimiterConfig, rdb *redis.Client) (*mw.RateLimiter, error) {
	lim, err := provideLimiter(cfg.Limiter, rdb)
	if err != nil {
		return nil, err
	}

	keyType := mw.Global
	if !(*cfg.IsGlobalLimiter) {
		keyType = mw.IP
	}

	limMetric := metrics.NewLimiterMetric(edgeLimiterMetricName)
	limMetric.StartCount()
	return mw.NewRateLimiter(lim, mw.WithKeyType(keyType), mw.WithMetric(limMetric)), nil
}

func provideLimiter(cfg config.LimiterSettings, rdb *redis.Client) (interfaces.Limiter, error) {
	fact, err := provideAlgorithmFacade(cfg.Type, cfg.Algorithm)
	if err != nil {
		return nil, err
	}

	stor := storage.NewRedisStorage(rdb, cfg.Storage.KeyTTL)
	return limiter.ProvideLimiter(fact, stor), nil
}

func provideAlgorithmFacade(algType config.AlgorithmType, settings any) (*limiter.AlgorithmFacade, error) {
	var (
		alg     limiter.Algorithm
		unmarsh limiter.Unmarshaler[limiter.State]
	)

	switch algType {
	case config.TokenBucketAlgorithm:
		algConf := settings.(*config.TokenBucketSettings)
		alg = tokenbucket.NewTokenBucket(algConf.Capacity, algConf.Rate)
		unmarsh = algorithm.NewStateUnmarshaler[*tokenbucket.Params]()

	case config.FixedWindowAlgorithm:
		algConf := settings.(*config.FixedWindowSettings)
		alg = fixedwindow.NewFixedWindow(algConf.Limit, algConf.WindowDuration)
		unmarsh = algorithm.NewStateUnmarshaler[*fixedwindow.Params]()

	case config.SlidingWindowLogAlgorithm:
		algConf := settings.(*config.SlidingWindowLogSettings)
		alg = slidingwindow.NewSlidingWindowCounter(
			algConf.WindowDuration, algConf.Limit, int64(algConf.Limit),
		)
		unmarsh = algorithm.NewStateUnmarshaler[*slidingwindow.LogParams]()

	case config.SlidingWindowCounterAlgorithm:
		algConf := settings.(*config.SlidingWindowCounterSettings)
		alg = slidingwindow.NewSlidingWindowCounter(
			algConf.WindowDuration, algConf.BucketsNum, algConf.Limit,
		)
		unmarsh = algorithm.NewStateUnmarshaler[*slidingwindow.CounterParams]()

	}

	facade := limiter.NewFacade(string(algType), alg, unmarsh)
	return facade, nil
}
func provideRedisClient(url string) (*redis.Client, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	return redis.NewClient(opt), nil
}
