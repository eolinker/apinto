package md5

import dynamic_params "github.com/eolinker/apinto/drivers/plugins/extra-params_v2/dynamic-params"

const name = "$md5"

func Register() {
	dynamic_params.Register(name, NewFactory())
}

func NewFactory() *Factory {
	return &Factory{}
}

type Factory struct {
}

func (f *Factory) Create(name string, value []string) (dynamic_params.IDynamicDriver, error) {
	return NewMD5(name, value), nil
}
