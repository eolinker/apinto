package syslog

import (
	"sync"

	"github.com/eolinker/eosc/common/bean"

	"github.com/eolinker/apinto/drivers"
	scope_manager "github.com/eolinker/apinto/drivers/scope-manager"
	"github.com/eolinker/eosc"
)

const name = "syslog_output"

var once = sync.Once{}
var scopeManager scope_manager.IManager

// Register 注册file_output驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(name, NewFactory())
}
func NewFactory() eosc.IExtenderDriverFactory {
	once.Do(func() {
		bean.Autowired(&scopeManager)
	})
	return drivers.NewFactory[Config](Create)
}
