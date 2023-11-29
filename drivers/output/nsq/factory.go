package nsq

import (
	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
)

const name = "nsqd"

// Register 注册nsqd驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(name, NewFactory())
}
func NewFactory() eosc.IExtenderDriverFactory {

	return drivers.NewFactory[Config](Create)
}
