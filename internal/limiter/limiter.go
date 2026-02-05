package limiter

import (
	"context"
	"fmt"
	"gateway/server/interfaces"
)

type limiter struct {
	fact *AlgorithmFactory
	stor Storage
}

func NewLimiter(fact *AlgorithmFactory, stor Storage) interfaces.Limiter {
	return &limiter{fact: fact, stor: stor}
}

func (l *limiter) Allow(ctx context.Context, key string) (bool, error) {
	state, err := l.stor.Get(key, l.fact.name)
	if err == ErrIvalidState {
		state = l.fact.firstState
	} else if err != nil {
		return false, fmt.Errorf("cannot get state by key: %w", err)
	}

	alg := l.fact.alg

	newState, err := alg.Action(ctx, state)
	if err != nil {
		return false, fmt.Errorf("cannot do action: %w", err)
	}

	if err = l.stor.Save(key, l.fact.name, newState); err != nil {
		return false, fmt.Errorf("cannot save state: %w", err)
	}
	return newState.Allow, nil
}
