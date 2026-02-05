package limiter

import "context"

type State struct {
	Allow  bool
	Params map[string]any
}

type Algorithm interface {
	Action(ctx context.Context, state *State) (*State, error)
}

type Storage interface {
	Save(ctx context.Context, key, algorithmName string, state *State) error
	Get(ctx context.Context, key, algorithmName string) (*State, error)
}
