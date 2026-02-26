package interfaces

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type LimiterMetric interface {
	Inc(allowed bool, dest string)
}

type ProxyMetric interface {
	Inc(dest string)
}

type CacheMetric interface {
	Inc(host, path, query string, hit bool)
}

type Limiter interface {
	Allow(ctx context.Context, key string) (bool, error)
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

type CacheContent interface {
	json.Marshaler
	json.Unmarshaler
}

type LoadMissedFunc[T CacheContent] func(context.Context) (T, time.Duration, error)

type CacheStorage[T CacheContent] interface {
	Get(ctx context.Context, key string, loader LoadMissedFunc[T]) (T, error)
}
