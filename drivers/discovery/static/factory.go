package static

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/utils/schema"
	"reflect"
)

var name = "discovery_static"

//Register 注册静态服务发现的驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(name, NewFactory())
}

type factory struct {
}

func (f *factory) Render() interface{} {
	render, err := schema.Generate(reflect.TypeOf((*Config)(nil)), map[string][]string{"health_on": []string{}})
	if err != nil {
		return nil
	}
	return render
}

func (f *factory) ConfigType() reflect.Type {
	return reflect.TypeOf((*Config)(nil))
}

//NewFactory 创建静态服务发现的驱动工厂
func NewFactory() eosc.IExtenderDriverFactory {
	return &factory{}
}

//Create 创建静态服务发现驱动
func (f *factory) Create(profession string, name string, label string, desc string, params map[string]interface{}) (eosc.IExtenderDriver, error) {
	return &driver{
		profession: profession,
		name:       name,
		label:      label,
		desc:       desc,

		configType: reflect.TypeOf((*Config)(nil)),
	}, nil
}
