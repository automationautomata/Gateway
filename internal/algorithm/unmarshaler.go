package algorithm

import (
	"encoding/json"
	"gateway/internal/limiter"
)

type unmarshaler[T limiter.Marshaler] struct{}

func (*unmarshaler[T]) Unmarshal(b []byte) (*limiter.State, error) {
	var p T
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	return &limiter.State{Params: p}, nil
}
