package manager

import (
	"github.com/eolinker/apinto/application"
	"github.com/eolinker/eosc"
	"strings"
	"sync"
)

// 管理器：可以通过driver快速获取驱动列表

var _ IManager = (*Manager)(nil)

type IManager interface {
	Get(id string) (application.IAuth, bool)
	List() []application.IAuthUser
	ListByDriver(driver string) []application.IAuthUser
	Set(appID string, labels map[string]string, disable bool, filters []application.IAuth, users map[string][]*application.User)
	Del(appID string)
	Count() int
}

type Manager struct {
	// filters map[string]application.IAuthUser
	filters     eosc.IUntyped
	appManager  *AppManager
	driverAlias map[string]string
	drivers     []string
	locker      sync.RWMutex
}

func (m *Manager) Count() int {
	return m.filters.Count()
}

func NewManager(driverAlias map[string]string, drivers []string) *Manager {
	return &Manager{filters: eosc.NewUntyped(), appManager: NewAppManager(), driverAlias: driverAlias, drivers: drivers}
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

func (m *Manager) List() []application.IAuthUser {
	filters := make([]application.IAuthUser, 0, len(m.drivers)*5)
	for _, driver := range m.drivers {
		filters = append(filters, m.getByDriver(driver)...)
	}
	return filters
}

func (m *Manager) ListByDriver(driver string) []application.IAuthUser {
	tmp := driver
	if v, ok := m.driverAlias[strings.ToLower(driver)]; ok {
		tmp = v
	}
	return m.getByDriver(tmp)
}

func (m *Manager) getByDriver(driver string) []application.IAuthUser {
	ids := m.appManager.GetByDriver(driver)
	filters := make([]application.IAuthUser, 0, len(ids))
	for _, id := range ids {
		filter, has := m.get(id)
		if has {
			filters = append(filters, filter)
		}
	}
	return filters
}

func (m *Manager) all() []application.IAuthUser {
	keys := m.filters.Keys()
	filters := make([]application.IAuthUser, 0, 10*len(keys))
	for _, key := range keys {
		filter, has := m.filters.Get(key)
		if !has {
			continue
		}
		f, ok := filter.(application.IAuthUser)
		if !ok {
			continue
		}
		filters = append(filters, f)
	}
	return filters
}

func (m *Manager) All() []application.IAuthUser {
	return m.all()
}

func (m *Manager) Set(appID string, labels map[string]string, disable bool, filters []application.IAuth, users map[string][]*application.User) {
	idMap := make(map[string][]string)
	for _, filter := range filters {
		f, has := m.get(filter.ID())
		if !has {
			f = filter
		}
		var us []*application.User
		if v, ok := users[f.ID()]; ok {
			us = v
		}
		f.Set(appID, labels, disable, us)
		m.filters.Set(filter.ID(), filter)
		if _, ok := idMap[filter.Driver()]; !ok {
			idMap[filter.Driver()] = make([]string, 0, len(filters))
		}
		idMap[filter.Driver()] = append(idMap[filter.Driver()], filter.ID())
	}
	for driver, ids := range idMap {
		m.appManager.Set(appID, driver, ids)
	}

	return
}

func (m *Manager) Del(appID string) {
	ids := m.appManager.GetByAppID(appID)
	for _, id := range ids {
		filter, has := m.get(id)
		if has {
			filter.Del(appID)
			if filter.UserCount() == 0 {
				m.filters.Del(id)
			}
		}
	}
	m.appManager.DelByAppID(appID)
}
