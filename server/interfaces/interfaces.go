package interfaces

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
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

type CacheMetric interface {
	Inc(host, path, query string, hit bool)
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

type LoadFunc[T json.Marshaler] func() (T, time.Duration, error)

type Cache[T json.Marshaler] interface {
	Get(ctx context.Context, key string, loader LoadFunc[T]) (T, error)
}
