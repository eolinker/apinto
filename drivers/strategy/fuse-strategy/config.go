package fuse_strategy

import (
	"github.com/eolinker/apinto/strategy"
)

type Config struct {
	Name        string                `json:"name" skip:"skip"`
	Description string                `json:"description" skip:"skip"`
	Stop        bool                  `json:"stop"`
	Priority    int                   `json:"priority" label:"优先级" description:"1-999"`
	Filters     strategy.FilterConfig `json:"filters" label:"过滤规则"`
	Rule        Rule                  `json:"fuse" label:"熔断规则"`
}

type Rule struct {
	Metric           string               `json:"metric"`                      //熔断维度
	FuseCondition    StatusConditionConf  `json:"fuse_condition" label:"熔断条件"` //熔断条件
	FuseTime         FuseTimeConf         `json:"fuse_time" label:"熔断时间"`
	RecoverCondition StatusConditionConf  `json:"recover_condition" label:"恢复条件"` //恢复条件
	Response         StrategyResponseConf `json:"response" label:"响应内容"`
}

type StatusConditionConf struct {
	StatusCodes []int `json:"status_codes"`
	Count       int64 `json:"count"`
}

type FuseTimeConf struct {
	Time    int64 `json:"time"`
	MaxTime int64 `json:"max_time"`
}

//StrategyResponseConf 策略返回内容配置
type StrategyResponseConf struct {
	StatusCode  int      `json:"status_code"`
	ContentType string   `json:"content_type"`
	Charset     string   `json:"charset"`
	Header      []Header `json:"header,omitempty"` //key:value
	Body        string   `json:"body"`
}

type Header struct {
	key   string
	value string
}
