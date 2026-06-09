package pricing_policy

import (
	"sync"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/drivers/pricing-policy/manager"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/common/bean"
)

const name = "pricing-policy"

var (
	policyManager manager.IManager
	ones          sync.Once
)

// Register AI Key Factory
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(name, NewFactory())
}

// NewFactory creates AI Key Factory
func NewFactory() eosc.IExtenderDriverFactory {
	ones.Do(func() {
		policyManager = manager.NewManager()
		bean.Injection(&policyManager)
	})
	return drivers.NewFactory[Config](Create)
}
