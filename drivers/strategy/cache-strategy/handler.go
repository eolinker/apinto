package cache_strategy

import (
	"github.com/eolinker/apinto/strategy"
)

type CacheValidTimeHandler struct {
	name      string
	filter    strategy.IFilter
	validTime int
	priority  int
	stop      bool
}

func NewCacheValidTimeHandler(conf *Config) (*CacheValidTimeHandler, error) {
	filter, err := strategy.ParseFilter(conf.Filters)
	if err != nil {
		return nil, err
	}

	return &CacheValidTimeHandler{
		name:      conf.Name,
		filter:    filter,
		validTime: conf.ValidTime,
		priority:  conf.Priority,
		stop:      conf.Stop,
	}, nil
}
