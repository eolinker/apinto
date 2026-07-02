package price_calcular

import (
	"github.com/eolinker/eosc"
)

// IManager 定价管理器接口
type IManager interface {
	Get(id string) (ICalculator, bool)
	Set(id string, executor ICalculator)
	Del(id string)
}

// Manager 实现 IManager 接口
type Manager struct {
	eosc.Untyped[string, ICalculator]
}

func NewManager() IManager {
	return &Manager{
		Untyped: eosc.BuildUntyped[string, ICalculator](),
	}
}

func (m *Manager) Del(id string) {
	m.Untyped.Del(id)
}

var (
	calculatorManager = NewManager()
)

func GetCalculator(key string) (ICalculator, bool) {
	return calculatorManager.Get(key)
}

func SetCalculator(key string, executor ICalculator) {
	calculatorManager.Set(key, executor)
}

func DelCalculator(key string) {
	calculatorManager.Del(key)
}
