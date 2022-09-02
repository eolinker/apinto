package plugin_manager

import (
	"encoding/json"
	"errors"
	"fmt"

	"reflect"

	"github.com/eolinker/apinto/plugin"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/common/bean"
	"github.com/eolinker/eosc/eocontext"
	"github.com/eolinker/eosc/log"
)

var (
	errConfig                      = errors.New("invalid config")
	ErrorDriverNotExit             = errors.New("drive not exit")
	ErrorGlobalPluginMastConfig    = errors.New("global must have config")
	ErrorGlobalPluginConfigInvalid = errors.New("invalid global config")
)

type PluginManager struct {
	name            string
	extenderDrivers eosc.IExtenderDrivers
	plugins         Plugins
	pluginObjs      eosc.IUntyped
	workers         eosc.IWorkers
}

func (p *PluginManager) Set(conf interface{}) error {
	log.Debug("plugin manager set")
	return p.Reset(conf)
}

func (p *PluginManager) Get() interface{} {
	return p.plugins
}

func (p *PluginManager) ReadOnly() bool {
	return false
}

func (p *PluginManager) ConfigType() reflect.Type {
	return reflect.TypeOf(new(PluginWorkerConfig))
}

func (p *PluginManager) CreateRequest(id string, conf map[string]*plugin.Config) eocontext.IChain {

	return p.createChain(id, conf)
}

func (p *PluginManager) GetConfigType(name string) (reflect.Type, bool) {
	log.Debug("plugin manager get config type:", p.plugins)
	for _, plg := range p.plugins {
		if name == plg.Name {
			return plg.drive.ConfigType(), true
		}
	}
	return nil, false
}

func (p *PluginManager) Reset(conf interface{}) error {

	plugins, err := p.check(conf)
	if err != nil {
		return err
	}

	p.plugins = plugins
	list := p.pluginObjs.List()
	// 遍历，全量更新
	for _, obj := range list {
		v, ok := obj.(*PluginObj)
		if !ok {

			continue
		}
		v.Filters = p.createFilters(v.conf)
	}

	return nil
}

func (p *PluginManager) createFilters(conf map[string]*plugin.Config) []eocontext.IFilter {
	log.Debug("all plugins len: ", len(p.plugins))
	filters := make([]eocontext.IFilter, 0, len(conf))
	plugins := p.plugins
	for _, plg := range plugins {
		if plg.Status == StatusDisable {
			// 禁用插件，跳过
			continue
		}
		c := plg.Config
		if v, ok := conf[plg.Name]; ok {
			if v.Disable {
				// 局部禁用
				continue
			}
			if v.Config != nil {
				c = v.Config
			}
		} else if plg.Status != StatusGlobal {
			continue
		}
		confObj, err := toConfig(c, plg.drive.ConfigType())
		if err != nil {
			log.Error("plg manager: fail to createFilters filter,error is ", err)
			continue
		}
		worker, err := plg.drive.Create(fmt.Sprintf("%s@%s", plg.Name, p.name), plg.Name, confObj, nil)
		if err != nil {
			log.Error("plg manager: fail to createFilters filter,error is ", err)
			continue
		}
		fi, ok := worker.(eocontext.IFilter)
		if !ok {
			log.Error("extender ", plg.ID, " not plg for http-service.Filter")
			continue
		}
		filters = append(filters, fi)
	}
	return filters
}

func (p *PluginManager) createChain(id string, conf map[string]*plugin.Config) *PluginObj {

	chain := p.createFilters(conf)
	obj, has := p.pluginObjs.Get(id)
	if !has {
		obj = NewPluginObj(chain, id, conf)
		p.pluginObjs.Set(id, obj)
	} else {
		obj.(*PluginObj).Filters = chain
	}
	log.Debug("create chain len: ", len(chain))
	return obj.(*PluginObj)
}

func (p *PluginManager) check(conf interface{}) (Plugins, error) {
	cfg, ok := conf.(*PluginWorkerConfig)
	if !ok {
		return nil, errConfig
	}

	plugins := make(Plugins, 0, len(cfg.Plugins))
	for i, cf := range cfg.Plugins {
		log.DebugF("new plugin:%d=>%v", i, cf)
		newPlugin, err := p.newPlugin(cf)
		if err != nil {
			return nil, err
		}
		plugins = append(plugins, newPlugin)
	}
	return plugins, nil

}
func (p *PluginManager) Check(conf interface{}) error {
	_, err := p.check(conf)
	if err != nil {
		return err
	}
	return nil
}

func (p *PluginManager) IsExists(id string) bool {
	_, has := p.extenderDrivers.GetDriver(id)
	return has
}

func NewPluginManager() *PluginManager {

	pm := &PluginManager{

		plugins:    nil,
		pluginObjs: eosc.NewUntyped(),
	}
	log.Debug("autowired extenderDrivers")
	bean.Autowired(&pm.extenderDrivers)
	bean.Autowired(&pm.workers)

	log.DebugF("autowired extenderDrivers = %p", pm.extenderDrivers)

	return pm
}

func toConfig(v interface{}, t reflect.Type) (interface{}, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	obj := newConfig(t)
	err = json.Unmarshal(data, obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func newConfig(t reflect.Type) interface{} {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return reflect.New(t).Interface()
}
