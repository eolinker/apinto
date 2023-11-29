package request_interception

import (
	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
)

const (
	Name = "request_interception"
)

func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(Name, NewFactory())
}

func NewFactory() eosc.IExtenderDriverFactory {
	return drivers.NewFactory[Config](Create)
}
