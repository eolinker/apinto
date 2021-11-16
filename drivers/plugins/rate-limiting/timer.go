package rate_limiting

import "time"

const (
	rateSecond = iota
	rateMinute
	rateHour
	rateDay
)

type rateInfo struct {
	second *rateTimer
	minute *rateTimer
	hour   *rateTimer
	day    *rateTimer
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

func CreateTimer(timerType int) *rateTimer {
	return &rateTimer{
		timerType: timerType,
		count: 0,
		startTime: time.Now(),
	}
}