package credit

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

const DriverName = "credit"

// Register 代金券规则驱动注册
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(DriverName, NewFactory())
}

// NewFactory 创建代金券工厂
func NewFactory() eosc.IExtenderDriverFactory {
	return drivers.NewFactory[Config](Create)
}
