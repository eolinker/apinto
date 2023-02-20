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
	if value < 0 {
		return
	}
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

var (
	defaultObjectives = map[float64]float64{
		0.5:  0.05,
		0.9:  0.01,
		0.99: 0.001,
	}
)

func (s *summaryVec) Observe(value float64, labels map[string]string) {
	s.summary.With(labels).Observe(value)
}

func (s *summaryVec) Register(registry *prometheus.Registry) error {
	return registry.Register(s.summary)
}

func newSummaryVec(name, description string, labels []string, objectives map[float64]float64) iMetric {
	if objectives == nil || len(objectives) == 0 {
		objectives = defaultObjectives
	}
	return &summaryVec{
		summary: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Name:       name,
			Help:       description,
			Objectives: objectives,
		}, labels),
	}
}

func newIMetric(collector, name, description string, metricInfo *metricInfoCfg, objectives map[float64]float64) (iMetric, error) {
	labels := make([]string, 0, len(metricInfo.labels))
	for _, l := range metricInfo.labels {
		labels = append(labels, l.Name)
	}

	switch collectorTypeSet[collector] {
	case typeCounter:
		return newCounterVec(name, description, labels), nil
	case typeGauge:
		return newGaugeVec(name, description, labels), nil
	case typeHistogram:
		return newHistogramVec(name, description, labels), nil
	case typeSummary:
		return newSummaryVec(name, description, labels, objectives), nil
	default:
		return nil, fmt.Errorf(errorMetricTypeFormat, collectorTypeSet[collector])
	}
}
