package bootstrap

import (
	"context"
	"fmt"
	"gateway/config"
	"gateway/server"
	"gateway/server/interfaces"

	"gateway/server/handlers"
	mw "gateway/server/middlewares"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	metricsPath = "/metrics"
	healthPath  = "/health"

	defaultIsGlobalLimiter = false
	defaultKeyTTL          = 0
	defaultLogLevel        = config.LevelError
)

type Shutdown func(context.Context)

func Run(fileConf config.FileConfig, envConf config.EnvConfig) Shutdown {
	setConfigDeafultValues(&fileConf, &envConf)

	err := checkProxyRoutes(fileConf.Proxy.Router, metricsPath, healthPath)
	if err != nil {
		panic(err)
	}

	rootLogger := provideRootLogger(*envConf.LogLevel)
	gateway, err := provideGateway(fileConf, envConf, rootLogger)
	if err != nil {
		panic(fmt.Errorf("cannot create gateway: %w", err))
	}

	recoverMw := mw.NewRecover(rootLogger)
	whitelistMw := mw.NewWhitelist(fileConf.Metrics.Hosts...)
	metricHandler := whitelistMw.Wrap(promhttp.Handler())

	opts := server.ServerOptions{
		Gateway:     gateway,
		Middlewares: []interfaces.Middleware{recoverMw},
		Handlers: map[string]http.Handler{
			healthPath:  handlers.Health(),
			metricsPath: metricHandler,
		},
	}
	srv := server.NewServer(envConf.ServerConfig, opts)

	go func() {
		fmt.Printf("running on %s\n", srv.Addr)

		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			fmt.Printf("server failed: %v", err)
		}
	}()

	return func(ctx context.Context) {
		if err := srv.Shutdown(ctx); err != nil {
			fmt.Printf("shutdown error: %v", err)
		}
	}
}

func setConfigDeafultValues(fileConf *config.FileConfig, envConf *config.EnvConfig) {
	if fileConf.EdgeLimiter.IsGlobal == nil {
		v := defaultIsGlobalLimiter
		fileConf.EdgeLimiter.IsGlobal = &v
	}

	keyTTL := &config.StorageSettings{KeyTTL: defaultKeyTTL}

	if fileConf.EdgeLimiter.Limiter.Storage == nil {
		fileConf.EdgeLimiter.Limiter.Storage = keyTTL
	}

	if fileConf.Proxy.Limiter != nil && fileConf.Proxy.Limiter.Storage == nil {
		fileConf.Proxy.Limiter.Storage = keyTTL
	}

	if envConf.LogLevel == nil {
		v := defaultLogLevel
		envConf.LogLevel = &v
	}
}
