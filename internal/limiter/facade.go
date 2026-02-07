package limiter

type AlgorithmFacade struct {
	Algorithm
	unmarsh Unmarshaler[State]
	name    string
}

func NewFacade(
	name string, alg Algorithm, unmarsh Unmarshaler[State],
) *AlgorithmFacade {
	return &AlgorithmFacade{
		Algorithm: alg,
		name:      name,
		unmarsh:   unmarsh,
	}
}
