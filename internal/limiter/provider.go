package limiter

import "gateway/server/interfaces"

func ProvideLimiter(facade *AlgorithmFacade, stor Storage) interfaces.Limiter {
	return &limiter{facade: facade, stor: stor}
}
