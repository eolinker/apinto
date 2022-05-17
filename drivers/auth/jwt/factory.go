package jwt

import (
	"github.com/eolinker/eosc/utils/schema"
	"reflect"

	"github.com/eolinker/eosc"
)

var name = "auth_jwt"

//Register 注册jwt鉴权驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(name, NewFactory())
}

type factory struct {
}

//NewFactory 创建jwt鉴权驱动工厂
func NewFactory() eosc.IExtenderDriverFactory {
	return &factory{}
}
func (f *factory) Render() *schema.Schema {
	render, err := schema.Generate(reflect.TypeOf((*Config)(nil)), nil)
	if err != nil {
		return nil
	}
	return render
}

//Create 创建jwt鉴权驱动
func (f *factory) Create(profession string, name string, label string, desc string, params map[string]interface{}) (eosc.IExtenderDriver, error) {
	return &driver{
		profession: profession,
		name:       name,
		label:      label,
		desc:       desc,
		driver:     driverName,
		configType: reflect.TypeOf((*Config)(nil)),
	}, nil
}
