package interfaces

import (
	"context"
	"net/http"
)

type Limiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}

type LimiterMetric interface {
	Inc(allowed bool, dest string)
}

type ProxyMetric interface {
	Inc(dest string)
}

type Middleware interface {
	Wrap(next http.Handler) http.Handler
}

type Logger interface {
	Debug(ctx context.Context, msg string, fields map[string]any)
	Info(ctx context.Context, msg string, fields map[string]any)
	Warn(ctx context.Context, msg string, fields map[string]any)
	Error(ctx context.Context, msg string, fields map[string]any)
}
