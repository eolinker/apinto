package protocbuf

import (
	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
)

var DriverName = "protobuf_transcode"

// Register 注册protobuf驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(DriverName, NewFactory())
}

// NewFactory 创建service_http驱动工厂
func NewFactory() eosc.IExtenderDriverFactory {

	return drivers.NewFactory[Config](Create)
}
