package cache_strategy

import "github.com/eolinker/apinto/strategy"

type Config struct {
	Name        string                `json:"name" skip:"skip"`
	Description string                `json:"description" skip:"skip"`
	Stop        bool                  `json:"stop" label:"禁用"`
	Priority    int                   `json:"priority" label:"优先级" description:"1-999"`
	Filters     strategy.FilterConfig `json:"filters" label:"过滤规则"`
	ValidTime   int                   `json:"valid_time" label:"有效期" description:"有效期"`
}
