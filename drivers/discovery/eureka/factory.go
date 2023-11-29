package eureka

import (
	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
)

var name = "discovery_eureka"

//Register 注册eureka驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(name, NewFactory())
}
func NewFactory() eosc.IExtenderDriverFactory {
	return drivers.NewFactory[Config](Create)
}
