package kubernetes

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

var name = "discovery_kubernetes"

// Register 注册nacos驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(name, NewFactory())
}
func NewFactory() eosc.IExtenderDriverFactory {
	return drivers.NewFactory[Config](Create)
}
