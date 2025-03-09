package ai_key

import (
	"sync"
	"time"

	"github.com/eolinker/eosc/eocontext"

	ai_convert "github.com/eolinker/apinto/ai-convert"
)

type key struct {
	id        string
	name      string
	priority  int
	disabled  bool
	breaker   bool
	expired   int64
	converter ai_convert.IConverter
	locker    sync.RWMutex
}

func newKey(id string, name string, expired int64, priority int, converter ai_convert.IConverter) ai_convert.IKeyResource {
	return &key{
		id:        id,
		name:      name,
		expired:   expired,
		priority:  priority,
		converter: converter,
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

func (k *key) RequestConvert(ctx eocontext.EoContext, extender map[string]interface{}) error {
	return k.converter.RequestConvert(ctx, extender)
}

func (k *key) ResponseConvert(ctx eocontext.EoContext) error {
	return k.converter.ResponseConvert(ctx)
}
