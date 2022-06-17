package circuit_breaker

import "fmt"

type Config struct {
	MatchCodes      string            `json:"match_codes" label:"匹配状态码" description:"多个状态码之间使用英文逗号隔开"`
	MonitorPeriod   int               `json:"monitor_period" label:"监控期" minimum:"1" description:"单位：秒，最小值：1"`
	MinimumRequests int               `json:"minimum_requests" label:"最低熔断阀值，达到熔断状态的最少请求次数" minimum:"1" description:"最小值：1"`
	FailurePercent  float64           `json:"failure_percent" label:"监控期内的请求错误率" minimum:"0" maximum:"1" description:"最小值：0，最大值：1"`
	BreakPeriod     int64             `json:"break_period" label:"熔断期" minimum:"1" description:"最小值：1"`
	SuccessCounts   int               `json:"success_counts" label:"连续请求成功次数，半开放状态下请求成功次数达到后会转变成健康状态" minimum:"1" description:"最小值：1"`
	BreakerCode     int               `json:"breaker_code" label:"熔断状态下返回的响应状态码" minimum:"100" description:"最小值：100"`
	Headers         map[string]string `json:"headers" label:"熔断状态下新增的返回头部值"`
	Body            string            `json:"body" label:"熔断状态下的返回响应体"`
}

var (
	breakerCodesErrInfo    = "[plugin circuit-breaker config err] breaker_codes must be not null. "
	minimumRequestsErrInfo = "[plugin circuit-breaker config err] minimum_requests must be greater than 0. "
	successCountsErrInfo   = "[plugin circuit-breaker config err] success_counts must be greater than 0. "
	statusCodeErrInfo      = "[plugin circuit-breaker config err] breaker_code must be in the range of [200,598]. err breaker_code: %d. "
	failurePercentErrInfo  = "[plugin circuit-breaker config err] failure_percent must be greater than 0 and less than 1. err failure_percent: %f. "
)

func (c *Config) doCheck() error {
	// 熔断条件的状态码 不能为空
	if c.MatchCodes == "" {
		return fmt.Errorf(breakerCodesErrInfo)
	}

	// FailurePercent 错误百分比要大于0，小于等于1
	if c.FailurePercent <= 0 || c.FailurePercent > 1 {
		return fmt.Errorf(failurePercentErrInfo, c.FailurePercent)
	}

	// BreakPeriod 熔断期 > 0, 没填或小于0则默认30
	if c.BreakPeriod <= 0 {
		c.BreakPeriod = 30
	}

	// MonitorPeriod 监控期 >0, 没填或小于0则默认30
	if c.MonitorPeriod <= 0 {
		c.MonitorPeriod = 30
	}

	// MinimumRequests 最低熔断阈值 > 0
	if c.MinimumRequests <= 0 {
		return fmt.Errorf(minimumRequestsErrInfo)
	}

	//SuccessCounts 连续成功次数 > 0
	if c.SuccessCounts <= 0 {
		return fmt.Errorf(successCountsErrInfo)
	}

	// 熔断返回的状态码 范围: [200,599]
	if c.BreakerCode < 200 || c.BreakerCode > 599 {
		return fmt.Errorf(statusCodeErrInfo, c.BreakerCode)
	}
	return nil
}
