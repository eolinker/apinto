package http_router

import (
	"fmt"
	"github.com/eolinker/eosc/utils/config"
	"reflect"

	"github.com/eolinker/apinto/plugin"
	"github.com/eolinker/eosc/common/bean"

	"github.com/eolinker/apinto/service"

	"github.com/eolinker/eosc"
)

//HTTPRouterDriver 实现github.com/eolinker/eosc.eosc.IProfessionDriver接口
type HTTPRouterDriver struct {
	configType    reflect.Type
	pluginManager plugin.IPluginManager
}

//NewHTTPRouter 创建一个http路由驱动
func NewHTTPRouter() *HTTPRouterDriver {

	h := &HTTPRouterDriver{
		configType: reflect.TypeOf(new(DriverConfig)),
	}
	bean.Autowired(&h.pluginManager)
	return h
}

//Create 创建一个http路由驱动实例
func (h *HTTPRouterDriver) Create(id, name string, v interface{}, workers map[eosc.RequireId]interface{}) (eosc.IWorker, error) {

	conf, iService, err := h.check(v, workers)
	if err != nil {
		return nil, err
	}
	r, err := h.NewRouter(id, name, conf, iService)
	if err != nil {
		return nil, err
	}
	return r, nil
}

//check 检查http路由驱动配置
func (h *HTTPRouterDriver) check(v interface{}, workers map[eosc.RequireId]interface{}) (*DriverConfig, service.IServiceCreate, error) {
	conf, ok := v.(*DriverConfig)
	if !ok {
		return nil, nil, fmt.Errorf("get %s but %s %w", config.TypeNameOf(v), config.TypeNameOf(new(DriverConfig)), eosc.ErrorRequire)
	}
	ser, has := workers[conf.Target]
	if !has {
		return nil, nil, fmt.Errorf("target %w", eosc.ErrorRequire)
	}
	target, ok := ser.(service.IServiceCreate)
	if !ok {
		return nil, nil, fmt.Errorf("target name: %s type of %s,target %w", conf.Target, config.TypeNameOf(ser), eosc.ErrorNotGetSillForRequire)
	}
	return conf, target, nil

}

//ConfigType 返回http路由驱动配置的反射类型
func (h *HTTPRouterDriver) ConfigType() reflect.Type {
	return h.configType
}

//NewRouter 创建http路由驱动实例
func (h *HTTPRouterDriver) NewRouter(id, name string, c *DriverConfig, target service.IServiceCreate) (*Router, error) {

	r := &Router{
		id:     id,
		name:   name,
		port:   c.Listen,
		driver: h,
	}
	routerHandler, err := r.create(c, target)

	if err != nil {
		return nil, err
	}
	r.handler = routerHandler
	return r, nil
}
