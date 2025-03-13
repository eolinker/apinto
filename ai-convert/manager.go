package ai_convert

import (
	"sort"
	"time"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/common/bean"
)

var _ IManager = (*Manager)(nil)

var (
	manager = NewManager()
)

func init() {
	bean.Injection(&manager)
	bean.Injection(&modelAcManager)
}

type IManager interface {
	Get(id string) (IConverterCreateFunc, bool)
	Set(id string, driver IConverterCreateFunc)
	Del(id string)
}

type Manager struct {
	createFns eosc.Untyped[string, IConverterCreateFunc]
}

func (m *Manager) Del(id string) {
	m.createFns.Del(id)
}

func (m *Manager) Get(id string) (IConverterCreateFunc, bool) {
	return m.createFns.Get(id)
}

func (m *Manager) Set(id string, driver IConverterCreateFunc) {
	m.createFns.Set(id, driver)
}

func NewManager() IManager {
	return &Manager{createFns: eosc.BuildUntyped[string, IConverterCreateFunc]()}
}

func RegisterConverterCreateFunc(id string, fn IConverterCreateFunc) {
	manager.Set(id, fn)
}

func GetConverterCreateFunc(id string) (IConverterCreateFunc, bool) {
	return manager.Get(id)
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

func DelProvider(id string) {
	provider := balanceManager.Del(id)
	if provider != "" {
		keyPoolManager.Del(provider)
	}

}

var (
	balanceManager = NewBalanceManager()
)

type BalanceManager struct {
	providers   eosc.Untyped[string, IProvider]
	ids         eosc.Untyped[string, eosc.Untyped[string, IProvider]]
	balances    eosc.Untyped[string, IProvider]
	balanceSort []IProvider
}

func NewBalanceManager() *BalanceManager {
	return &BalanceManager{
		providers: eosc.BuildUntyped[string, IProvider](),
		balances:  eosc.BuildUntyped[string, IProvider](),
		ids:       eosc.BuildUntyped[string, eosc.Untyped[string, IProvider]](),
	}
}

func (m *BalanceManager) SetProvider(id string, p IProvider) {
	if p.Priority() < 1 {
		m.providers.Set(p.Provider(), p)
	}

	m.balances.Set(id, p)
	tmp, has := m.ids.Get(p.Provider())
	if !has {
		tmp = eosc.BuildUntyped[string, IProvider]()
	}
	tmp.Set(id, p)
	m.ids.Set(p.Provider(), tmp)
	m.sortBalances()
}

func (m *BalanceManager) Get(provider string) (IProvider, bool) {
	return m.providers.Get(provider)
}

func (m *BalanceManager) sortBalances() {
	balances := m.balances.List()
	tmpBalances := make([]IProvider, 0, len(balances))
	for _, b := range balances {
		if b.Priority() == 0 {
			continue
		}
		tmpBalances = append(tmpBalances, b)
	}
	sort.Slice(tmpBalances, func(i, j int) bool {
		return tmpBalances[i].Priority() < tmpBalances[j].Priority()
	})
	m.balanceSort = tmpBalances
}

func (m *BalanceManager) Del(id string) string {
	p, ok := m.balances.Del(id)
	if !ok {
		return ""
	}
	if p.Priority() == 0 {
		// 供应商本身
		m.providers.Del(p.Provider())
		m.ids.Del(p.Provider())
		m.sortBalances()
		return p.Provider()
	}
	tmp, has := m.ids.Get(p.Provider())
	if !has {
		return ""
	}
	tmp.Del(id)
	m.sortBalances()
	if tmp.Count() < 1 {
		m.providers.Del(p.Provider())
		return p.Provider()
	}

	return ""
}

func (m *BalanceManager) Balances() []IProvider {
	return m.balanceSort
}

func Balances() []IProvider {
	return balanceManager.Balances()
}

func SetProvider(id string, p IProvider) {
	balanceManager.SetProvider(id, p)
}

func GetProvider(provider string) (IProvider, bool) {
	return balanceManager.Get(provider)
}

type IModelAccessConfigManager interface {
	Get(id string) (IModelAccessConfig, bool)
	Set(id string, config IModelAccessConfig)
	Del(id string)
}

type IModelAccessConfig interface {
	Provider() string
	Model() string
	Config() map[string]string
}

type modelAccessConfigManager struct {
	configs eosc.Untyped[string, IModelAccessConfig]
}

func (m *modelAccessConfigManager) Get(id string) (IModelAccessConfig, bool) {
	return m.configs.Get(id)
}

func (m *modelAccessConfigManager) Set(id string, config IModelAccessConfig) {
	m.configs.Set(id, config)
}

func (m *modelAccessConfigManager) Del(id string) {
	m.configs.Del(id)
}

func NewModelAccessConfigManager() IModelAccessConfigManager {
	return &modelAccessConfigManager{
		configs: eosc.BuildUntyped[string, IModelAccessConfig](),
	}
}

var modelAcManager = NewModelAccessConfigManager()
