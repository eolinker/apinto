package http_router

import (
	"fmt"
	"sync"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/drivers/router/http-router/manager"
	"github.com/eolinker/apinto/plugin"
	"github.com/eolinker/apinto/service"
	"github.com/eolinker/apinto/template"
	"github.com/eolinker/eosc/utils/config"

	"github.com/eolinker/eosc"
)

var (
	routerManager manager.IManger
	pluginManager plugin.IPluginManager
	once          sync.Once
)

func Check(v *Config, workers map[eosc.RequireId]eosc.IWorker) error {
	_, _, _, err := check(v, workers)
	if err != nil {
		return err
	}
	return nil
}

// Create 创建一个http路由驱动实例
func Create(id, name string, v *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	//log.Debug("create http router worker: ", pluginManager)
	r := &HttpRouter{
		id:            id,
		name:          name,
		routerManager: routerManager,
		pluginManager: pluginManager,
	}

	err := r.reset(v, workers)
	if err != nil {
		log.Error("create http router worker error: ", err)
		return nil, err
	}
	return r, err
}

// check 检查http路由驱动配置
func check(v interface{}, workers map[eosc.RequireId]eosc.IWorker) (*Config, service.IService, template.ITemplate, error) {
	conf, ok := v.(*Config)
	if !ok {
		return nil, nil, nil, fmt.Errorf("get %s but %s %w", config.TypeNameOf(v), config.TypeNameOf(new(Config)), eosc.ErrorRequire)
	}
	var target service.IService
	ser, has := workers[conf.Service]
	if has {
		target, ok = ser.(service.IService)
		if !ok {
			return nil, nil, nil, fmt.Errorf("target name: %s type of %s,target %w", conf.Service, config.TypeNameOf(ser), eosc.ErrorNotGetSillForRequire)
		}
		//return nil, nil, nil, fmt.Errorf("target %s: %w", conf.Service, eosc.ErrorRequire)
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
