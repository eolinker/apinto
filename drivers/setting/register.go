package setting

import "github.com/eolinker/eosc"

//Register 注册http路由驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver("setting", NewFactory())
}

//Factory http路由驱动工厂结构体
type Factory struct {
}

//Create 创建http路由驱动
func (r *Factory) Create(profession string, name string, label string, desc string, params map[string]string) (eosc.IExtenderDriver, error) {
	return NewDriver(profession, name, label, desc, params), nil
}

//NewFactory 创建一个http路由驱动工厂
func NewFactory() *Factory {
	return &Factory{}
}
