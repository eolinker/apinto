package polaris

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

var name = "discovery_polaris"

// Register 注册北极星驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	_ = register.RegisterExtenderDriver(name, NewFactory())
}

func NewFactory() eosc.IExtenderDriverFactory {
	return drivers.NewFactory[Config](Create)
}
