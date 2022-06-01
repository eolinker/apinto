package plugin_manager

import (
	"fmt"

	"github.com/eolinker/eosc/log"

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
	if conf.Status == StatusGlobal && conf.Config == nil {
		return nil, ErrorGlobalPluginMastConfig
	}
	if conf.Status == StatusGlobal {
		v, err := toConfig(conf.Config, d.ConfigType())
		if err != nil {
			log.Info("global plugin:", conf.Name, "config:", err)
			return nil, fmt.Errorf("%s:%w", conf.Name, ErrorGlobalPluginConfigInvalid)
		}
		if dc, ok := d.(eosc.IExtenderConfigChecker); ok {
			errCheck := dc.Check(v, nil)
			if errCheck != nil {
				return nil, errCheck
			}
		}

	}

	return &Plugin{
		PluginConfig: conf,
		drive:        d,
	}, nil

}

func (p *PluginManager) getExtenderDriver(config *PluginConfig) (eosc.IExtenderDriver, error) {
	log.DebugF("getExtenderDriver:%p.get(%v)", p, config)
	driverFactory, has := p.extenderDrivers.GetDriver(config.ID)
	if !has {
		return nil, fmt.Errorf("id:%w", ErrorDriverNotExit)
	}
	return driverFactory.Create(p.id, config.Name, config.Name, config.Type, config.InitConfig)
}
