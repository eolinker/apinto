package syslog

import (
	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
)

const name = "syslog_output"

// Register 注册file_output驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(name, NewFactory())
}
func NewFactory() eosc.IExtenderDriverFactory {

	return drivers.NewFactory[Config](Create)
}
