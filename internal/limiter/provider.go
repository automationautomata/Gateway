package limiter

import "gateway/server/interfaces"

func ProvideLimiter(fact *AlgorithmFactory, stor Storage) interfaces.Limiter {
	return &limiter{fact: fact, stor: stor}
}
