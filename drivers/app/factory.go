package app

import (
	"github.com/eolinker/apinto/application/auth"
	"github.com/eolinker/apinto/application/auth/aksk"
	"github.com/eolinker/apinto/application/auth/apikey"
	"github.com/eolinker/apinto/application/auth/basic"
	"github.com/eolinker/apinto/application/auth/jwt"
	"github.com/eolinker/apinto/drivers/app/manager"
	"github.com/eolinker/eosc/common/bean"
	"github.com/eolinker/eosc/utils/schema"
	"reflect"
	"sync"

	"github.com/eolinker/eosc"
)

var name = "app"

var (
	appManager manager.IManager
	ones       sync.Once
)

//Register 注册service_http驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(name, NewFactory())
}

type factory struct {
}

func (f *factory) Render() interface{} {
	render, err := schema.Generate(reflect.TypeOf((*Config)(nil)), nil)
	if err != nil {
		return nil
	}
	return render
}

//NewFactory 创建service_http驱动工厂
func NewFactory() eosc.IExtenderDriverFactory {
	ones.Do(func() {
		apikey.Register()
		basic.Register()
		aksk.Register()
		jwt.Register()
		appManager = manager.NewManager(auth.Alias(), auth.Keys())
		bean.Injection(&appManager)
	})
	return &factory{}
}

//Create 创建service_http驱动
func (f *factory) Create(profession string, name string, label string, desc string, params map[string]interface{}) (eosc.IExtenderDriver, error) {

	return &driver{
		profession: profession,
		label:      label,
		desc:       desc,
		driver:     name,
		configType: reflect.TypeOf((*Config)(nil)),
	}, nil
}
