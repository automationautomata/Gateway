package bootstrap

import (
	"gateway/config"
	"gateway/internal/algorithm"
	"gateway/internal/algorithm/fixedwindow"
	"gateway/internal/algorithm/slidingwindow"
	"gateway/internal/algorithm/tokenbucket"
	"gateway/internal/limiter"
	"gateway/internal/logging"
	"gateway/internal/metrics"
	"gateway/internal/storage"
	"gateway/server/interfaces"
	mw "gateway/server/middlewares"
	"gateway/server/proxy"
	"log/slog"
	"os"

	"github.com/redis/go-redis/v9"
)

const (
	proxyMetricName        = "proxy"
	proxyLimiterMetricName = "limiter"
	edgeLimiterMetricName  = "rate_limit_iddlewar"
)

func provideProxyHandler(cfg config.ReverseProxyConfig, rdb *redis.Client, log interfaces.Logger) (*proxy.HttpReverseProxy, error) {
	proxyMetric := metrics.NewProxyMetric(proxyMetricName)
	proxyMetric.StartCount()

	input := proxy.HttpProxyInput{
		Rules:       cfg.Rules,
		Log:         log,
		ProxyMetric: proxyMetric,
	}
	if cfg.LimiterConfig == nil {
		return proxy.NewHttpReverseProxy(
			input,
		)
	}

	lim, err := provideLimiter(*cfg.LimiterConfig, rdb)
	if err != nil {
		return nil, err
	}

	limMetric := metrics.NewLimiterMetric(proxyLimiterMetricName)
	limMetric.StartCount()
	return proxy.NewHttpReverseProxy(input, proxy.WithLimiter(lim, limMetric))
}

func provideRateLimitMiddleware(cfg config.EdgeLimiterConfig, rdb *redis.Client, log interfaces.Logger) (*mw.RateLimiter, error) {
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
	return mw.NewRateLimiter(lim, log, mw.WithKeyType(keyType), mw.WithMetric(limMetric)), nil
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

func provideLogger(level config.LogLevel) interfaces.Logger {
	var slogLevel slog.Level
	switch level {
	case config.Debug:
		slogLevel = slog.LevelDebug
	case config.Info:
		slogLevel = slog.LevelInfo
	case config.Warn:
		slogLevel = slog.LevelWarn
	case config.Error:
		slogLevel = slog.LevelError
	}

	return logging.NewSlogAdapter(
		slog.New(
			slog.NewJSONHandler(
				os.Stdout,
				&slog.HandlerOptions{
					Level: slogLevel,
				},
			),
		),
	)
}

func provideRedisClient(url string) (*redis.Client, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	return redis.NewClient(opt), nil
}
