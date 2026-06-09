package manager

import (
	"github.com/eolinker/eosc"
)

// IPolicyExecutor 定价规则策略执行器接口
type IPolicyExecutor interface {
	eosc.IWorker
	Calculator() interface{}
}

// IManager 定价管理器接口
type IManager interface {
	Get(id string) (IPolicyExecutor, bool)
	Set(id string, executor IPolicyExecutor)
	Del(id string)
}

// Manager 实现 IManager 接口
type Manager struct {
	eosc.Untyped[string, IPolicyExecutor]
}

func NewManager() IManager {
	return &Manager{
		Untyped: eosc.BuildUntyped[string, IPolicyExecutor](),
	}
}

func (m *Manager) Del(id string) {
	m.Untyped.Del(id)
}
