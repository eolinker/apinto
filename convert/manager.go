package convert

import "github.com/eolinker/eosc"

var _ IManager = (*Manager)(nil)

var (
	manager = NewManager()
)

type IManager interface {
	Get(id string) (IConverterDriver, bool)
	Set(id string, driver IConverterDriver)
	Del(id string)
}

type Manager struct {
	drivers eosc.Untyped[string, IConverterDriver]
}

func (m *Manager) Del(id string) {
	m.drivers.Del(id)
}

func (m *Manager) Get(id string) (IConverterDriver, bool) {
	return m.drivers.Get(id)
}

func (m *Manager) Set(id string, driver IConverterDriver) {
	m.drivers.Set(id, driver)
}

func NewManager() *Manager {
	return &Manager{drivers: eosc.BuildUntyped[string, IConverterDriver]()}
}

func Set(id string, driver IConverterDriver) {
	manager.Set(id, driver)
}

func Get(id string) (IConverterDriver, bool) {
	return manager.Get(id)
}

func Del(id string) {
	manager.Del(id)
}
