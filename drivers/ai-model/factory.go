package ai_model

import (
	"sync"

	ai_convert "github.com/eolinker/apinto/ai-convert"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/common/bean"
)

var name = "ai-model"

var (
	once                sync.Once
	accessConfigManager ai_convert.IModelAccessConfigManager
)

type Factory struct {
}

// Register AI供应商Factory
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(name, NewFactory())
}

// NewFactory 创建service_http驱动工厂
func NewFactory() eosc.IExtenderDriverFactory {
	once.Do(func() {
		bean.Autowired(&accessConfigManager)
	})
	return drivers.NewFactory[Config](Create)
}
