package fuse

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/utils/schema"
	"reflect"
)

const (
	Name = "strategy-plugin-fuse"
)

var (
	configType = reflect.TypeOf((*Config)(nil))
	render, _  = schema.Generate(configType, nil)
)

func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(Name, NewFactory())
}
func NewFactory() *factory {
	return &factory{}
}

type factory struct {
}

func (f *factory) Render() interface{} {
	return render
}

func (f *factory) Create(profession string, name string, label string, desc string, params map[string]interface{}) (eosc.IExtenderDriver, error) {

	return &driver{}, nil
}
