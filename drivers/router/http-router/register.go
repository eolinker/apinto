package http_router

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/drivers/router/http-router/manager"
	"github.com/eolinker/apinto/plugin"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/common/bean"
	trafficConfig "github.com/eolinker/eosc/config"
	"github.com/eolinker/eosc/log"
	"github.com/eolinker/eosc/traffic"
)

var name = "http_router"

// Register 注册http路由驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(name, NewRouterDriverFactory())
}

// RouterDriverFactory http路由驱动工厂结构体
type RouterDriverFactory struct {
	eosc.IExtenderDriverFactory
}

// Create 创建http路由驱动
func (r *RouterDriverFactory) Create(profession string, name string, label string, desc string, params map[string]interface{}) (eosc.IExtenderDriver, error) {
	once.Do(func() {
		var tf traffic.ITraffic
		var cfg *trafficConfig.ListenUrl

		bean.Autowired(&tf)
		bean.Autowired(&cfg)
		bean.Autowired(&pluginManager)
		log.Debug("new router driver: ")
		bean.AddInitializingBeanFunc(func() {
			log.Debug("init router manager")

			routerManager = manager.NewManager(tf, cfg, pluginManager.CreateRequest("global", map[string]*plugin.Config{}))

		})
	})

	return r.IExtenderDriverFactory.Create(profession, name, label, desc, params)

}

// NewRouterDriverFactory 创建一个http路由驱动工厂
func NewRouterDriverFactory() *RouterDriverFactory {
	return &RouterDriverFactory{
		IExtenderDriverFactory: drivers.NewFactory[Config](Create, Check),
	}
}
