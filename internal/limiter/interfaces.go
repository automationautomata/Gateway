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
	Save(key, algorithmName string, state *State) error
	Get(key, algorithmName string) (*State, error)
}
