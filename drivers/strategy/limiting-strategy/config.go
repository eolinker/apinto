package limiting_strategy

import (
	"fmt"
	"strings"

	"github.com/eolinker/apinto/strategy"
)

type Threshold struct {
	Second int64 `json:"second" label:"每秒限制"`
	Minute int64 `json:"minute" label:"每分钟限制"`
	Hour   int64 `json:"hour" label:"每小时限制"`
}

type Header struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// StrategyResponseConf 策略返回内容配置
type StrategyResponseConf struct {
	StatusCode  int       `json:"status_code" label:"HTTP状态码"`
	ContentType string    `json:"content_type" label:"Content-Type"`
	Charset     string    `json:"charset" label:"Charset"`
	Headers     []*Header `json:"header" label:"Header参数"`
	Body        string    `json:"body" label:"Body"`
}

func (s *StrategyResponseConf) SetBodyLabel(labels map[string]string) string {
	body := s.Body
	body = strings.ReplaceAll(body, "$api", fmt.Sprintf("%s(%s)", labels["api"], labels["api_id"]))
	body = strings.ReplaceAll(body, "$api_id", labels["api_id"])
	body = strings.ReplaceAll(body, "$api_name", labels["api"])

	body = strings.ReplaceAll(body, "$application", fmt.Sprintf("%s(%s)", labels["application"], labels["application_id"]))
	body = strings.ReplaceAll(body, "$application_id", labels["application_id"])
	body = strings.ReplaceAll(body, "$application_name", labels["application"])

	body = strings.ReplaceAll(body, "$service", labels["service"])
	body = strings.ReplaceAll(body, "$service_id", labels["service"])
	body = strings.ReplaceAll(body, "$service_name", labels["service"])
	body = strings.ReplaceAll(body, "ip", labels["ip"])

	return body
}

type Rule struct {
	Metrics  []string             `json:"metrics" label:"限流计数器名"`
	Query    Threshold            `json:"query" label:"请求限制" description:"按请求次数"`
	Traffic  Threshold            `json:"traffic" label:"流量限制" description:"按请求内容大小"`
	Response StrategyResponseConf `json:"response" label:"响应内容"`
}

type Config struct {
	Name        string                `json:"name" skip:"skip"`
	Description string                `json:"description" skip:"skip"`
	Stop        bool                  `json:"stop" label:"禁用"`
	Priority    int                   `json:"priority" label:"优先级" description:"1-999"`
	Filters     strategy.FilterConfig `json:"filters" label:"过滤规则"`
	Rule        Rule                  `json:"limiting" label:"限流规则" description:"限流规则"`
}

func parseThreshold(t Threshold, unit ...int64) ThresholdUint {
	u := int64(1)
	if len(unit) > 0 {
		u = unit[0]
	}
	if u < 1 {
		u = 1
	}
	return ThresholdUint{
		Second: t.Second * u,
		Minute: t.Minute * u,
		Hour:   t.Hour * u,
	}
}
