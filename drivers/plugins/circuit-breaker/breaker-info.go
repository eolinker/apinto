package circuit_breaker

import "time"

type CircuitBreakerInfo struct {
	CircuitBreakerState     int   `json:"circuit_breaker_state"`
	FailCounts              int   `json:"fail_counts"`
	SuccessCounts           int   `json:"success_counts"`
	TripTime                int64 `json:"trip_time"`
	StartTime               int64 `json:"start_time"`
	RecoveringSuccessCounts int   `json:"recovering_success_counts"`
}

// 重置熔断当前状态信息
func (cb *CircuitBreakerInfo) ResetCircuitBreakerInfo(state int, resetStartTime bool) {
	startTime := cb.StartTime
	if resetStartTime {
		startTime = time.Now().Unix()
	}
	recoveringSuccessCounts := cb.RecoveringSuccessCounts
	if state == BreakerDisable {
		recoveringSuccessCounts = 0
	}
	cb.CircuitBreakerState = state
	cb.FailCounts = 0
	cb.SuccessCounts = 0
	cb.RecoveringSuccessCounts = recoveringSuccessCounts
	cb.StartTime = startTime
	cb.TripTime = time.Now().Unix()
}

// 重置熔断当前状态信息
func initCircuitBreakerInfo() *CircuitBreakerInfo {
	info := &CircuitBreakerInfo{
		CircuitBreakerState:     BreakerDisable,
		FailCounts:              0,
		StartTime:               time.Now().Unix(),
		SuccessCounts:           0,
		TripTime:                time.Now().Unix(),
		RecoveringSuccessCounts: 0,
	}
	return info
}
