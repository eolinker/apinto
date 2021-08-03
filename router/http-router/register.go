package http_router

import "github.com/eolinker/eosc"

var (
	driverInfo = eosc.ExtendInfo{
		ID:      "eolinker:goku:http_router",
		Group:   "eolinker",
		Project: "goku",
		Name:    "https_router",
	}
)

func Register() {
	eosc.DefaultProfessionDriverRegister.RegisterProfessionDriver(driverInfo.ID, NewRouterDriverFactory())
}

type RouterDriverFactory struct {
}

func (r *RouterDriverFactory) ExtendInfo() eosc.ExtendInfo {
	return driverInfo
}

func (r *RouterDriverFactory) Create(profession string, name string, label string, desc string, params map[string]string) (eosc.IProfessionDriver, error) {
	return NewHttpRouter(profession, name, label, desc, params), nil
}

func NewRouterDriverFactory() *RouterDriverFactory {
	return &RouterDriverFactory{}
}
