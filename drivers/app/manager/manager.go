package manager

import (
	"github.com/eolinker/apinto/application"
	"github.com/eolinker/apinto/application/auth"
	"github.com/eolinker/eosc"
)

var _ IManager = (*Manager)(nil)

type IManager interface {
	Get(id string) (application.IAuthFilter, bool)
	All() []application.IAuthFilter
	Set(appID string, driver string, tokenName string, position string, users []*application.User, rule interface{}) (string, error)
	Del(id string)
	DelByAppID(id string, appID string)
}

type Manager struct {
	// filters map[string]application.IAuthFilter
	filters eosc.IUntyped
}

func NewManager() *Manager {
	return &Manager{filters: eosc.NewUntyped()}
}

func (m *Manager) Get(id string) (application.IAuthFilter, bool) {
	return m.get(id)
}

func (m *Manager) get(id string) (application.IAuth, bool) {
	filter, has := m.filters.Get(id)
	if !has {
		return nil, false
	}
	f, ok := filter.(application.IAuth)
	return f, ok
}

func (m *Manager) all() []application.IAuthFilter {
	keys := m.filters.Keys()
	filters := make([]application.IAuthFilter, 0, 10*len(keys))
	for _, key := range keys {
		filter, has := m.filters.Get(key)
		if !has {
			continue
		}
		f, ok := filter.(application.IAuthFilter)
		if !ok {
			continue
		}
		filters = append(filters, f)
	}
	return filters
}

func (m *Manager) All() []application.IAuthFilter {
	return m.all()
}

func (m *Manager) Set(id string, driver string, tokenName string, position string, users []*application.User, rule interface{}) (string, error) {
	factory, err := auth.GetFactory(driver)
	if err != nil {
		return "", err
	}
	filter, err := factory.Create(tokenName, position, users, rule)
	if err != nil {
		return "", err
	}
	old, has := m.get(filter.ID())
	if has {
		old.Set(id, users)
	} else {
		m.filters.Set(filter.ID(), old)
	}
	return filter.ID(), nil
}

func (m *Manager) DelByAppID(id string, appID string) {
	filter, has := m.get(id)
	if !has {
		return
	}
	filter.Del(appID)
}

func (m *Manager) Del(id string) {
	m.filters.Del(id)
}
