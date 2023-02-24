package protocbuf

import (
	"github.com/eolinker/apinto/drivers"
	round_robin "github.com/eolinker/apinto/upstream/round-robin"
	"github.com/eolinker/eosc"
)

var DriverName = "protobuf_transcode"

// Register 注册protobuf驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(DriverName, NewFactory())
}

// NewFactory 创建service_http驱动工厂
func NewFactory() eosc.IExtenderDriverFactory {
	round_robin.Register()
	return drivers.NewFactory[Config](Create)
}
