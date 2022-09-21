package scalar

import "sync"

type Manager interface {
	Get(key string) Scalar
}

type Scalar interface {
	Second() Vectors
	Minute() Vectors
	Hour() Vectors
}
type _Scalar struct {
	second Vectors
	minute Vectors
	hour   Vectors
}

func (s *_Scalar) Second() Vectors {
	return s.second
}

func (s _Scalar) Minute() Vectors {
	return s.minute
}

func (s _Scalar) Hour() Vectors {
	return s.hour
}

type _Manager struct {
	lock   sync.RWMutex
	values map[string]Scalar
}

func (m *_Manager) Get(key string) Scalar {
	m.lock.RLock()
	scalar, has := m.values[key]
	m.lock.RUnlock()
	if has {
		return scalar
	}
	m.lock.Lock()
	defer m.lock.Unlock()
	scalar, has = m.values[key]
	if has {
		return scalar
	}
	scalar = &_Scalar{
		second: newVectors(1000, 500),
		minute: newVectors(60000, 5000),
		hour:   newVectors(3600000, 360000),
	}
	m.values[key] = scalar

	return scalar
}

func NewManager() *_Manager {
	return &_Manager{
		values: make(map[string]Scalar),
	}
}
