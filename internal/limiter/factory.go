package limiter

type AlgorithmFactory struct {
	name       string
	alg        Algorithm
	firstState *State
}

func NewFactory(name string, alg Algorithm, firstState *State) *AlgorithmFactory {
	return &AlgorithmFactory{
		name:       name,
		alg:        alg,
		firstState: firstState,
	}
}
