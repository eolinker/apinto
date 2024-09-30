package ai_service

import (
	"github.com/eolinker/apinto/drivers"
	iphash "github.com/eolinker/apinto/upstream/ip-hash"
	roundrobin "github.com/eolinker/apinto/upstream/round-robin"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/log"
)

var DriverName = "service_ai"

// Register 注册service_http驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	err := register.RegisterExtenderDriver(DriverName, NewFactory())
	if err != nil {
		log.Errorf("register %s %s", DriverName, err)
		return

	}
}

// NewFactory 创建service_http驱动工厂
func NewFactory() eosc.IExtenderDriverFactory {
	roundrobin.Register()
	iphash.Register()
	return drivers.NewFactory[Config](Create)
}
