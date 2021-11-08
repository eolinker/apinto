package plugin_manager

import (
	"context"
	"errors"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/http"
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
	CreateRouter(id string, conf map[string]interface{}) http.IChain
	CreateService(id string, conf map[string]interface{}) http.IChain
	CreateUpstream(id string, conf map[string]interface{}) http.IChain
}

type PluginManager struct {
	factories eosc.IUntyped
	pluginCfg []*Plugin
}

func (p *PluginManager) CreateRouter(id string, conf map[string]interface{}) http.IChain {
	panic("implement me")
}

func (p *PluginManager) CreateService(id string, conf map[string]interface{}) http.IChain {
	panic("implement me")
}

func (p *PluginManager) CreateUpstream(id string, conf map[string]interface{}) http.IChain {
	panic("implement me")
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
		pluginCfg: make([]*Plugin, 0),
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, "abc", "123")
	ctx.Value("abc")
}

func (p *PluginManager) Create(id string, conf map[string]interface{}) http.IChain {
	filters := make([]http.IFilter, 0, len(conf))
	for _, cfg := range p.pluginCfg {
		c := cfg.Config
		if v, ok := conf[cfg.ID]; ok {
			c = v
		}
		f, has := p.factories.Get(cfg.ID)
		if !has {
			log.Warn("")
			continue
		}
		factory, ok := f.(IPluginFactory)
		if !ok {
			log.Warn("")
			continue
		}
		filter, err := factory.Create(c)
		if err != nil {
			log.Error("")
			continue
		}
		filters = append(filters, filter)
	}
	return filter.CreateChain(filters)
}
