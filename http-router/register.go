package http_router

import "github.com/eolinker/eosc"

func Register()  {
	eosc.DefaultProfessionDriverRegister.RegisterProfessionDriver("eolinker:goku:http_router",NewRouterDriverFactory())
}

type RouterDriverFactory struct {

}

func (r *RouterDriverFactory) ExtendInfo() eosc.ExtendInfo {
	panic("implement me")
}

func (r *RouterDriverFactory) Create(profession string, name string, label string, desc string, params map[string]string) (eosc.IProfessionDriver, error) {
	panic("implement me")
}

func NewRouterDriverFactory() *RouterDriverFactory {
	return &RouterDriverFactory{}
}

