package limiter

import "gateway/server/interfaces"

func ProvideLimiter(facade *AlgorithmFacede, stor Storage) interfaces.Limiter {
	return &limiter{facade: facade, stor: stor}
}
