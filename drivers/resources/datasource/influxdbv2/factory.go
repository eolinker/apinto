package influxdbv2

import (
	"reflect"
	"sync"

	"github.com/eolinker/eosc/common/bean"

	scope_manager "github.com/eolinker/apinto/drivers/scope-manager"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/utils/schema"
)

var (
	configType = reflect.TypeOf(new(Config))
	render     interface{}
)

var once = sync.Once{}
var scopeManager scope_manager.IManager

func init() {
	render, _ = schema.Generate(configType, nil)
}

func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver("influxdbv2", NewFactory())
}

// NewFactory 创建service_http驱动工厂
func NewFactory() eosc.IExtenderDriverFactory {
	once.Do(func() {
		bean.Autowired(&scopeManager)
	})
	return drivers.NewFactory[Config](Create)
}
