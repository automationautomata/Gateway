package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	limiterLabels = []string{"allowed", "dest"}
	proxyLabels   = []string{"dest"}
)

type limiterMetric struct {
	counter *prometheus.CounterVec
}

func newLimiterMetric(name string) *limiterMetric {
	return &limiterMetric{
		counter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: name,
			},
			limiterLabels,
		),
	}
}

func (m *limiterMetric) Record(allow bool, dest string) {
	m.counter.WithLabelValues("allow", dest).Inc()
}

type proxyMetric struct {
	counter *prometheus.CounterVec
}

func newProxyMetric(name string) *proxyMetric {
	return &proxyMetric{
		counter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: name,
			},
			proxyLabels,
		),
	}
}

func (m *proxyMetric) Record(dest string) {
	m.counter.WithLabelValues(dest).Inc()
}
