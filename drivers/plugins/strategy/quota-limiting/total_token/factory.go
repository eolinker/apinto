package total_token

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

const (
	Name = "strategy-plugin-quota-limiting-total-token"
)

func Register(register eosc.IExtenderDriverRegister) {
	_ = register.RegisterExtenderDriver(Name, NewFactory())
}

func NewFactory() eosc.IExtenderDriverFactory {
	return drivers.NewFactory[Config](Create, CheckConfig)
}
