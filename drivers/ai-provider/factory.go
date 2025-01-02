package ai_provider

import (
	"sync"

	"github.com/eolinker/eosc/common/bean"

	"github.com/eolinker/apinto/convert"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

var name = "ai-provider"

type Factory struct {
}

var (
	providerManager convert.IManager
	ones            sync.Once
)

// Register AI供应商Factory
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(name, NewFactory())
}

// NewFactory 创建service_http驱动工厂
func NewFactory() eosc.IExtenderDriverFactory {
	ones.Do(func() {
		bean.Autowired(&providerManager)
	})
	return drivers.NewFactory[Config](Create)
}
