package cors

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/utils/schema"
	"reflect"
)

const (
	Name = "cors"
)

func Register(register eosc.IExtenderDriverRegister) {
	err := register.RegisterExtenderDriver(Name, NewFactory())
	if err != nil {
		return
	}
}

type Factory struct {
}

func (f *Factory) Render() *schema.Schema {
	render, err := schema.Generate(reflect.TypeOf((*Config)(nil)), nil)
	if err != nil {
		return nil
	}
	return render
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
