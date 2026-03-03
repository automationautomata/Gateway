package storages

import (
	"context"
	"fmt"
	"gateway/server/interfaces"
	"reflect"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/redis/go-redis/v9"
)

type redisCache[T interfaces.CacheContent] struct {
	rdb   *redis.Client
	group singleflight.Group
}

func NewRedisCache[T interfaces.CacheContent](rdb *redis.Client) *redisCache[T] {
	return &redisCache[T]{rdb, singleflight.Group{}}
}

func (c *redisCache[T]) Get(ctx context.Context, key string) (T, error) {
	groupKey := fmt.Sprint("get:", key)

	res, err, _ := c.group.Do(groupKey, func() (any, error) {
		rawVal, err := c.rdb.Get(ctx, key).Bytes()
		if err == redis.Nil {
			return nil, interfaces.ErrCacheNotFound
		}
		if err != nil {
			return nil, fmt.Errorf("redis GET failed for key %q: %w", key, err)
		}

		tType := reflect.TypeOf((*T)(nil)).Elem()
		structType := tType.Elem()
		val := reflect.New(structType).Interface().(T)

		if err = val.UnmarshalJSON(rawVal); err != nil {
			return nil, err
		}
		return val, nil
	})

	if err != nil {
		var empty T
		return empty, err
	}
	return res.(T), nil
}

func (c *redisCache[T]) Set(ctx context.Context, key string, value T, ttl time.Duration) error {
	groupKey := fmt.Sprint("set:", key)

	_, err, _ := c.group.Do(groupKey, func() (any, error) {
		jsonData, err := value.MarshalJSON()
		if err != nil {
			return nil, fmt.Errorf("cannot marshal value for key %q: %w", key, err)
		}

		err = c.rdb.Set(ctx, key, jsonData, ttl).Err()
		if err != nil {
			return nil, fmt.Errorf("redis SET failed for key %q: %w", key, err)
		}
		return nil, nil
	})
	return err
}
