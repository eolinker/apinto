package convert

import (
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

func newManager() *Manager {
	return &Manager{factories: eosc.BuildUntyped[string, IConverterFactory]()}
}
