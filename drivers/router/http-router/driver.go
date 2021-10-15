package http_router

import (
	"fmt"
	"reflect"

	"github.com/eolinker/goku/service"

	"github.com/eolinker/eosc"
)

//HTTPRouterDriver 实现github.com/eolinker/eosc.eosc.IProfessionDriver接口
type HTTPRouterDriver struct {
	configType reflect.Type
}

//NewHTTPRouter 创建一个http路由驱动
func NewHTTPRouter(profession, name, label, desc string, params map[string]string) *HTTPRouterDriver {
	return &HTTPRouterDriver{
		configType: reflect.TypeOf(new(DriverConfig)),
	}
}

//Create 创建一个http路由驱动实例
func (h *HTTPRouterDriver) Create(id, name string, v interface{}, workers map[eosc.RequireId]interface{}) (eosc.IWorker, error) {
	conf, iService, err := h.check(v, workers)
	if err != nil {
		return nil, err
	}
	return NewRouter(id, name, conf, iService, h), nil
}

//check 检查http路由驱动配置
func (h *HTTPRouterDriver) check(v interface{}, workers map[eosc.RequireId]interface{}) (*DriverConfig, service.IService, error) {
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
		return nil, nil, fmt.Errorf("target name: %s type of %s,target %w", conf.Target, eosc.TypeNameOf(ser), eosc.ErrorNotGetSillForRequire)
	}
	return conf, target, nil

}

//ConfigType 返回http路由驱动配置的反射类型
func (h *HTTPRouterDriver) ConfigType() reflect.Type {
	return h.configType
}
