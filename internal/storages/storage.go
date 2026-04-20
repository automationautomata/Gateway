package storages

import (
	"context"
	"fmt"
	lim "gateway/internal/limiter"
	"gateway/pkg/keymutex"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisStorage struct {
	rdb    *redis.Client
	keyTTL time.Duration
	mu     *keymutex.KeyMutex[string]
}

func NewRedisStorage(rdb *redis.Client, keyTTL time.Duration) *redisStorage {
	return &redisStorage{
		rdb:    rdb,
		keyTTL: keyTTL,
		mu:     keymutex.New[string](),
	}
}

func (s *redisStorage) Update(ctx context.Context, input lim.UpdateInput, update lim.UpdateFunc) error {
	key := s.redisKey(input.Key, input.Algorithm)

	s.mu.Lock(key)
	defer s.mu.Unlock(key)

	state, err := s.get(ctx, key, input.Unmarsh)
	if err != nil {
		return err
	}

	newState, err := update(state)
	if err != nil {
		return err
	}

	data, err := newState.Params.Marshal()
	if err != nil {
		return err
	}

	err = s.rdb.Set(ctx, key, data, s.keyTTL).Err()
	return err
}

func (s *redisStorage) redisKey(key, algorithm string) string {
	return fmt.Sprintf("state:%s:%s", key, algorithm)
}

func (s *redisStorage) get(ctx context.Context, key string, unmarsh lim.Unmarshaler[lim.State]) (*lim.State, error) {
	val, err := s.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return unmarsh.Unmarshal(val)
}
