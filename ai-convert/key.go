package ai_convert

import (
	"sync"
	"time"
)

type IKeyResource interface {
	ID() string
	Health() bool
	Priority() int
	// Up 上线
	Up()
	// Down 下线
	Down()

	IsBreaker() bool
	// Breaker 熔断
	Breaker()

	IConverter
}

type key struct {
	id       string
	name     string
	priority int
	disabled bool
	breaker  bool
	expired  int64
	locker   sync.RWMutex
	IConverter
}

func NewKey(id string, name string, expired int64, priority int, converter IConverter) IKeyResource {
	return &key{
		id:         id,
		name:       name,
		expired:    expired,
		priority:   priority,
		IConverter: converter,
	}
}

func (k *key) ID() string {
	return k.id
}

func (k *key) Priority() int {
	return k.priority
}

func (k *key) IsBreaker() bool {
	return k.breaker
}

func (k *key) Health() bool {
	k.locker.RLock()
	defer k.locker.RUnlock()
	if k.expired != 0 {
		if time.Now().Unix() > k.expired {
			k.disabled = true
		}
	}
	return !k.disabled
}

func (k *key) Up() {
	k.locker.Lock()
	defer k.locker.Unlock()
	k.disabled = false
	k.breaker = false
}

func (k *key) Down() {
	k.locker.Lock()
	defer k.locker.Unlock()
	k.disabled = true
}

func (k *key) Breaker() {
	k.locker.Lock()
	defer k.locker.Unlock()
	k.breaker = true
	k.disabled = true
}
