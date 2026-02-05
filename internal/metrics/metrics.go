package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	limiterLabels = []string{"allowed", "dest"}
	proxyLabels   = []string{"dest"}
)

type LimiterMetric struct {
	counter *prometheus.CounterVec
}

func NewLimiterMetric(name string) *LimiterMetric {
	return &LimiterMetric{
		counter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: name,
			},
			limiterLabels,
		),
	}
}

func (m *LimiterMetric) Record(allow bool, dest string) {
	m.counter.WithLabelValues("allow", dest).Inc()
}

type ProxyMetric struct {
	counter *prometheus.CounterVec
}

func NewProxyMetric(name string) *LimiterMetric {
	return &LimiterMetric{
		counter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: name,
			},
			proxyLabels,
		),
	}
}

func (m *ProxyMetric) Record(dest string) {
	m.counter.WithLabelValues(dest).Inc()
}
