package proxy_mirror

import (
	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/drivers/discovery/static"
)

const (
	Name = "proxy_mirror"
)

var (
	defaultProxyDiscovery = static.CreateAnonymous(&static.Config{
		Health:   nil,
		HealthOn: false,
	})
)

func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(Name, NewFactory())
}

func NewFactory() eosc.IExtenderDriverFactory {
	return drivers.NewFactory[Config](Create, Check)
}
