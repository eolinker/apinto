package limiting_strategy

import (
	"github.com/eolinker/apinto/strategy"
	"github.com/eolinker/apinto/utils/response"
	"github.com/eolinker/eosc/metrics"
)

type ThresholdUint struct {
	Second int64
	Minute int64
	Hour   int64
}
type LimitingHandler struct {
	name     string
	filter   strategy.IFilter
	metrics  metrics.Metrics
	query    ThresholdUint
	traffic  ThresholdUint
	response response.IResponse
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

func (l *LimitingHandler) Priority() int {
	return l.priority
}

func (l *LimitingHandler) Response() response.IResponse {
	return l.response
}

func (l *LimitingHandler) Stop() bool {
	return l.stop
}

func NewLimitingHandler(name string, conf *Config) (*LimitingHandler, error) {
	filter, err := strategy.ParseFilter(conf.Filters)
	if err != nil {
		return nil, err
	}

	mts := metrics.ParseArray(conf.Rule.Metrics)

	return &LimitingHandler{
		name:     name,
		filter:   filter,
		metrics:  mts,
		query:    parseThreshold(conf.Rule.Query),
		traffic:  parseThreshold(conf.Rule.Traffic, 1024*1024),
		response: response.Parse(&conf.Rule.Response),
		priority: conf.Priority,
		stop:     conf.Stop,
	}, nil
}
