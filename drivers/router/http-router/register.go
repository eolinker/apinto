package http_router

import "github.com/eolinker/eosc"

//Register 注册http路由驱动工厂
func Register(register eosc.IExtenderRegister) {
	register.RegisterExtender("http_router", NewRouterDriverFactory())
}

//RouterDriverFactory http路由驱动工厂结构体
type RouterDriverFactory struct {
}

//Create 创建http路由驱动
func (r *RouterDriverFactory) Create(profession string, name string, label string, desc string, params map[string]string) (eosc.IExtenderDriver, error) {
	return NewHTTPRouter(profession, name, label, desc, params), nil
}

//NewRouterDriverFactory 创建一个http路由驱动工厂
func NewRouterDriverFactory() *RouterDriverFactory {
	return &RouterDriverFactory{}
}
