package plugin_manager

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/eolinker/goku/plugin"

	"github.com/eolinker/eosc/common/bean"

	"github.com/eolinker/eosc"
	http_service "github.com/eolinker/eosc/http-service"
	"github.com/eolinker/eosc/log"
	"github.com/eolinker/goku/filter"
)

var (
	errConfig                      = errors.New("invalid config")
	ErrorDriverNotExit             = errors.New("drive not exit")
	ErrorGlobalPluginMastConfig    = errors.New("global mast have config")
	ErrorGlobalPluginConfigInvalid = errors.New("invalid global config")
)

type PluginManager struct {
	id string

	profession      string
	name            string
	extenderDrivers eosc.IExtenderDrivers
	plugins         Plugins
	pluginObjs      eosc.IUntyped
}

func (p *PluginManager) CreateRouter(id string, conf map[string]*plugin.Config) plugin.IPlugin {
	return p.newChain(id, conf, pluginRouter)
}

func (p *PluginManager) CreateService(id string, conf map[string]*plugin.Config) plugin.IPlugin {
	return p.newChain(id, conf, pluginService)
}

func (p *PluginManager) CreateUpstream(id string, conf map[string]*plugin.Config) plugin.IPlugin {
	return p.newChain(id, conf, pluginUpstream)
}

func (p *PluginManager) Id() string {

	return p.id
}

func (p *PluginManager) Start() error {
	return nil
}

func (p *PluginManager) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {

	plugins, err := p.check(conf)
	if err != nil {
		return err
	}

	p.plugins = plugins

	// 遍历，全量更新
	for _, obj := range p.pluginObjs.All() {
		v, ok := obj.(*PluginObj)
		if !ok {
			continue
		}
		v.IChainHandler.Reset(p.createFilters(v.conf, v.t)...)
	}

	return nil
}

func (p *PluginManager) Stop() error {
	return nil
}

func (p *PluginManager) CheckSkill(skill string) bool {
	return false
}

func (p *PluginManager) RemoveObj(id string) (*PluginObj, bool) {

	value, ok := p.pluginObjs.Del(id)
	if !ok {
		return nil, false
	}
	v, ok := value.(*PluginObj)
	return v, ok
}

func (p *PluginManager) createFilters(conf map[string]*plugin.Config, t string) []http_service.IFilter {
	filters := make([]http_service.IFilter, 0, len(conf))
	for _, plugin := range p.plugins {
		if plugin.Status == StatusDisable || plugin.Status == "" || plugin.Type != t {
			// 当插件类型不匹配，跳过
			continue
		}
		c := plugin.Config
		if v, ok := conf[plugin.Name]; ok {
			if v.Disable {
				// 不启用该插件
				continue
			}
			if plugin.Status != StatusGlobal && plugin.Status != StatusEnable {
				continue
			}
			c = v
		} else if plugin.Status != StatusGlobal && plugin.Status != StatusEnable {
			continue
		}
		confObj, err := toConfig(c, plugin.drive.ConfigType())
		if err != nil {
			log.Error("plugin manager: fail to createFilters filter,error is ", err)
			continue
		}
		worker, err := plugin.drive.Create(fmt.Sprintf("%s@%s", plugin.Name, p.name), plugin.Name, confObj, nil)
		if err != nil {
			log.Error("plugin manager: fail to createFilters filter,error is ", err)
			continue
		}
		fi, ok := worker.(http_service.IFilter)
		if !ok {
			log.Error("extender ", plugin.ID, " not plugin for http-service.Filter")
			continue
		}
		filters = append(filters, fi)
	}
	return filters
}

func (p *PluginManager) newChain(id string, conf map[string]*plugin.Config, t string) *PluginObj {
	chain := filter.NewChain(p.createFilters(conf, t))
	obj := &PluginObj{
		IChainHandler: chain,
		id:            id,
		conf:          conf,
		t:             t,
	}
	p.pluginObjs.Set(fmt.Sprintf("%s:%s", id, t), obj)
	return obj
}

func (p *PluginManager) check(conf interface{}) (Plugins, error) {
	cfg, ok := conf.(*PluginWorkerConfig)
	if !ok {
		return nil, errConfig
	}

	plugins := make(Plugins, 0, len(cfg.Plugins))
	for _, cf := range cfg.Plugins {
		plugin, err := p.newPlugin(cf)
		if err != nil {
			return nil, err
		}
		plugins = append(plugins, plugin)
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

func NewPluginManager(profession, name string) *PluginManager {

	pm := &PluginManager{
		id:         fmt.Sprintf("%s@%s", name, profession),
		profession: profession,
		name:       name,
		plugins:    nil,
		pluginObjs: eosc.NewUntyped(),
	}

	bean.Autowired(&pm.extenderDrivers)
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
