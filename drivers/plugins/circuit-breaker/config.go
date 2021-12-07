package circuit_breaker

import "fmt"

type Config struct {
	MatchCodes      string            `json:"match_codes"`
	MonitorPeriod   int               `json:"monitor_period"`
	MinimumRequests int               `json:"minimum_requests"`
	FailurePercent  float64           `json:"failure_percent"`
	BreakPeriod     int64             `json:"break_period"`
	SuccessCounts   int               `json:"success_counts"`
	BreakerCode     int               `json:"breaker_code"`
	Headers         map[string]string `json:"headers"`
	Body            string            `json:"body"`
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
