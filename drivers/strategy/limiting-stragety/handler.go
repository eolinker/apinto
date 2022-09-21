package limiting_stragety

import (
	"github.com/eolinker/apinto/metrics"
	"github.com/eolinker/apinto/strategy"
)

type ThresholdUint struct {
	Second uint64
	Minute uint64
	Hour   uint64
}
type LimitingHandler struct {
	name     string
	filter   strategy.IFilter
	metrics  metrics.Metrics
	query    ThresholdUint
	traffic  ThresholdUint
	priority int
	stop     bool
}

func (l *LimitingHandler) Name() string {
	return l.name
}

func (l *LimitingHandler) Filter() strategy.IFilter {
	return l.filter
}

func (l *LimitingHandler) Metrics() metrics.Metrics {
	return l.metrics
}

func (l *LimitingHandler) Query() ThresholdUint {
	return l.query
}

func (l *LimitingHandler) Traffic() ThresholdUint {
	return l.traffic
}

func (l *LimitingHandler) Priority() int {
	return l.priority
}

func (l *LimitingHandler) Stop() bool {
	return l.stop
}

func NewLimitingHandler(conf *Config) (*LimitingHandler, error) {
	filter, err := strategy.ParseFilter(conf.Filters)
	if err != nil {
		return nil, err
	}

	mts := metrics.Parse(conf.Rule.Metrics)

	return &LimitingHandler{
		filter:   filter,
		metrics:  mts,
		stop:     conf.Stop,
		query:    parseThreshold(conf.Rule.Query),
		traffic:  parseThreshold(conf.Rule.Traffic),
		priority: conf.Priority,
	}, nil
}
