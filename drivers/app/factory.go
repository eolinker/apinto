package app

import (
	"sync"

	"github.com/eolinker/apinto/application/auth"
	"github.com/eolinker/apinto/application/auth/aksk"
	"github.com/eolinker/apinto/application/auth/apikey"
	"github.com/eolinker/apinto/application/auth/basic"
	"github.com/eolinker/apinto/application/auth/jwt"
	"github.com/eolinker/apinto/application/auth/oauth2"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/drivers/app/manager"
	"github.com/eolinker/eosc/common/bean"

	"github.com/eolinker/eosc"
)

var name = "app"

var (
	appManager manager.IManager
	ones       sync.Once
)

// Register 注册service_http驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(name, NewFactory())
}

// NewFactory 创建service_http驱动工厂
func NewFactory() eosc.IExtenderDriverFactory {
	ones.Do(func() {
		apikey.Register()
		basic.Register()
		aksk.Register()
		jwt.Register()
		oauth2.Register()
		appManager = manager.NewManager(auth.Alias(), auth.Keys())
		bean.Injection(&appManager)
	})
	return drivers.NewFactory[Config](Create)
}
