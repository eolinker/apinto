package http_router

import (
	"fmt"
	"github.com/eolinker/apinto/drivers/router/http-router/manager"
	"github.com/eolinker/apinto/plugin"
	"github.com/eolinker/apinto/service"
	"github.com/eolinker/apinto/template"
	"github.com/eolinker/eosc/common/bean"
	trafficConfig "github.com/eolinker/eosc/config"
	"github.com/eolinker/eosc/log"
	"github.com/eolinker/eosc/traffic"
	"github.com/eolinker/eosc/utils/config"
	"reflect"
	
	"github.com/eolinker/eosc"
)

//HTTPRouterDriver 实现github.com/eolinker/eosc.eosc.IProfessionDriver接口
type HTTPRouterDriver struct {
	configType    reflect.Type
	routerManager manager.IManger
	pluginManager plugin.IPluginManager
}

func (h *HTTPRouterDriver) Check(v interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	_, _, _, err := h.check(v, workers)
	if err != nil {
		return err
	}
	return nil
}

//NewHTTPRouter 创建一个http路由驱动
func NewHTTPRouterDriver() *HTTPRouterDriver {
	
	h := &HTTPRouterDriver{
		configType: reflect.TypeOf(new(Config)),
	}
	var tf traffic.ITraffic
	var cfg *trafficConfig.ListensMsg
	var pluginManager plugin.IPluginManager
	bean.Autowired(&tf)
	bean.Autowired(&cfg)
	bean.Autowired(&pluginManager)
	
	bean.AddInitializingBeanFunc(func() {
		log.Debug("init router manager")
		
		h.pluginManager = pluginManager
		h.routerManager = manager.NewManager(tf, cfg, pluginManager.CreateRequest("global", map[string]*plugin.Config{}))
		
	})
	return h
}

//Create 创建一个http路由驱动实例
func (h *HTTPRouterDriver) Create(id, name string, v interface{}, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	r := &HttpRouter{
		id:            id,
		name:          name,
		routerManager: h.routerManager,
		pluginManager: h.pluginManager,
	}
	
	err := r.reset(v, workers)
	if err != nil {
		return nil, err
	}
	return r, err
}

//check 检查http路由驱动配置
func (h *HTTPRouterDriver) check(v interface{}, workers map[eosc.RequireId]eosc.IWorker) (*Config, service.IService, template.ITemplate, error) {
	conf, ok := v.(*Config)
	if !ok {
		return nil, nil, nil, fmt.Errorf("get %s but %s %w", config.TypeNameOf(v), config.TypeNameOf(new(Config)), eosc.ErrorRequire)
	}
	ser, has := workers[conf.Service]
	if !has {
		return nil, nil, nil, fmt.Errorf("target %s: %w", conf.Service, eosc.ErrorRequire)
	}
	target, ok := ser.(service.IService)
	if !ok {
		return nil, nil, nil, fmt.Errorf("target name: %s type of %s,target %w", conf.Service, config.TypeNameOf(ser), eosc.ErrorNotGetSillForRequire)
	}
	var tmp template.ITemplate
	if conf.Template != "" {
		tp, has := workers[conf.Template]
		if !has {
			return nil, nil, nil, fmt.Errorf("target %s %w", conf.Template, eosc.ErrorRequire)
		}
		tmp, ok = tp.(template.ITemplate)
		if !ok {
			return nil, nil, nil, fmt.Errorf("target name: %s type of %s,target %w", conf.Template, config.TypeNameOf(tp), eosc.ErrorNotGetSillForRequire)
		}
	}
	return conf, target, tmp, nil
	
}

//ConfigType 返回http路由驱动配置的反射类型
func (h *HTTPRouterDriver) ConfigType() reflect.Type {
	return h.configType
}
