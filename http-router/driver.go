package http_router

import (
	"reflect"

	"github.com/eolinker/eosc"
)

type HttpRouterDriver struct {
	info       eosc.DriverInfo
	configType reflect.Type
}

func (h *HttpRouterDriver) Create(id, name string, v interface{}, workers map[string]interface{}) (eosc.IWorker, error) {
	panic("implement me")
}

func NewHttpRouter(profession, name, label, desc string, params map[string]string) *HttpRouterDriver {
	return &HttpRouterDriver{
		info: eosc.DriverInfo{
			Name:       name,
			Label:      label,
			Desc:       desc,
			Profession: profession,
			Params:     params,
		},
	}
}

func (h *HttpRouterDriver) ConfigType() reflect.Type {
	return h.configType
}
