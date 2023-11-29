package prometheus

import (
	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
)

const name = "prometheus_output"

// Register 注册prometheus_output驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(name, NewFactory())
}

func NewFactory() eosc.IExtenderDriverFactory {

	return drivers.NewFactory[Config](Create, Check)
}
