package ai_key

import (
	"sync"
	"time"

	"github.com/eolinker/apinto/convert"
)

type key struct {
	id            string
	name          string
	priority      int
	disabled      bool
	breaker       bool
	expired       int64
	convertDriver convert.IConverterDriver
	locker        sync.RWMutex
}

func newKey(id string, name string, expired int64, priority int, convertDriver convert.IConverterDriver) convert.IKeyResource {
	return &key{
		id:            id,
		name:          name,
		expired:       expired,
		priority:      priority,
		convertDriver: convertDriver,
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

func (k *key) ConverterDriver() convert.IConverterDriver {
	return k.convertDriver
}
