package nsq

import (
	"sync"

	scope_manager "github.com/eolinker/apinto/scope-manager"

	"github.com/eolinker/eosc/common/bean"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

const name = "nsqd"

var once = sync.Once{}
var scopeManager scope_manager.IManager

// Register 注册nsqd驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(name, NewFactory())
}
func NewFactory() eosc.IExtenderDriverFactory {
	once.Do(func() {
		bean.Autowired(&scopeManager)
	})
	return drivers.NewFactory[Config](Create)
}
