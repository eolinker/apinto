package circuit_breaker

import (
	"sync"
)

type circuitCount struct {
	locker sync.RWMutex
	//UpdateTag          map[string]string
	CircuitBreakerInfo *CircuitBreakerInfo
}

func newCircuitCount() *circuitCount {
	return &circuitCount{
		locker:             sync.RWMutex{},
		CircuitBreakerInfo: nil,
	}
}

func (dr *circuitCount) getCircuitBreakerInfo() *CircuitBreakerInfo {

	dr.locker.RLock()
	value := dr.CircuitBreakerInfo
	dr.locker.RUnlock()

	if value == nil {
		dr.locker.Lock()
		value = dr.CircuitBreakerInfo

		if value == nil {
			value = initCircuitBreakerInfo()
			dr.CircuitBreakerInfo = value
		}
		dr.locker.Unlock()

	}

	return value
}

func (dr *circuitCount) writeCircuitBreakerInfo(info *CircuitBreakerInfo) {

	dr.locker.Lock()
	dr.CircuitBreakerInfo = info
	dr.locker.Unlock()
	return
}
