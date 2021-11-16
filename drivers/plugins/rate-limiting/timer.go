package rate_limiting

import (
	"sync/atomic"
	"time"
)

const (
	rateSecond = iota
	rateMinute
	rateHour
	rateDay
)

var expireTime = []time.Duration{
	1 * time.Second, 1 * time.Minute, 1 * time.Hour, 24 * time.Hour,
}

type rateInfo struct {
	second *rateTimer
	minute *rateTimer
	hour   *rateTimer
	day    *rateTimer
}

func (r *rateInfo) close()  {
	r.second = nil
	r.minute = nil
	r.hour = nil
	r.day = nil
}

type rateTimer struct {
	// 已请求数量
	requestCount     int64
	// 限制数量
	limitCount	int64
	timerType int
	expire    time.Duration
	startTime time.Time

}

func CreateRateInfo(conf *Config) *rateInfo {
	info := &rateInfo{}
	if conf.Second > 0 {
		info.second = createTimer(rateSecond, conf.Second)
	}
	if conf.Minute > 0 {
		info.second = createTimer(rateMinute, conf.Minute)
	}
	if conf.Hour > 0 {
		info.second = createTimer(rateHour, conf.Hour)
	}
	if conf.Day > 0 {
		info.second = createTimer(rateDay, conf.Day)
	}
	return info
}

func createTimer(timerType int, limitCount int64) *rateTimer {
	return &rateTimer{
		timerType: timerType,
		requestCount: 0,
		limitCount: limitCount,
		startTime: time.Now(),
		expire:expireTime[timerType],
	}
}

func (r *rateTimer) add()  {
	atomic.AddInt64(&r.requestCount, 1)
}

func (r *rateTimer) reset()  {
	atomic.StoreInt64(&r.requestCount, 1)
	r.startTime = time.Now()
}

// 检查是否超过了限制
func (r *rateTimer) check() bool {
	if time.Now().Sub(r.startTime) > r.expire {
		r.reset()
		return true
	}
	c := atomic.LoadInt64(&r.requestCount)
	localCount := c + 1
	r.add()
	if localCount > r.limitCount {
		return false
	}
	return true
}