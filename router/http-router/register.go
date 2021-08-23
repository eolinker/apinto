package http_router

import "github.com/eolinker/eosc"

var (
	driverInfo = eosc.ExtendInfo{
		ID:      "eolinker:goku:http_router",
		Group:   "eolinker",
		Project: "goku",
		Name:    "https_router",
	}
)

//Register 注册http路由驱动工厂
func Register() {
	eosc.DefaultProfessionDriverRegister.RegisterProfessionDriver(driverInfo.ID, NewRouterDriverFactory())
}

type RouterDriverFactory struct {
}

//ExtendInfo 返回http路由驱动工厂的信息
func (r *RouterDriverFactory) ExtendInfo() eosc.ExtendInfo {
	return driverInfo
}

//Create 创建http路由驱动
func (r *RouterDriverFactory) Create(profession string, name string, label string, desc string, params map[string]string) (eosc.IProfessionDriver, error) {
	return NewHttpRouter(profession, name, label, desc, params), nil
}

//NewRouterDriverFactory 创建一个http路由驱动工厂
func NewRouterDriverFactory() *RouterDriverFactory {
	return &RouterDriverFactory{}
}
