package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	limiterLabels = []string{"allowed", "dest"}
	proxyLabels   = []string{"dest"}
)

type metric struct {
	valuesChan chan []string
	counter    *prometheus.CounterVec
}

func newMetric(name string, labels []string) *metric {
	m := &metric{
		counter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: name,
			},
			labels,
		),
		valuesChan: make(chan []string),
	}
	return m
}

func (m *metric) StartCount() {
	go func() {
		for val := range m.valuesChan {
			m.counter.WithLabelValues(val...).Inc()
		}
	}()
}

type limiterMetric struct {
	*metric
}

func NewLimiterMetric(name string) *limiterMetric {
	return &limiterMetric{
		metric: newMetric(name, limiterLabels),
	}
}

func (m *limiterMetric) Inc(allow bool, dest string) {
	m.metric.valuesChan <- []string{strconv.FormatBool(allow), dest}
}

type proxyMetric struct {
	*metric
}

func NewProxyMetric(name string) *proxyMetric {
	return &proxyMetric{
		metric: newMetric(name, proxyLabels),
	}
}

func (m *proxyMetric) Inc(dest string) {
	m.metric.valuesChan <- []string{dest}
}
