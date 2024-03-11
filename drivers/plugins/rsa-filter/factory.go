package rsa_filter

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

const (
	Name = "rsa"
)

func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(Name, NewFactory())
}

func NewFactory() eosc.IExtenderDriverFactory {
	return drivers.NewFactory[Config](Create)
}
