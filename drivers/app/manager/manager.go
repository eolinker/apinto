package manager

import (
	"github.com/eolinker/apinto/application"
	"github.com/eolinker/eosc"
)

// 管理器：可以通过driver快速获取驱动列表

var _ IManager = (*Manager)(nil)

type IManager interface {
	Get(id string) (application.IAuth, bool)
	Set(appID string, labels map[string]string, disable bool, filters []application.IAuth)
	Del(appID string)
}

type Manager struct {
	// filters map[string]application.IAuthFilter
	filters       eosc.IUntyped
	groupByDriver map[string]RequireManager
}

func NewManager() *Manager {
	return &Manager{filters: eosc.NewUntyped()}
}

func (m *Manager) Get(id string) (application.IAuth, bool) {
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

func (m *Manager) List() []application.IAuthFilter {
	keys := m.filters.Keys()
	filters := make([]application.IAuthFilter, 0, len(keys))
	for _, key := range keys {
		filter, has := m.get(key)
		if !has {
			continue
		}
		filters = append(filters, filter)
	}
	return filters
}

func (m *Manager) ListByDriver(driver string) []application.IAuthFilter {
	c, has := m.getConnFilter(driver)
	if !has {
		return nil
	}
	ids := c.All()
	filters := make([]application.IAuthFilter, 0, len(ids))
	for _, id := range ids {
		filter, has := m.get(id)
		if has {
			filters = append(filters, filter)
		}
	}
	return filters
}

func (m *Manager) getConnFilter(driver string) (*connIDs, bool) {
	d, has := m.groupByDriver.Get(driver)
	if !has {
		return nil, false
	}
	v, ok := d.(*connIDs)
	if !ok {
		return nil, false
	}
	return v, true
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

func (m *Manager) Set(appID string, labels map[string]string, disable bool, filters []application.IAuth) {
	
	return
}

func (m *Manager) Del(appID string) {

}
