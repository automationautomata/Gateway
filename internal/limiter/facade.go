package limiter

type AlgorithmFacede struct {
	Algorithm
	unmarsh    Unmarshaler[State]
	name       string
	firstState *State
}

func NewFacade(
	name string, alg Algorithm, firstState *State, unmarsh Unmarshaler[State],
) *AlgorithmFacede {
	return &AlgorithmFacede{
		Algorithm:  alg,
		name:       name,
		unmarsh:    unmarsh,
		firstState: firstState,
	}
}
