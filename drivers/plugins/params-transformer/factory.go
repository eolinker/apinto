package params_transformer

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/log"
)

const (
	Name = "params_transformer"
)

func Register(register eosc.IExtenderDriverRegister) {
	log.Debug("register params_transformer is ", Name)
	register.RegisterExtenderDriver(Name, NewFactory())
}
func NewFactory() eosc.IExtenderDriverFactory {
	return drivers.NewFactory[Config](Create, Check)
}
