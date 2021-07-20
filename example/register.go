package example

import (
	"github.com/eolinker/eosc"
	"reflect"
)

func Register() {
	eosc.DefaultProfessionDriverRegister.RegisterProfessionDriver("eolinker:goku:example", NewRouterDriverFactory())
}

type RouterDriverFactory struct {
}

func NewRouterDriverFactory() *RouterDriverFactory {
	return &RouterDriverFactory{}
}

func (r *RouterDriverFactory) ExtendInfo() eosc.ExtendInfo {
	return eosc.ExtendInfo{
		ID:      "eolinker:goku:example",
		Group:   "eolinker",
		Project: "goku",
		Name:    "example",
	}
}

func (r *RouterDriverFactory) Create(profession, name, label, desc string, params map[string]string) (eosc.IProfessionDriver, error) {
	if params == nil {
		params = make(map[string]string)
	}
	return &Driver{

		configType: reflect.TypeOf(new(Config)),
		params:     params,
	}, nil
}
