package ai_provider

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/eolinker/apinto/convert"
)

type keyPool struct {
	closeChan chan struct{}
	disable   bool
	keys      []convert.IKeyResource
	locker    sync.RWMutex
}

func (k *keyPool) Close() {
	if k.closeChan != nil {
		close(k.closeChan)
		k.closeChan = nil
	}
}

func (k *keyPool) Health() bool {
	return !k.disable
}

func (k *keyPool) Down() {
	k.disable = true
}

func newKeyPool(ctx context.Context, cfg *Config) (*keyPool, map[string]interface{}, error) {
	factory, has := providerManager.Get(cfg.Provider)
	if !has {
		return nil, nil, errors.New("provider not found")
	}
	keys := make([]convert.IKeyResource, 0, len(cfg.Keys))
	var extender map[string]interface{}
	for _, v := range cfg.Keys {
		if v.Disabled {
			continue
		}
		cv, err := factory.Create(v.Config)
		if err != nil {
			return nil, nil, err
		}
		if extender == nil {
			fn, has := cv.GetModel(cfg.Model)
			if !has {
				return nil, nil, fmt.Errorf("default model not found")
			}
			extender, err = fn(cfg.ModelConfig)
			if err != nil {
				return nil, nil, err
			}
		}

		keys = append(keys, newKey(v.ID, v.Name, v.Expired, cv))
	}
	k := &keyPool{
		keys: keys,
	}
	go k.doLoop(ctx)
	return k, extender, nil
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
	index := k.index
	k.index++
	if index < k.size {
		return k.keys[index], true
	}
	return nil, false
}

type key struct {
	id            string
	name          string
	disabled      bool
	breaker       bool
	expired       int64
	convertDriver convert.IConverterDriver
	locker        sync.RWMutex
}

func newKey(id string, name string, expired int64, convertDriver convert.IConverterDriver) convert.IKeyResource {
	return &key{id: id, name: name, expired: expired, convertDriver: convertDriver}
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
