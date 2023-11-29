package cors

import (
	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
)

const (
	Name = "cors"
)

func Register(register eosc.IExtenderDriverRegister) {
	err := register.RegisterExtenderDriver(Name, NewFactory())
	if err != nil {
		return
	}
}

func NewFactory() eosc.IExtenderDriverFactory {
	return drivers.NewFactory[Config](Create)
}
