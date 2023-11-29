package static

import (
	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
)

var name = "discovery_static"

// Register 注册静态服务发现的驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	_ = register.RegisterExtenderDriver(name, NewFactory())
}
func NewFactory() eosc.IExtenderDriverFactory {
	return drivers.NewFactory[Config](Create)
}
