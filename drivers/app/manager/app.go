package manager

import (
	"sync"

	"github.com/eolinker/apinto/drivers/router/http-router/manager"

	"github.com/eolinker/apinto/application/auth"
)

var _ IAppManager = (*AppManager)(nil)

type IAppManager interface {
	GetByAppID(appID string) []string
	GetByDriver(driver string) []string
	Set(appID string, driver string, ids []string)
	DelByDriver(driver string)
	DelByAppID(appID string)
}

type AppManager struct {
	apps   map[string]*AppData
	locker sync.RWMutex
}

func NewAppManager() *AppManager {
	return &AppManager{apps: make(map[string]*AppData)}
}

func (a *AppManager) GetByAppID(appID string) []string {
	a.locker.RLock()
	defer a.locker.RUnlock()
	newIDs := make([]string, 0, len(a.apps)*5)
	for _, app := range a.apps {
		ids, has := app.Get(appID)
		if has {
			newIDs = append(newIDs, ids...)
		}
	}
	return newIDs
}

func (a *AppManager) GetByDriver(driver string) []string {
	a.locker.RLock()
	defer a.locker.RUnlock()
	app, ok := a.apps[driver]
	if !ok {
		return nil
	}
	return app.All()
}

func (a *AppManager) Set(appID string, driver string, ids []string) {
	a.locker.Lock()
	defer a.locker.Unlock()
	app, ok := a.apps[driver]
	if ok {
		app.Set(appID, ids)
		return
	}

	app = NewAppData()
	app.Set(appID, ids)
	a.apps[driver] = app
	fac, _ := auth.GetFactory(driver)
	for _, r := range fac.PreRouters() {
		manager.AddPreRouter(r.ID, r.Method, r.Path, r.PreHandler)
	}
}

func (a *AppManager) DelByDriver(driver string) {
	a.locker.Lock()
	defer a.locker.Unlock()
	delete(a.apps, driver)
}

func (a *AppManager) DelByAppID(appID string) {
	a.locker.RLock()
	defer a.locker.RUnlock()
	for driver, app := range a.apps {
		app.Del(appID)
		ids := app.All()
		if len(ids) == 0 {
			fac, _ := auth.GetFactory(driver)
			for _, r := range fac.PreRouters() {
				manager.DeletePreRouter(r.ID)
			}
		}
	}
}

type AppData struct {
	data   map[string][]string
	locker sync.RWMutex
}

func NewAppData() *AppData {
	return &AppData{data: make(map[string][]string)}
}

func (a *AppData) Get(appID string) ([]string, bool) {
	a.locker.RLock()
	defer a.locker.RUnlock()
	v, ok := a.data[appID]
	return v, ok
}

func (a *AppData) Set(appID string, ids []string) {
	a.locker.Lock()
	defer a.locker.Unlock()
	a.data[appID] = ids
}

func (a *AppData) All() []string {
	a.locker.RLock()
	defer a.locker.RUnlock()
	idMap := make(map[string]bool)

	for _, v := range a.data {
		for _, id := range v {
			if _, ok := idMap[id]; !ok {
				idMap[id] = true
			}
		}
	}

	ids := make([]string, 0, len(idMap))
	for key := range idMap {
		ids = append(ids, key)
	}
	return ids
}

func (a *AppData) Del(appID string) {
	a.locker.Lock()
	defer a.locker.Unlock()
	delete(a.data, appID)
}
