package limiting

import (
	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
)

const (
	Name = "strategy-plugin-limiting"
)

func Register(register eosc.IExtenderDriverRegister) {
	_ = register.RegisterExtenderDriver(Name, NewFactory())
}
func NewFactory() eosc.IExtenderDriverFactory {
	return drivers.NewFactory[Config](Create)
}
