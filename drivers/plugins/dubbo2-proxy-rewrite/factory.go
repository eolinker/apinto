package grpc_proxy_rewrite

import (
	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
)

const (
	Name = "dubbo2-proxy-rewrite"
)

func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(Name, NewFactory())
}

func NewFactory() eosc.IExtenderDriverFactory {
	return drivers.NewFactory[Config](Create)
}
