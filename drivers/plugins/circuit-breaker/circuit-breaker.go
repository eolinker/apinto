package circuit_breaker

import (
	"encoding/json"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/context"
	http_service "github.com/eolinker/eosc/context/http-context"
	"strconv"
	"time"
)

var _ context.IFilter = (*CircuitBreaker)(nil)
var _ http_service.HttpFilter = (*CircuitBreaker)(nil)

type CircuitBreaker struct {
	*Driver
	id      string
	counter *circuitCount
	conf    *Config
}

const (
	BreakerDisable    = 0
	BreakerEnable     = 1
	BreakerRecovering = 2
)

func (c *CircuitBreaker) Id() string {
	return c.id
}

func (c *CircuitBreaker) Start() error {
	return nil
}

func (c *CircuitBreaker) Reset(v interface{}, workers map[eosc.RequireId]interface{}) error {
	conf, err := c.check(v)
	if err != nil {
		return err
	}

	c.counter = newCircuitCount()
	c.conf = conf

	return nil
}

func (c *CircuitBreaker) Stop() error {
	return nil
}

func (c *CircuitBreaker) Destroy() {
	c.counter = nil
	c.conf = nil
}

func (c *CircuitBreaker) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}
func (c *CircuitBreaker) DoFilter(ctx context.Context, next context.IChain) (err error) {
	return http_service.DoHttpFilter(c, ctx, next)
}
func (c *CircuitBreaker) DoHttpFilter(ctx http_service.IHttpContext, next context.IChain) error {
	isContinue, err := c.access(ctx)
	if !isContinue {
		if err != nil {
			return err
		}
		return nil
	}

	if next != nil {
		err = next.DoChain(ctx)
	}
	if err != nil {
		return err
	}

	return c.proxy(ctx)
}

func (c *CircuitBreaker) access(ctx http_service.IHttpContext) (isContinue bool, e error) {

	if c.counter == nil {
		return true, nil
	}

	info := c.counter.getCircuitBreakerInfo()

	switch info.CircuitBreakerState {
	case BreakerDisable:
		{
			return true, nil
		}
	case BreakerEnable:
		{
			if (time.Now().Unix() - info.TripTime) >= c.conf.BreakPeriod {
				info.ResetCircuitBreakerInfo(BreakerRecovering, false)
				return true, nil
			} else {

				writeResponse(ctx, c.conf.Headers, c.conf.Body, c.conf.BreakerCode)
				return false, nil
			}
		}
	case BreakerRecovering:
		{
			if info.RecoveringSuccessCounts >= c.conf.SuccessCounts {
				info.ResetCircuitBreakerInfo(BreakerDisable, false)
				return true, nil
			}
		}
	}
	return true, nil
}

// 转发后执行
func (c *CircuitBreaker) proxy(ctx http_service.IHttpContext) (e error) {
	if c.counter == nil {
		return nil
	}

	info := c.counter.getCircuitBreakerInfo()

	switch info.CircuitBreakerState {
	case BreakerDisable:
		{
			if info.StartTime != 0 && (time.Now().Unix()-info.StartTime) <= int64(c.conf.MonitorPeriod) {
				// 请求错误率达到预设的值，并且该监控期内的请求总数达到最低熔断阀值，则进入熔断期
				if (info.FailCounts+info.SuccessCounts) >= c.conf.MinimumRequests && float64(info.FailCounts)/float64(info.FailCounts+info.SuccessCounts) >= c.conf.FailurePercent {
					info.CircuitBreakerState = BreakerEnable
					info.TripTime = time.Now().Unix()
				}
			} else {
				info.StartTime = time.Now().Unix()
				info.FailCounts = 0
				info.SuccessCounts = 0
			}
			if MatchStatusCode(c.conf.MatchCodes, ctx) {
				info.FailCounts++
			} else {
				info.SuccessCounts++
			}
			if info.CircuitBreakerState == BreakerEnable {
				info.FailCounts = 0
				info.SuccessCounts = 0
			}
		}
	case BreakerRecovering:
		{
			if MatchStatusCode(c.conf.MatchCodes, ctx) {
				info.FailCounts++
			} else {
				info.RecoveringSuccessCounts++
			}
			if info.FailCounts == 0 {
				if info.RecoveringSuccessCounts >= c.conf.SuccessCounts {
					info.CircuitBreakerState = BreakerDisable
					info.RecoveringSuccessCounts = 0
					info.SuccessCounts = 0
					info.StartTime = time.Now().Unix()
				}
			} else {
				info.CircuitBreakerState = BreakerEnable
				info.TripTime = time.Now().Unix()
				info.RecoveringSuccessCounts = 0
			}
		}
	}
	jsonInfo, _ := json.Marshal(info)
	ctx.Response().SetHeader("Fail-Counts", strconv.Itoa(info.FailCounts))
	ctx.Response().SetHeader("Success-Counts", strconv.Itoa(info.SuccessCounts))
	ctx.Response().SetHeader("Monitor-Info", string(jsonInfo))

	return nil
}
