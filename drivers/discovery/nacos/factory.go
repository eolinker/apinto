package nacos

import (
	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
)

var name = "discovery_nacos"

//Register 注册nacos驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(name, NewFactory())
}
func NewFactory() eosc.IExtenderDriverFactory {
	return drivers.NewFactory[Config](Create)
}
