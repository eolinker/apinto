package discover_eureka

import (
	"github.com/eolinker/eosc"
	"reflect"
)

func Register() {
	eosc.DefaultProfessionDriverRegister.RegisterProfessionDriver("eolinker:goku:discover_eureka", NewFactory())
}

type factory struct {
	profession string
	name       string
	label      string
	desc       string
	params     map[string]string
}

func NewFactory() *factory {
	return &factory{}
}

func (f *factory) ExtendInfo() eosc.ExtendInfo {
	return eosc.ExtendInfo{
		ID:      "eolinker:goku:discover_eureka",
		Group:   "eolinker",
		Project: "goku",
		Name:    "eureka",
	}
}

func (f *factory) Create(profession string, name string, label string, desc string, params map[string]string) (eosc.IProfessionDriver, error) {
	return &driver{
		profession: profession,
		name:       name,
		label:      label,
		desc:       desc,
		driver:     driverName,
		configType: reflect.TypeOf((*Config)(nil)),
		params:     params,
	}, nil
}
