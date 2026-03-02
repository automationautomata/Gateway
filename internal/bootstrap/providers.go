package bootstrap

import (
	"fmt"
	"gateway/config"
	"gateway/internal/algorithm"
	"gateway/internal/algorithm/fixedwindow"
	"gateway/internal/algorithm/slidingwindow"
	"gateway/internal/algorithm/tokenbucket"
	"gateway/internal/limiter"
	"gateway/internal/logging"
	"gateway/internal/metrics"
	"gateway/internal/storages"
	"gateway/server"
	"gateway/server/cache"
	"gateway/server/interfaces"
	"log/slog"
	"os"

	"github.com/redis/go-redis/v9"
)

const (
	proxyMetricName           = "proxy"
	httpCacheMetricName       = "http_cache"
	edgeLimiterMetricName     = "edge_limiter"
	internalLimiterMetricName = "internal_limiter"

	gatewayLoggerName         = "gateway"
	cacheLoggerName           = "http_cache"
	edgeLimiterLoggerName     = "edge_limiter"
	internalLimiterLoggerName = "internal_limiter"

	redisEdgeLimiterDB     = "/0"
	redisInternalLimiterDB = "/1"
	redisCacheDB           = "/2"
)

func provideGateway(fileConf config.FileConfig, envConf config.EnvConfig, rootLogger *logging.SlogAdapter) (*server.Gateway, error) {
	redisURL := fmt.Sprint(envConf.RedisURL, redisEdgeLimiterDB)
	edgeLimiterRedis, err := provideRedisClient(redisURL)
	if err != nil {
		return nil, fmt.Errorf("cannot create redis client %s: %w", redisURL, err)
	}

	egdeLim, err := provideLimiter(fileConf.EdgeLimiter.Limiter, edgeLimiterRedis)
	if err != nil {
		return nil, fmt.Errorf("cannot create edge limiter %w", err)
	}

	redisURL = fmt.Sprint(envConf.RedisURL, redisCacheDB)
	cacheRedis, err := provideRedisClient(redisURL)
	if err != nil {
		return nil, fmt.Errorf("cannot create redis client %s: %w", redisURL, err)
	}

	isGlobal := *fileConf.EdgeLimiter.IsGlobal
	proxyConfig := fileConf.Proxy

	var defProxy *config.UpstreamSettings
	if proxyConfig.Router.Default != nil {
		defProxy = proxyConfig.Router.Default.UpstreamSettings
	}

	proxyMetric, err := provideProxyMetric()
	if err != nil {
		return nil, fmt.Errorf("cannot create proxy metric: %w", err)
	}

	edgeLimMetric, err := provideEdgeLimiterMetric()
	if err != nil {
		return nil, fmt.Errorf("cannot edge limiter metric: %w", err)
	}

	cacheMetric, err := provideCacheMetric()
	if err != nil {
		return nil, fmt.Errorf("cannot cache storage metric: %w", err)
	}

	routerOpts := server.RouterOptions{
		Settings: fileConf.Proxy.Router,
		Proxy: server.ProxyOptions{
			Metric:  proxyMetric,
			Default: defProxy,
		},
		Cache: &server.CacheOptions{
			Metric: cacheMetric,
			Log:    rootLogger.Component(cacheLoggerName),
			Store:  provideCacheStorage[*cache.ResponseContent](cacheRedis),
		},
	}
	limOpts := server.LimiterOptions{
		Log:     rootLogger.Component(edgeLimiterLoggerName),
		Metric:  edgeLimMetric,
		Limiter: egdeLim,
	}

	builder := server.NewGatewayBuilder().
		Router(routerOpts).
		EdgeLimiter(limOpts, isGlobal).
		Logger(rootLogger.Component(gatewayLoggerName))

	if proxyConfig.Limiter != nil {
		redisURL = fmt.Sprint(envConf.RedisURL, redisInternalLimiterDB)
		internalLimiterRedis, err := provideRedisClient(redisURL)
		if err != nil {
			return nil, fmt.Errorf("cannot create redis client %s: %w", redisURL, err)
		}

		internalLim, err := provideLimiter(fileConf.EdgeLimiter.Limiter, internalLimiterRedis)
		if err != nil {
			return nil, fmt.Errorf("cannot create internal limiter %w", err)
		}

		internalLimMetric, err := provideInternalLimiterMetric()
		if err != nil {
			return nil, fmt.Errorf("cannot create internal limiter metric: %w", err)
		}
		builder = builder.InternalLimiter(
			server.LimiterOptions{
				Log:     rootLogger.Component(internalLimiterLoggerName),
				Metric:  internalLimMetric,
				Limiter: internalLim,
			},
		)
	}

	return builder.Build()
}

func provideProxyMetric() (interfaces.ProxyMetric, error) {
	proxyMetric := metrics.NewProxyMetric(proxyMetricName)
	if err := proxyMetric.StartCount(); err != nil {
		return nil, err
	}
	return proxyMetric, nil
}

func provideEdgeLimiterMetric() (interfaces.LimiterMetric, error) {
	limMetric := metrics.NewLimiterMetric(edgeLimiterMetricName)
	if err := limMetric.StartCount(); err != nil {
		return nil, err
	}
	return limMetric, nil
}

func provideInternalLimiterMetric() (interfaces.LimiterMetric, error) {
	limMetric := metrics.NewLimiterMetric(internalLimiterMetricName)
	if err := limMetric.StartCount(); err != nil {
		return nil, err
	}
	return limMetric, nil
}

func provideCacheMetric() (interfaces.CacheMetric, error) {
	cacheMetric := metrics.NewCacheMetric(httpCacheMetricName)
	if err := cacheMetric.StartCount(); err != nil {
		return nil, err
	}
	return cacheMetric, nil
}

func provideCacheStorage[T interfaces.CacheContent](rdb *redis.Client) interfaces.CacheStorage[T] {
	return storages.NewRedisCache[T](rdb)
}

func provideLimiter(cfg config.LimiterSettings, rdb *redis.Client) (interfaces.Limiter, error) {
	fact, err := provideAlgorithmFacade(cfg.Type, cfg.Algorithm)
	if err != nil {
		return nil, err
	}

	stor := storages.NewRedisStorage(rdb, cfg.Storage.KeyTTL)
	return limiter.NewLimiter(fact, stor), nil
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

func provideRootLogger(level config.LogLevel) *logging.SlogAdapter {
	var slogLevel slog.Level
	switch level {
	case config.LevelDebug:
		slogLevel = slog.LevelDebug
	case config.LevelInfo:
		slogLevel = slog.LevelInfo
	case config.LevelWarn:
		slogLevel = slog.LevelWarn
	case config.LevelError:
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
