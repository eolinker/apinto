package dubbo2_to_http

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/drivers"
)

const (
	Name = "dubbo2_to_http"
)

func Register(register eosc.IExtenderDriverRegister) {
	err := register.RegisterExtenderDriver(Name, NewFactory())
	if err != nil {
		log.Warnf("register %s:%s", Name, err)
		return
	}
}

func NewFactory() eosc.IExtenderDriverFactory {
	return drivers.NewFactory[Config](Create)
}
