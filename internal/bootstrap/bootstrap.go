package bootstrap

import (
	"context"
	"fmt"
	"gateway/config"
	"gateway/internal/common"
	"gateway/server"

	"gateway/server/handlers"
	mw "gateway/server/middlewares"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	isGlobalLimiterDeafult = false
	metricEndpoint         = "/metric"
	healthEndpoint         = "/health"
)

var reservedEndpoints = []string{metricEndpoint, healthEndpoint}

type Shutdown func(context.Context)

func Run(cfg config.Config) Shutdown {
	if err := checkProxyRules(cfg.Proxy.Rules); err != nil {
		log.Fatal(err)
	}

	proxy, err := provideProxyHandler(cfg.Proxy)
	if err != nil {
		log.Fatalf("cannot bootstrap reverse proxy: %v", err)
	}

	if cfg.EdgeLimiter.IsGlobalLimiter == nil {
		*cfg.EdgeLimiter.IsGlobalLimiter = isGlobalLimiterDeafult
	}

	limiter, err := provideRateLimitMiddleware(cfg.EdgeLimiter)
	if err != nil {
		log.Fatalf("cannot bootstrap rate limit middleware: %v", err)
	}

	whitelistMw := mw.NewWhitelist(cfg.Metrics.Hosts...)

	mux := http.NewServeMux()
	mux.Handle("", proxy)
	mux.Handle(healthEndpoint, handlers.Health())
	mux.Handle(metricEndpoint, whitelistMw.Wrap(promhttp.Handler()))

	srv := server.NewServer(cfg.Server, mux, limiter)

	go func() {
		log.Printf("running on %s\n", srv.Addr)

		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	return func(ctx context.Context) {
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}
}

func checkProxyRules(cfg config.ReverseProxyRules) error {
	reserved := common.NewSet(reservedEndpoints...)

	for _, hostCfg := range cfg.Hosts {
		for endpoint := range hostCfg.Pathes {
			if reserved.Has(endpoint) {
				return fmt.Errorf(
					"proxy rule contain reserved endpoint: %s%s",
					hostCfg.Host, endpoint,
				)
			}
		}
	}

	return nil
}
