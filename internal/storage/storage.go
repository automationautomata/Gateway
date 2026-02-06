package storage

import (
	"context"
	"fmt"
	lim "gateway/internal/limiter"
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

// func (s *redisStorage) Save(ctx context.Context, key, algorithmName string, state *lim.State) error {
// 	if state == nil {
// 		return errors.New("state is nil")
// 	}

// 	data, err := json.Marshal(state)
// 	if err != nil {
// 		return err
// 	}

// 	status := s.rdb.Set(ctx, s.redisKey(key), data, s.keyTTL)
// 	return status.Err()
// }

func (s *redisStorage) Update(ctx context.Context, input lim.UpdateInput, update lim.UpdateFunc) error {
	key := s.redisKey(input.Key, input.Algorithm)

	state, err := s.get(ctx, key, input.Unmarsh)
	if err != nil {
		return err
	}

	newState, err := update(state)
	if err != nil {
		return err
	}

	if err := s.set(ctx, key, newState); err != nil {
		return err
	}
	return nil
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

func (s *redisStorage) set(ctx context.Context, key string, state *lim.State) error {
	data, err := state.Params.Marshal()
	if err != nil {
		return err
	}

	status := s.rdb.Set(ctx, key, data, s.keyTTL)
	return status.Err()
}
