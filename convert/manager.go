package convert

import (
	"sort"
	"time"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/common/bean"
)

var _ IManager = (*Manager)(nil)

var (
	manager = newManager()
)

func init() {
	bean.Injection(&manager)
}

type IManager interface {
	Get(id string) (IConverterFactory, bool)
	Set(id string, driver IConverterFactory)
	Del(id string)
}

type Manager struct {
	factories eosc.Untyped[string, IConverterFactory]
}

func (m *Manager) Del(id string) {
	m.factories.Del(id)
}

func (m *Manager) Get(id string) (IConverterFactory, bool) {
	return m.factories.Get(id)
}

func (m *Manager) Set(id string, driver IConverterFactory) {
	m.factories.Set(id, driver)
}

func newManager() IManager {
	return &Manager{factories: eosc.BuildUntyped[string, IConverterFactory]()}
}

var (
	keyPoolManager = NewKeyPoolManager()
)

type KeyPoolManager struct {
	keys     eosc.Untyped[string, KeyPool]
	keySorts eosc.Untyped[string, []IKeyResource]
}

func NewKeyPoolManager() *KeyPoolManager {
	return &KeyPoolManager{
		keys:     eosc.BuildUntyped[string, KeyPool](),
		keySorts: eosc.BuildUntyped[string, []IKeyResource](),
	}
}

type KeyPool eosc.Untyped[string, IKeyResource]

func (m *KeyPoolManager) KeyResources(id string) ([]IKeyResource, bool) {
	return m.keySorts.Get(id)
}

func (m *KeyPoolManager) Set(id string, resource IKeyResource) {
	keyPools, has := m.keys.Get(id)
	if !has {
		keyPools = eosc.BuildUntyped[string, IKeyResource]()
		m.keys.Set(id, keyPools)
	}
	keyPools.Set(resource.ID(), resource)
	keys := keyPools.List()
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Priority() < keys[j].Priority()
	})
	m.keySorts.Set(id, keys)
}

func (m *KeyPoolManager) DelKeySource(id, resourceId string) {
	keyPool, has := m.keys.Get(id)
	if !has {
		return
	}
	keyPool.Del(resourceId)
	keys := keyPool.List()
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Priority() > keys[j].Priority()
	})
	m.keySorts.Set(id, keys)
}

func (m *KeyPoolManager) Del(id string) {
	m.keys.Del(id)
	m.keySorts.Del(id)
}

func (m *KeyPoolManager) doLoop() {
	ticket := time.NewTicker(20 * time.Second)
	defer ticket.Stop()
	for {
		select {
		case <-ticket.C:
			for _, keyPool := range m.keys.List() {
				for _, key := range keyPool.List() {
					if key.IsBreaker() {
						key.Up()
					}
				}
			}
		}
	}
}

func SetKeyResource(provider string, resource IKeyResource) {
	keyPoolManager.Set(provider, resource)
}

func KeyResources(provider string) ([]IKeyResource, bool) {
	return keyPoolManager.KeyResources(provider)
}

func DelKeyResource(provider string, resourceId string) {
	keyPoolManager.DelKeySource(provider, resourceId)
}

func DelProvider(provider string) {
	providerManager.Del(provider)
	keyPoolManager.Del(provider)
}

var (
	providerManager = NewProviderManager()
)

type ProviderManager struct {
	providers     eosc.Untyped[string, IProvider]
	providerSorts []IProvider
}

func NewProviderManager() *ProviderManager {
	return &ProviderManager{
		providers: eosc.BuildUntyped[string, IProvider](),
	}
}

func (m *ProviderManager) Set(provider string, p IProvider) {
	m.providers.Set(provider, p)
	m.sortProviders()
}

func (m *ProviderManager) Get(provider string) (IProvider, bool) {
	return m.providers.Get(provider)
}

func (m *ProviderManager) sortProviders() {
	providers := m.providers.List()
	sort.Slice(providers, func(i, j int) bool {
		return providers[i].Priority() < providers[j].Priority()
	})
	m.providerSorts = providers
}

func (m *ProviderManager) Del(provider string) {
	m.providers.Del(provider)
	m.sortProviders()
}

func (m *ProviderManager) Providers() []IProvider {
	return m.providerSorts
}

func Providers() []IProvider {
	return providerManager.Providers()
}

func SetProvider(provider string, p IProvider) {
	providerManager.Set(provider, p)
}

func GetProvider(provider string) (IProvider, bool) {
	return providerManager.Get(provider)
}
