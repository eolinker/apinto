package plugin_manager

import (
	"errors"

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
	CreateRouter(id string, conf map[string]*OrdinaryPlugin) http.IChain
	CreateService(id string, conf map[string]*OrdinaryPlugin) http.IChain
	CreateUpstream(id string, conf map[string]*OrdinaryPlugin) http.IChain
}

type PluginManager struct {
	factories eosc.IUntyped
	pluginCfg []*GlobalPlugin
	pluginObj eosc.IUntyped
}

type PluginObj struct {
	http.IChain
	id   string
	conf map[string]*OrdinaryPlugin
}

func (p *PluginManager) create(id string, conf map[string]*OrdinaryPlugin, t string) http.IChain {
	filters := make([]http.IFilter, 0, len(conf))
	for _, cfg := range p.pluginCfg {
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
		} else if cfg.Status != StatusGlobal {
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
			log.Error("plugin manager: fail to create filter,error is ", err)
			continue
		}
		filters = append(filters, filter)
	}
	return filter.CreateChain(filters)
}

func (p *PluginManager) CreateRouter(id string, conf map[string]*OrdinaryPlugin) http.IChain {
	return p.create(id, conf, pluginRouter)
}

func (p *PluginManager) CreateService(id string, conf map[string]*OrdinaryPlugin) http.IChain {
	return p.create(id, conf, pluginService)
}

func (p *PluginManager) CreateUpstream(id string, conf map[string]*OrdinaryPlugin) http.IChain {
	return p.create(id, conf, pluginUpstream)
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
	p.pluginCfg = cfg
	// TODO: 此处
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
		factories: eosc.NewUntyped(),
		pluginCfg: make([]*GlobalPlugin, 0),
	}
}
