package bootstrap

import (
	"context"
	"fmt"
	"gateway/config"
	"gateway/internal/common"
	"gateway/server"
	"strings"
	"time"

	"gateway/server/handlers"
	mw "gateway/server/middlewares"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	isGlobalLimiterDeafult               = false
	metricEndpoint                       = "/metrics"
	healthEndpoint                       = "/health"
	DefaultKeyTTL          time.Duration = time.Hour
)

var reservedEndpoints = []string{metricEndpoint, healthEndpoint}

type Shutdown func(context.Context)

func Run(fileConf config.FileConfig, envConf config.EnvConfig) Shutdown {
	setConfigDeafultValues(&fileConf)
	srv, err := buildServer(fileConf, envConf)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		log.Printf("running on %s\n", srv.Addr)

		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Printf("server failed: %v", err)
		}
	}()

	return func(ctx context.Context) {
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}
}

func buildServer(fileConf config.FileConfig, envConf config.EnvConfig) (*http.Server, error) {
	edgeLimiterRedis, err := provideRedisClient(envConf.EdgeLimiterRedisURL)
	if err != nil {
		return nil, fmt.Errorf("cannot connect redis %s: %w", envConf.EdgeLimiterRedisURL, err)
	}

	proxyLimiterRedis, err := provideRedisClient(envConf.ProxyLimiterRedisURL)
	if err != nil {
		return nil, fmt.Errorf("cannot connect redis %s: %w", envConf.ProxyLimiterRedisURL, err)
	}

	if err := checkProxyRoutes(fileConf.Proxy); err != nil {
		return nil, err
	}

	logger := provideLogger(envConf.LogLevel)

	proxy, err := provideProxyHandler(fileConf.Proxy, edgeLimiterRedis, logger)
	if err != nil {
		return nil, fmt.Errorf("cannot bootstrap reverse proxy: %w", err)
	}

	limiter, err := provideEdgeLimiter(*fileConf.EdgeLimiter, proxyLimiterRedis, logger)
	if err != nil {
		return nil, fmt.Errorf("cannot bootstrap rate limit middleware: %w", err)
	}

	recoverMw := mw.NewRecover(logger)

	whitelistMw := mw.NewWhitelist(fileConf.Metrics.Hosts...)

	mux := http.NewServeMux()
	mux.Handle("/", proxy)
	mux.Handle(healthEndpoint, handlers.Health())
	mux.Handle(metricEndpoint, whitelistMw.Wrap(promhttp.Handler()))

	return server.NewServer(envConf.ServerConfig, mux, recoverMw, limiter), nil
}

func setConfigDeafultValues(cfg *config.FileConfig) {
	if cfg.EdgeLimiter.IsGlobalLimiter == nil {
		v := isGlobalLimiterDeafult
		cfg.EdgeLimiter.IsGlobalLimiter = &v
	}

	v := &config.StorageSettings{KeyTTL: 0}

	if cfg.EdgeLimiter.Limiter.Storage == nil {
		cfg.EdgeLimiter.Limiter.Storage = v
	}

	if cfg.Proxy.Limiter != nil && cfg.Proxy.Limiter.Storage == nil {
		cfg.Proxy.Limiter.Storage = v
	}
}

func checkProxyRoutes(cfg config.ReverseProxyConfig) error {
	reserved := common.NewSet(reservedEndpoints...)

	for _, route := range cfg.Routes {
		for _, path := range route.Paths {
			if reserved.Has(strings.TrimSuffix(path.Path, "/")) {
				return fmt.Errorf(
					"proxy rule contain reserved endpoint: %s%s",
					route.Host, path.Path,
				)
			}
		}
	}

	return nil
}
