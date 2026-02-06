package limiter

import "context"

type Marshaler interface {
	Marshal() ([]byte, error)
}

type Unmarshaler[T any] interface {
	Unmarshal(data []byte) (*T, error)
}

type State struct {
	Params Marshaler
}

type Algorithm interface {
	Action(ctx context.Context, state *State) (bool, *State, error)
}

type UpdateInput struct {
	Key       string
	Algorithm string
	Unmarsh   Unmarshaler[State]
}

type Storage interface {
	Update(ctx context.Context, input UpdateInput, update UpdateFunc) error
}

type UpdateFunc func(*State) (new *State, err error)
