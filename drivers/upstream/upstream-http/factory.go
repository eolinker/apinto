package upstream_http

import (
	"github.com/eolinker/eosc/utils/schema"
	"reflect"
	"sync"

	"github.com/eolinker/apinto/plugin"
	"github.com/eolinker/eosc/common/bean"

	round_robin "github.com/eolinker/apinto/upstream/round-robin"

	"github.com/eolinker/eosc"
)

var name = "upstream_http_proxy"
var (
	pluginManager plugin.IPluginManager
	once          sync.Once
)

//Register 注册http_proxy驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(name, NewFactory())

}

type factory struct {
}

//NewFactory 创建http_proxy驱动工厂
func NewFactory() eosc.IExtenderDriverFactory {
	round_robin.Register()
	return &factory{}
}

func (f *factory) Render() *schema.Schema {
	render, err := schema.Generate(reflect.TypeOf((*Config)(nil)), nil)
	if err != nil {
		return nil
	}
	return render
}

//Create 创建http_proxy驱动
func (f *factory) Create(profession string, name string, label string, desc string, params map[string]interface{}) (eosc.IExtenderDriver, error) {

	once.Do(func() {
		bean.Autowired(&pluginManager)
	})
	return &driver{
		profession: profession,
		name:       name,
		label:      label,
		desc:       desc,
		driver:     driverName,
		configType: reflect.TypeOf((*Config)(nil)),
	}, nil
}
