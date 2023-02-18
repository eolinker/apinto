package prometheus

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
)

type iMetric interface {
	Observe(value float64, labels map[string]string)
	Register(registry *prometheus.Registry) error
}

type counterVec struct {
	counter *prometheus.CounterVec
}

func (c *counterVec) Observe(value float64, labels map[string]string) {
	//counter的value必须大于0
	c.counter.With(labels).Add(value)
}

func (c *counterVec) Register(registry *prometheus.Registry) error {
	return registry.Register(c.counter)
}

func newCounterVec(name, description string, labels []string) iMetric {
	return &counterVec{
		counter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: name,
			Help: description,
		}, labels),
	}
}

type gaugeVec struct {
	gauge *prometheus.GaugeVec
}

func (g *gaugeVec) Observe(value float64, labels map[string]string) {
	g.gauge.With(labels).Add(value)
}

func (g *gaugeVec) Register(registry *prometheus.Registry) error {
	return registry.Register(g.gauge)
}

func newGaugeVec(name, description string, labels []string) iMetric {
	return &gaugeVec{
		gauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: name,
			Help: description,
		}, labels),
	}
}

type histogramVec struct {
	histogram *prometheus.HistogramVec
}

func (h *histogramVec) Observe(value float64, labels map[string]string) {
	h.histogram.With(labels).Observe(value)
}

func (h *histogramVec) Register(registry *prometheus.Registry) error {
	return registry.Register(h.histogram)
}

func newHistogramVec(name, description string, labels []string) iMetric {
	return &histogramVec{
		histogram: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: name,
			Help: description,
		}, labels),
	}
}

type summaryVec struct {
	summary *prometheus.SummaryVec
}

func (s *summaryVec) Observe(value float64, labels map[string]string) {
	s.summary.With(labels).Observe(value)
}

func (s *summaryVec) Register(registry *prometheus.Registry) error {
	return registry.Register(s.summary)
}

func newSummaryVec(name, description string, labels []string) iMetric {
	return &summaryVec{
		summary: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Name: name,
			Help: description,
		}, labels),
	}
}

func newIMetric(metricType, name, description string, labels []string) (iMetric, error) {
	switch metricType {
	case typeCounter:
		return newCounterVec(name, description, labels), nil
	case typeGauge:
		return newGaugeVec(name, description, labels), nil
	case typeHistogram:
		return newHistogramVec(name, description, labels), nil
	case typeSummary:
		return newSummaryVec(name, description, labels), nil
	default:
		return nil, fmt.Errorf(errorMetricTypeFormat, metricType)
	}
}
