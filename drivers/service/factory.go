package service

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/drivers/discovery/static"
	ip_hash "github.com/eolinker/apinto/upstream/ip-hash"
	round_robin "github.com/eolinker/apinto/upstream/round-robin"
	"github.com/eolinker/eosc"
)

var DriverName = "service_http"
var (
	defaultHttpDiscovery = static.CreateAnonymous(&static.Config{
		Health:   nil,
		HealthOn: false,
	})
)

// Register 注册service_http驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(DriverName, NewFactory())
}

// NewFactory 创建service_http驱动工厂
func NewFactory() eosc.IExtenderDriverFactory {
	round_robin.Register()
	ip_hash.Register()
	return drivers.NewFactory[Config](Create)
}
