package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gateway/internal/limiter"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisStorage struct {
	rdb    *redis.Client
	keyTTL time.Duration
}

func newRedisStorage(rdb *redis.Client, keyTTL time.Duration) *redisStorage {
	return &redisStorage{
		rdb:    rdb,
		keyTTL: keyTTL,
	}
}

func (s *redisStorage) redisKey(key, algorithm string) string {
	return fmt.Sprintf("state:%s:%s", key, algorithm)
}

func (s *redisStorage) Save(ctx context.Context, key, algorithmName string, state *limiter.State) error {
	if state == nil {
		return errors.New("state is nil")
	}

	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	status := s.rdb.Set(ctx, s.redisKey(key, algorithmName), data, s.keyTTL)
	return status.Err()
}

func (s *redisStorage) Get(ctx context.Context, key, algorithmName string) (*limiter.State, error) {
	val, err := s.rdb.Get(ctx, s.redisKey(key, algorithmName)).Bytes()

	if err == redis.Nil {
		return nil, limiter.ErrStateNotFount
	}
	if err != nil {
		return nil, err
	}

	var state limiter.State
	if err := json.Unmarshal(val, &state); err != nil {
		return nil, err
	}
	return &state, nil
}
