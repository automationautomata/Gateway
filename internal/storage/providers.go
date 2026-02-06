package storage

import (
	"gateway/internal/limiter"
	"time"

	"github.com/redis/go-redis/v9"
)

func ProvideStorage(rdb *redis.Client, keyTTL time.Duration) limiter.Storage {
	return newRedisStorage(rdb, keyTTL)
}
