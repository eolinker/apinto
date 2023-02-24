package prometheus

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"strings"
)

type iMetric interface {
	Observe(value float64, labels map[string]string)
	Register(registry *prometheus.Registry) error
	UnRegister(registry *prometheus.Registry)
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

func (c *counterVec) UnRegister(registry *prometheus.Registry) {
	registry.Unregister(c.counter)
}

func newCounterVec(name, description string, labels []string) (iMetric, error) {
	return &counterVec{
		counter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: name,
			Help: description,
		}, labels),
	}, nil
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

func (g *gaugeVec) UnRegister(registry *prometheus.Registry) {
	registry.Unregister(g.gauge)
}

func newGaugeVec(name, description string, labels []string) (iMetric, error) {
	return &gaugeVec{
		gauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: name,
			Help: description,
		}, labels),
	}, nil
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

func (h *histogramVec) UnRegister(registry *prometheus.Registry) {
	registry.Unregister(h.histogram)
}

func newHistogramVec(name, description string, labels []string) (iMetric, error) {
	return &histogramVec{
		histogram: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: name,
			Help: description,
		}, labels),
	}, nil
}

type summaryVec struct {
	summary *prometheus.SummaryVec
}

var (
	defaultObjectives = "0.5:0.05,0.9:0.01,0.99:0.001"
)

func (s *summaryVec) Observe(value float64, labels map[string]string) {
	s.summary.With(labels).Observe(value)
}

func (s *summaryVec) Register(registry *prometheus.Registry) error {
	return registry.Register(s.summary)
}

func (s *summaryVec) UnRegister(registry *prometheus.Registry) {
	registry.Unregister(s.summary)
}

func newSummaryVec(name, description string, labels []string, objectives string) (iMetric, error) {
	objectivesList := strings.Split(objectives, ",")

	objectivesCfg := make(map[float64]float64, len(objectivesList))
	for _, obj := range objectivesList {
		if obj == "" {
			continue
		}
		idx := strings.Index(obj, ":")
		quantile, err := strconv.ParseFloat(obj[:idx], 64)
		if err != nil {
			return nil, err
		}
		estimate, err := strconv.ParseFloat(obj[idx+1:], 64)
		if err != nil {
			return nil, err
		}
		objectivesCfg[quantile] = estimate
	}

	return &summaryVec{
		summary: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Name:       name,
			Help:       description,
			Objectives: objectivesCfg,
		}, labels),
	}, nil
}

func newIMetric(metricInfo *metricInfoCfg, name, description string, objectives string) (iMetric, error) {
	labels := make([]string, 0, len(metricInfo.labels))
	for _, l := range metricInfo.labels {
		labels = append(labels, l.Name)
	}

	switch collectorTypeSet[metricInfo.collector] {
	case typeCounter:
		return newCounterVec(name, description, labels)
	case typeGauge:
		return newGaugeVec(name, description, labels)
	case typeHistogram:
		return newHistogramVec(name, description, labels)
	case typeSummary:
		return newSummaryVec(name, description, labels, objectives)
	default:
		return nil, fmt.Errorf(errorMetricTypeFormat, collectorTypeSet[metricInfo.collector])
	}
}
