package plugin_manager

import (
	"fmt"

	"github.com/eolinker/eosc"
)

type Plugins []*Plugin

type Plugin struct {
	*PluginConfig
	drive eosc.IExtenderDriver
}

func (p *PluginManager) newPlugin(conf *PluginConfig) (*Plugin, error) {

	d, err := p.getExtenderDriver(conf)
	if err != nil {
		return nil, err
	}

	return &Plugin{
		PluginConfig: conf,
		drive:        d,
	}, nil

}

func (p *PluginManager) getExtenderDriver(config *PluginConfig) (eosc.IExtenderDriver, error) {

	driverFactory, has := p.extenderDrivers.GetDriver(config.ID)
	if !has {
		return nil, fmt.Errorf("id:%w", ErrorDriverNotExit)
	}
	return driverFactory.Create(p.name, config.Name, config.Name, config.Type, nil)
}
