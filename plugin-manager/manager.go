package plugin_manager

import (
	"errors"
	"fmt"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/http"
	"github.com/eolinker/eosc/log"
	"github.com/eolinker/goku/filter"
)

var id = "plugin@setting"

func RegisterFilter(factory IPluginFactory) {
	if factory == nil {
		return
	}
	manager.factories.Set(factory.Name(), factory)
}

var manager = newPluginManager()

type IPluginManager interface {
	CreateRouter(id string, conf map[string]*OrdinaryPlugin) filter.IChain
	CreateService(id string, conf map[string]*OrdinaryPlugin) filter.IChain
	CreateUpstream(id string, conf map[string]*OrdinaryPlugin) filter.IChain
}

type PluginManager struct {
	factories  eosc.IUntyped
	plugins    []*GlobalPlugin
	pluginObjs eosc.IUntyped
}

type PluginObj struct {
	filter.IChain
	id   string
	t    string
	conf map[string]*OrdinaryPlugin
}

func (p *PluginManager) RemoveObj(id string) (*PluginObj, bool) {
	value, ok := p.pluginObjs.Del(id)
	if !ok {
		return nil, false
	}
	v, ok := value.(*PluginObj)
	return v, ok
}

func (p *PluginManager) createFilters(conf map[string]*OrdinaryPlugin, t string) []http.IFilter {
	filters := make([]http.IFilter, 0, len(conf))
	for _, cfg := range p.plugins {
		if cfg.Status == StatusDisable || cfg.Status == "" || cfg.Type != t {
			// 当插件类型不匹配，跳过
			continue
		}
		c := cfg.Config
		if v, ok := conf[cfg.ID]; ok {
			if v.Disable {
				// 不启用该插件
				continue
			}
			c = v
		} else if cfg.Status != StatusGlobal && cfg.Status != StatusEnable {
			continue
		}

		f, has := p.factories.Get(cfg.ID)
		if !has {
			log.Warn("plugin manager: no plugin factory,id is ", cfg.ID)
			continue
		}
		factory, ok := f.(IPluginFactory)
		if !ok {
			log.Warn("plugin manager: no implement factory interface,id is ", cfg.ID)
			continue
		}
		filter, err := factory.Create(c)
		if err != nil {
			log.Error("plugin manager: fail to createFilters filter,error is ", err)
			continue
		}
		filters = append(filters, filter)
	}
	return filters
}

func (p *PluginManager) new(id string, conf map[string]*OrdinaryPlugin, t string) filter.IChain {
	chain := filter.Create(p.createFilters(conf, t))
	obj := &PluginObj{
		IChain: chain,
		id:     id,
		conf:   conf,
		t:      t,
	}
	p.pluginObjs.Set(fmt.Sprintf("%s:%s", id, t), obj)
	return chain
}

func (p *PluginManager) CreateRouter(id string, conf map[string]*OrdinaryPlugin) filter.IChain {
	return p.new(id, conf, pluginRouter)
}

func (p *PluginManager) CreateService(id string, conf map[string]*OrdinaryPlugin) filter.IChain {
	return p.new(id, conf, pluginService)
}

func (p *PluginManager) CreateUpstream(id string, conf map[string]*OrdinaryPlugin) filter.IChain {
	return p.new(id, conf, pluginUpstream)
}

func (p *PluginManager) Id() string {
	return id
}

func (p *PluginManager) Start() error {
	return nil
}

func (p *PluginManager) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	cfg, ok := conf.(Plugins)
	if !ok {
		return errors.New("")
	}
	p.plugins = cfg
	// 遍历，全量更新
	for _, obj := range p.pluginObjs.All() {
		v, ok := obj.(*PluginObj)
		if !ok {
			continue
		}
		v.IChain.Reset(p.createFilters(v.conf, v.t))
	}

	return nil
}

func (p *PluginManager) Stop() error {
	return nil
}

func (p *PluginManager) CheckSkill(skill string) bool {
	panic("implement me")
}

func newPluginManager() *PluginManager {
	return &PluginManager{
		factories:  eosc.NewUntyped(),
		plugins:    make([]*GlobalPlugin, 0),
		pluginObjs: eosc.NewUntyped(),
	}
}
