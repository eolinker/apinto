package moonshot

import (
	"sync"

	"github.com/eolinker/apinto/convert"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

var name = "moonshot"

var (
	converterManager convert.IManager
	once             sync.Once
)

// Register 注册驱动
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(name, NewFactory())
}

// NewFactory 创建service_http驱动工厂
func NewFactory() eosc.IExtenderDriverFactory {
	return drivers.NewFactory[Config](Create)
}

// Create 创建驱动实例
func Create(id, name string, v *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	_, err := checkConfig(v)
	if err != nil {
		return nil, err
	}
	w := &executor{
		WorkerBase: drivers.Worker(id, name),
	}
	w.reset(v, workers)
	return w, nil
}
