package http_router

import (
	_ "github.com/eolinker/apinto/router/router-http"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/utils/schema"
	"reflect"
)

var name = "http_router"

//Register 注册http路由驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(name, NewRouterDriverFactory())
}

//RouterDriverFactory http路由驱动工厂结构体
type RouterDriverFactory struct {
}

func (r *RouterDriverFactory) Render() interface{} {
	render, err := schema.Generate(reflect.TypeOf((*Config)(nil)), nil)
	if err != nil {
		return nil
	}
	return render
}

//Create 创建http路由驱动
func (r *RouterDriverFactory) Create(profession string, name string, label string, desc string, params map[string]interface{}) (eosc.IExtenderDriver, error) {

	return NewHTTPRouter(), nil
}

//NewRouterDriverFactory 创建一个http路由驱动工厂
func NewRouterDriverFactory() *RouterDriverFactory {
	return &RouterDriverFactory{}
}
