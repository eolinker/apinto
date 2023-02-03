package scope_manager

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/common/bean"
	"sync"
)

func init() {
	manager := NewManager()
	bean.Injection(&manager)
}

var _ IManager = (*Manager)(nil)

type Manager struct {
	scopes     eosc.Untyped[string, IProxy]
	connScope  eosc.Untyped[string, []string]
	connOutput eosc.Untyped[string, eosc.Untyped[string, interface{}]]
	locker     sync.Mutex
}

func NewManager() IManager {

	return &Manager{
		scopes:     eosc.BuildUntyped[string, IProxy](),
		connScope:  eosc.BuildUntyped[string, []string](),
		connOutput: eosc.BuildUntyped[string, eosc.Untyped[string, interface{}]](),
	}
}

func (m *Manager) Get(scopeName string) IProxyOutput {
	proxy, has := m.scopes.Get(scopeName)
	if !has {
		m.locker.Lock()
		defer m.locker.Unlock()
		proxy, has = m.scopes.Get(scopeName)
		if !has {
			proxy = NewProxy()
			m.scopes.Set(scopeName, proxy)
		}
	}
	return proxy
}

func (m *Manager) Set(name string, value interface{}, scopes []string) {
	m.locker.Lock()
	defer m.locker.Unlock()
	m.del(name)
	m.set(name, value, scopes)
	m.rebuild()
}

func (m *Manager) set(name string, value interface{}, scopes []string) {
	if len(scopes) < 1 {
		return
	}
	m.connScope.Set(name, scopes)
	for _, scope := range scopes {
		output, has := m.connOutput.Get(scope)
		if !has {
			output = eosc.BuildUntyped[string, interface{}]()
			m.connOutput.Set(scope, output)
		}
		output.Set(name, value)
	}
}

func (m *Manager) rebuild() {
	outputs := m.connOutput.All()
	for key, value := range outputs {
		proxy, has := m.scopes.Get(key)
		if !has {
			proxy = NewProxy()
		}
		proxy.Set(value.List())
		m.scopes.Set(key, proxy)
	}
}

func (m *Manager) Del(name string) {
	m.locker.Lock()
	defer m.locker.Unlock()
	m.del(name)
	m.rebuild()
}

func (m *Manager) del(name string) {
	scopes, has := m.connScope.Del(name)
	if has {
		for _, scope := range scopes {
			output, has := m.connOutput.Get(scope)
			if !has {
				continue
			}
			output.Del(name)
		}
	}
}

type IManager interface {
	Get(scopeName string) IProxyOutput
	Set(name string, value interface{}, scopes []string)
	Del(name string)
}
