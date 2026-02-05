package storage

import (
	"fmt"
	"gateway/config"
	"gateway/internal/limiter"

	"github.com/redis/go-redis/v9"
)

func createRedisClient(url string) (*redis.Client, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	return redis.NewClient(opt), nil
}

func ProvideStorage(cfg config.StorageSettings) (limiter.Storage, error) {
	redis, err := createRedisClient(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("connot create redis client: %w", err)
	}
	return newRedisStorage(redis, *cfg.TTL), nil
}
