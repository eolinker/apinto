package http_router

import (
	"fmt"
	"reflect"

	"github.com/eolinker/goku/service"

	"github.com/eolinker/eosc"
)

type HttpRouterDriver struct {
	info       eosc.DriverInfo
	configType reflect.Type
}

func (h *HttpRouterDriver) Create(id, name string, v interface{}, workers map[eosc.RequireId]interface{}) (eosc.IWorker, error) {
	conf, iService, err := h.check(v, workers)
	if err != nil {
		return nil, err
	}
	return NewRouter(id, name, conf, iService), nil
}

func (h *HttpRouterDriver) check(v interface{}, workers map[eosc.RequireId]interface{}) (*DriverConfig, service.IService, error) {
	conf, ok := v.(*DriverConfig)
	if !ok {
		return nil, nil, fmt.Errorf("get %s but %s %w", eosc.TypeNameOf(v), eosc.TypeNameOf(new(DriverConfig)), eosc.ErrorRequire)
	}

	ser, has := workers[conf.Target]
	if !has {
		return nil, nil, fmt.Errorf("target %w", eosc.ErrorRequire)
	}
	target, ok := ser.(service.IService)
	if !ok {
		return nil, nil, fmt.Errorf("target %w", eosc.ErrorNotGetSillForRequire)
	}
	return conf, target, nil

}
func NewHttpRouter(profession, name, label, desc string, params map[string]string) *HttpRouterDriver {
	return &HttpRouterDriver{
		configType: reflect.TypeOf(new(DriverConfig)),
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
