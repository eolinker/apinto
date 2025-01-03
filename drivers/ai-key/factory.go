package ai_key

import (
	"sync"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/common/bean"

	"github.com/eolinker/apinto/convert"
)

const name = "ai-key"

var (
	providerManager convert.IManager
	ones            sync.Once
)

// Register AI Key Factory
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(name, NewFactory())
}

// NewFactory creates AI Key Factory
func NewFactory() eosc.IExtenderDriverFactory {
	ones.Do(func() {
		bean.Autowired(&providerManager)
	})
	return drivers.NewFactory[Config](Create)
}
