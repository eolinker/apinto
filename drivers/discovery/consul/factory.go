package consul

import (
	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
)

var name = "discovery_consul"

//Register 注册consul驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(name, NewFactory())
}

func NewFactory() eosc.IExtenderDriverFactory {
	return drivers.NewFactory[Config](Create)
}
