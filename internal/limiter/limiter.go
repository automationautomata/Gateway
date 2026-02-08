package limiter

import (
	"context"
	"fmt"
)

type limiter struct {
	facade *AlgorithmFacade
	stor   Storage
}

func (l *limiter) Allow(ctx context.Context, key string) (bool, error) {
	input := UpdateInput{key, l.facade.name, l.facade.unmarsh}

	var allow bool
	err := l.stor.Update(
		ctx,
		input,
		func(s *State) (new *State, err error) {
			if s == nil {
				s = l.facade.FirstState()
			}
			allow, new, err = l.facade.Action(s)
			return new, err
		},
	)
	if err != nil {
		return false, fmt.Errorf("cannot update state: %w", err)
	}
	return allow, nil
}
