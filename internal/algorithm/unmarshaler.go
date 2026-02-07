package algorithm

import (
	"encoding/json"
	"gateway/internal/limiter"
)

type stateUnmarshaler[T limiter.Marshaler] struct{}

func NewStateUnmarshaler[T limiter.Marshaler]() *stateUnmarshaler[T] {
	return &stateUnmarshaler[T]{}
}

func (*stateUnmarshaler[T]) Unmarshal(data []byte) (*limiter.State, error) {
	var p T
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &limiter.State{Params: p}, nil
}
