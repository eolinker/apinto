package strategy

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/utils/schema"
	"reflect"
)

const (
	Name = "strategy"
)

func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(Name, NewFactory())
}
func NewFactory() *factory {
	return &factory{
		configType: reflect.TypeOf((*Config)(nil)),
	}
}

type factory struct {
	configType reflect.Type
}

func (f *factory) Render() interface{} {
	render, err := schema.Generate(f.configType, nil)
	if err != nil {
		return nil
	}
	return render

}

func (f *factory) Create(profession string, name string, label string, desc string, params map[string]interface{}) (eosc.IExtenderDriver, error) {

	return &driver{
		configType: f.configType,
	}, nil
}
