package plugin_manager

import "github.com/eolinker/eosc"

//Register 插件注册
func Register() {

}

type Manager struct {
	plugins eosc.IUntyped
}

func (m *Manager) Get(id string) (interface{}, bool) {
	return m.plugins.Get(id)
}

func (m *Manager) Set() {

}

func NewManager() *Manager {
	return nil
}
