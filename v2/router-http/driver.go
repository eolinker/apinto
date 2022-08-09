package http_router

import (
	"fmt"
	"github.com/eolinker/eosc/utils/config"
	"reflect"

	"github.com/eolinker/apinto/plugin"
	"github.com/eolinker/eosc/common/bean"

	service "github.com/eolinker/apinto/v2"

	"github.com/eolinker/eosc"
)

//HTTPRouterDriver 实现github.com/eolinker/eosc.eosc.IProfessionDriver接口
type HTTPRouterDriver struct {
	configType    reflect.Type
	pluginManager plugin.IPluginManager
}

//NewHTTPRouter 创建一个http路由驱动
func NewHTTPRouterDriver() *HTTPRouterDriver {

	h := &HTTPRouterDriver{
		configType: reflect.TypeOf(new(Config)),
	}
	bean.Autowired(&h.pluginManager)
	return h
}

//Create 创建一个http路由驱动实例
func (h *HTTPRouterDriver) Create(id, name string, v interface{}, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	r := &HttpRouter{
		id:   id,
		name: name,
	}

	err := r.reset(v, workers)
	if err != nil {
		return nil, err
	}
	return r, err
}

//check 检查http路由驱动配置
func (h *HTTPRouterDriver) check(v interface{}, workers map[eosc.RequireId]eosc.IWorker) (*Config, service.IService, error) {
	conf, ok := v.(*Config)
	if !ok {
		return nil, nil, fmt.Errorf("get %s but %s %w", config.TypeNameOf(v), config.TypeNameOf(new(Config)), eosc.ErrorRequire)
	}
	ser, has := workers[conf.Service]
	if !has {
		return nil, nil, fmt.Errorf("target %w", eosc.ErrorRequire)
	}
	target, ok := ser.(service.IService)
	if !ok {
		return nil, nil, fmt.Errorf("target name: %s type of %s,target %w", conf.Service, config.TypeNameOf(ser), eosc.ErrorNotGetSillForRequire)
	}
	return conf, target, nil

}

//ConfigType 返回http路由驱动配置的反射类型
func (h *HTTPRouterDriver) ConfigType() reflect.Type {
	return h.configType
}
