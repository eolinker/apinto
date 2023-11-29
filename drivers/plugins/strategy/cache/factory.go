package cache

import (
	"reflect"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/utils/schema"

	"github.com/eolinker/apinto/drivers"
)

const (
	Name = "strategy-plugin-cache"
)

var (
	configType = reflect.TypeOf((*Config)(nil))
	render, _  = schema.Generate(configType, nil)
)

func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(Name, NewFactory())
}
func NewFactory() eosc.IExtenderDriverFactory {
	return drivers.NewFactory[Config](Create)
}
