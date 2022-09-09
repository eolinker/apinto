package limiting_stragety

import "github.com/eolinker/apinto/strategy"

type Threshold struct {
	Second int64 `json:"second" label:"每秒限制"`
	Minute int64 `json:"minute" label:"每分钟限制"`
	Hour   int64 `json:"hour" label:"每小时限制"`
}
type Rule struct {
	Metrics []string  `json:"metrics" label:"限流计数器名"`
	Query   Threshold `json:"query" label:"请求限制" description:"按请求次数"`
	Traffic Threshold `json:"traffic" label:"流量限制" description:"按请求内容大小"`
}
type Config struct {
	Profession  string                `json:"profession" skip:"skip"`
	Name        string                `json:"name" skip:"skip"`
	Driver      string                `json:"driver" skip:"skip"`
	Description string                `json:"description" skip:"skip"`
	Priority    int                   `json:"priority" label:"优先级" description:"1-999"`
	Filters     strategy.FilterConfig `json:"filters" label:"过滤规则"`

	Rule Rule `json:"limiting" label:"限流规则" description:"限流规则"`
}
