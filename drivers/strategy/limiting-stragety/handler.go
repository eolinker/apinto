package limiting_stragety

import (
	"github.com/eolinker/apinto/metrics"
	"github.com/eolinker/apinto/strategy"
)

type LimitingHandler struct {
	filter   strategy.IFilter
	metrics  metrics.Metrics
	query    Threshold
	traffic  Threshold
	priority int
	stop     bool
}

func (l *LimitingHandler) Filter() strategy.IFilter {
	return l.filter
}

func (l *LimitingHandler) Metrics() metrics.Metrics {
	return l.metrics
}

func (l *LimitingHandler) Query() Threshold {
	return l.query
}

func (l *LimitingHandler) Traffic() Threshold {
	return l.traffic
}

func (l *LimitingHandler) Priority() int {
	return l.priority
}

func (l *LimitingHandler) Stop() bool {
	return l.stop
}

func NewLimitingHandler(conf *ConfigCore) (*LimitingHandler, error) {
	filter, err := strategy.ParseFilter(conf.Filters)
	if err != nil {
		return nil, err
	}

	mts := metrics.Parse(conf.Rule.Metrics)

	return &LimitingHandler{
		filter:   filter,
		metrics:  mts,
		stop:     conf.Stop,
		query:    conf.Rule.Query,
		traffic:  conf.Rule.Traffic,
		priority: conf.Priority,
	}, nil
}
