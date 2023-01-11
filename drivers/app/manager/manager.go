package manager

import (
	"strings"
	"sync"

	"github.com/eolinker/apinto/application"
	"github.com/eolinker/eosc"
)

// 管理器：可以通过driver快速获取驱动列表

var _ IManager = (*Manager)(nil)

type IManager interface {
	Get(id string) (application.IAuth, bool)
	List() []application.IAuthUser
	ListByDriver(driver string) []application.IAuthUser
	Set(app application.IApp, filters []application.IAuth, users map[string][]application.ITransformConfig)
	Del(appID string)
	Count() int
	AnonymousApp() application.IApp
	SetAnonymousApp(app application.IApp)
}

type Manager struct {
	// filters map[string]application.IAuthUser
	eosc.Untyped[string, application.IAuth]
	appManager  *AppManager
	driverAlias map[string]string
	drivers     []string
	locker      sync.RWMutex
	app         application.IApp
}

func (m *Manager) AnonymousApp() application.IApp {
	m.locker.RLock()
	app := m.app
	m.locker.RUnlock()
	return app
}

func (m *Manager) SetAnonymousApp(app application.IApp) {
	m.locker.Lock()
	m.app = app
	m.locker.Unlock()
}

func NewManager(driverAlias map[string]string, drivers []string) IManager {
	return &Manager{Untyped: eosc.BuildUntyped[string, application.IAuth](), appManager: NewAppManager(), driverAlias: driverAlias, drivers: drivers}
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
		filter, has := m.Get(id)
		if has {
			filters = append(filters, filter)
		}
	}
	return filters
}

func (m *Manager) all() []application.IAuthUser {
	list := m.List()
	filters := make([]application.IAuthUser, 0, len(list))
	for _, filter := range list {
		filters = append(filters, filter)
	}
	return filters
}

func (m *Manager) Set(app application.IApp, filters []application.IAuth, users map[string][]application.ITransformConfig) {
	idMap := make(map[string][]string)
	for _, filter := range filters {
		f, has := m.Untyped.Get(filter.ID())
		if !has {
			f = filter
		}
		v, ok := users[f.ID()]
		if !ok {
			continue
		}
		f.Set(app, v)
		m.Untyped.Set(filter.ID(), filter)
		if _, ok := idMap[filter.Driver()]; !ok {
			idMap[filter.Driver()] = make([]string, 0, len(filters))
		}
		idMap[filter.Driver()] = append(idMap[filter.Driver()], filter.ID())
	}
	for driver, ids := range idMap {
		old := m.appManager.GetByAppID(app.Id())
		m.appManager.Set(app.Id(), driver, ids)
		cs := compareArray(old, ids)
		for id := range cs {
			filter, has := m.Untyped.Get(id)
			if has {
				filter.Del(app.Id())
				if filter.UserCount() == 0 {
					m.Untyped.Del(id)
				}
			}
		}
	}

	return
}

func compareArray[T comparable](o, n []T) map[T]struct{} {
	m := make(map[T]struct{})
	for _, i := range o {
		m[i] = struct{}{}
	}
	for _, i := range n {
		if _, ok := m[i]; ok {
			delete(m, i)
		}
	}
	return m
}

func (m *Manager) Del(appID string) {
	ids := m.appManager.GetByAppID(appID)
	for _, id := range ids {
		filter, has := m.Untyped.Get(id)
		if has {
			filter.Del(appID)
			if filter.UserCount() == 0 {
				m.Untyped.Del(id)
			}
		}
	}
	m.appManager.DelByAppID(appID)
}
