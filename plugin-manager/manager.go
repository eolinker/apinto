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
	ErrorGlobalPluginMastConfig    = errors.New("global must have config")
	ErrorGlobalPluginConfigInvalid = errors.New("invalid global config")
)

type PluginManager struct {
	id string

	profession      string
	name            string
	extenderDrivers eosc.IExtenderDrivers
	plugins         Plugins
	pluginObjs      eosc.IUntyped
	workers         eosc.IWorkers
}

func (p *PluginManager) CreateRequest(id string, conf map[string]*plugin.Config) plugin.IPlugin {
	return p.createChain(id, conf, pluginRequest)
}

func (p *PluginManager) CreateUpstream(id string, conf map[string]*plugin.Config) plugin.IPlugin {
	return p.createChain(id, conf, pluginUpstream)
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
	list := p.pluginObjs.List()
	// 遍历，全量更新
	for _, obj := range list {
		v, ok := obj.(*PluginObj)
		if !ok {
			continue
		}
		v.IChainHandler.Reset(p.createFilters(v.conf, v.filterType)...)
	}

	return nil
}

func (p *PluginManager) Stop() error {
	return nil
}

func (p *PluginManager) CheckSkill(skill string) bool {
	return false
}

func (p *PluginManager) createFilters(conf map[string]*plugin.Config, filterType string) []http_service.IFilter {
	filters := make([]http_service.IFilter, 0, len(conf))
	plugins := p.plugins
	for _, plg := range plugins {
		if plg.Status == StatusDisable || plg.Status == "" || plg.Type != filterType {
			// 当插件类型不匹配，跳过
			continue
		}
		c := plg.Config
		if v, ok := conf[plg.Name]; ok {
			if v.Disable {
				// 不启用该插件
				continue
			}
			if plg.Status != StatusGlobal && plg.Status != StatusEnable {
				continue
			}
			if v.Config == nil && plg.Status != StatusGlobal {
				continue
			}
			c = v.Config
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
		fi, ok := worker.(http_service.IFilter)
		if !ok {
			log.Error("extender ", plg.ID, " not plg for http-service.Filter")
			continue
		}
		filters = append(filters, fi)
	}
	return filters
}

func (p *PluginManager) createChain(id string, conf map[string]*plugin.Config, filterType string) plugin.IPlugin {
	chain := filter.NewChain(p.createFilters(conf, filterType))

	obj, has := p.pluginObjs.Del(fmt.Sprintf("%s:%s", id, filterType))
	if has {
		o := obj.(*PluginObj)
		o.Destroy()
	}
	obj = NewPluginObj(chain, id, filterType, conf, p.pluginObjs)

	return obj.(*PluginObj)
}

func (p *PluginManager) check(conf interface{}) (Plugins, error) {
	cfg, ok := conf.(*PluginWorkerConfig)
	if !ok {
		return nil, errConfig
	}

	plugins := make(Plugins, 0, len(cfg.Plugins))
	for _, cf := range cfg.Plugins {
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
