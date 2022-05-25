package auth

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/common/bean"
	"github.com/eolinker/eosc/utils/schema"
	"reflect"
)

const (
	Name = "auth"
)

func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(Name, NewFactory())
}

type Factory struct {
}

func NewFactory() *Factory {
	return &Factory{}
}

func (f *Factory) Render() *schema.Schema {
	render, err := schema.Generate(reflect.TypeOf((*Config)(nil)), nil)
	if err != nil {
		return nil
	}
	return render
}
func (f *Factory) Create(profession string, name string, label string, desc string, params map[string]interface{}) (eosc.IExtenderDriver, error) {
	d := &Driver{
		profession: profession,
		name:       name,
		label:      label,
		desc:       desc,
		configType: reflect.TypeOf((*Config)(nil)),
	}
	bean.Autowired(&d.workers)
	return d, nil
}
