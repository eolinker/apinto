package circuit_breaker

import (
	"github.com/eolinker/eosc"
	"reflect"
)

const (
	Name = "circuit_breaker"
)

func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(Name, NewFactory())
}

type Factory struct {
}

func NewFactory() *Factory {
	return &Factory{}
}

func (f *Factory) Create(profession string, name string, label string, desc string, params map[string]interface{}) (eosc.IExtenderDriver, error) {
	d := &Driver{
		profession: profession,
		name:       name,
		label:      label,
		desc:       desc,
		configType: reflect.TypeOf((*Config)(nil)),
	}
	return d, nil
}
