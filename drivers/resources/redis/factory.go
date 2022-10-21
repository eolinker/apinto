package redis

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/utils/schema"
	"reflect"
)

var (
	configType = reflect.TypeOf(new(Config))
	render     interface{}
)

func init() {
	render, _ = schema.Generate(configType, nil)

}

func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver("redis", new(Factory))
}

type Factory struct {
}

func (f *Factory) Render() interface{} {
	return render
}

func (f *Factory) Create(profession string, name string, label string, desc string, params map[string]interface{}) (eosc.IExtenderDriver, error) {
	return new(Driver), nil
}
