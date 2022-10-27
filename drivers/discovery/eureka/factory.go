package eureka

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

var name = "discovery_eureka"

//Register 注册eureka驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(name, NewFactory())
}
func NewFactory() eosc.IExtenderDriverFactory {
	return drivers.NewFactory[Config](Create)
}
