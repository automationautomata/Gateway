package limiter

type AlgorithmFacade struct {
	Algorithm
	unmarsh    Unmarshaler[State]
	name       string
	firstState *State
}

func NewFacade(
	name string, alg Algorithm, firstState *State, unmarsh Unmarshaler[State],
) *AlgorithmFacade {
	return &AlgorithmFacade{
		Algorithm:  alg,
		name:       name,
		unmarsh:    unmarsh,
		firstState: firstState,
	}
}
