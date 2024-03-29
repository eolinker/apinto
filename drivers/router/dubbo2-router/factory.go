package dubbo2_router

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/common/bean"
)

var name = "dubbo2_router"

// Register 注册grpc路由驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(name, NewRouterDriverFactory())
}

// RouterDriverFactory dubbo路由驱动工厂结构体
type RouterDriverFactory struct {
	eosc.IExtenderDriverFactory
}

// Create 创建http路由驱动
func (r *RouterDriverFactory) Create(profession string, name string, label string, desc string, params map[string]interface{}) (eosc.IExtenderDriver, error) {
	once.Do(func() {
		bean.Autowired(&pluginManager)
		bean.Autowired(&routerManager)
	})

	return r.IExtenderDriverFactory.Create(profession, name, label, desc, params)

}

// NewRouterDriverFactory 创建一个http路由驱动工厂
func NewRouterDriverFactory() *RouterDriverFactory {
	return &RouterDriverFactory{
		IExtenderDriverFactory: drivers.NewFactory[Config](Create, Check),
	}
}
