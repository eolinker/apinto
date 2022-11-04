package fileoutput

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

const name = "file_output"

//Register 注册file_output驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(name, NewFactory())
}

func NewFactory() eosc.IExtenderDriverFactory {
	return drivers.NewFactory[Config](Create, Check)
}
