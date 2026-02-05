package metrics

import "gateway/server/interfaces"

func ProvideLimiterMetric(name string) interfaces.LimiterMetric {
	return newLimiterMetric(name)
}

func ProvideProxyMetric(name string) interfaces.ProxyMetric {
	return newProxyMetric(name)
}
