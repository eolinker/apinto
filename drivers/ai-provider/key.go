package ai_provider

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/eolinker/apinto/convert"
)

var _ convert.IKeyPool = (*keyPool)(nil)

type keyPool struct {
	provider  string
	model     string
	priority  int
	closeChan chan struct{}
	keys      []convert.IKeyResource
	locker    sync.RWMutex
}

func (k *keyPool) Provider() string {
	return k.provider
}

func (k *keyPool) Model() string {
	return k.model
}

func (k *keyPool) Close() {
	if k.closeChan != nil {
		close(k.closeChan)
		k.closeChan = nil
	}
}

func newKeyPool(ctx context.Context, cfg *Config) (*keyPool, error) {
	factory, has := providerManager.Get(cfg.Provider)
	if !has {
		return nil, errors.New("provider not found")
	}
	keys := make([]convert.IKeyResource, 0, len(cfg.Keys))
	for _, v := range cfg.Keys {
		if v.Disabled {
			continue
		}
		cv, err := factory.Create(v.Config)
		if err != nil {
			return nil, err
		}

		keys = append(keys, &key{
			id:            v.ID,
			name:          v.Name,
			convertDriver: cv,
			locker:        sync.RWMutex{},
		})
	}
	k := &keyPool{
		provider: cfg.Provider,
		model:    cfg.Model,
		priority: cfg.Priority,
		keys:     keys,
	}
	go k.doLoop(ctx)
	return k, nil
}

func (k *keyPool) Selector() convert.IKeySelector {
	k.locker.RLock()
	keysCopy := append([]convert.IKeyResource{}, k.keys...) // 复制 keys
	k.locker.RUnlock()

	keys := make([]convert.IKeyResource, 0, len(keysCopy))
	for _, v := range keysCopy {
		if v.Health() {
			keys = append(keys, v)
		}
	}
	return newKeySelector(keys)
}

func (k *keyPool) doLoop(ctx context.Context) {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			k.locker.Lock()
			for _, v := range k.keys {
				v.Up()
			}
			k.locker.Unlock()
		case <-k.closeChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

func newKeySelector(keys []convert.IKeyResource) convert.IKeySelector {
	return &keySelector{
		keys: keys,
		size: len(keys)}
}

type keySelector struct {
	provider string
	model    string
	priority int
	keys     []convert.IKeyResource
	index    int
	size     int
}

func (k *keySelector) Provider() string {
	return k.provider
}

func (k *keySelector) DefaultModel() string {
	return k.model
}

func (k *keySelector) Priority() int {
	return k.priority
}

func (k *keySelector) Next() (convert.IKeyResource, bool) {
	for ; k.index < k.size; k.index++ {
		if k.keys[k.index].Health() {
			return k.keys[k.index], true
		}
	}
	return nil, false
}

type key struct {
	id            string
	name          string
	disabled      bool
	breaker       bool
	convertDriver convert.IConverterDriver
	locker        sync.RWMutex
}

func newKey(id string, name string, convertDriver convert.IConverterDriver) convert.IKeyResource {
	return &key{id: id, name: name, convertDriver: convertDriver}
}

func (k *key) Health() bool {
	k.locker.RLock()
	defer k.locker.RUnlock()
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
