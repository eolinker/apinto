package router_http_manager

import (
	"github.com/eolinker/apinto/plugin"
	"github.com/eolinker/eosc/common/bean"
	"github.com/eolinker/eosc/config"
	"github.com/eolinker/eosc/log"
	"github.com/eolinker/eosc/traffic"
)

func init() {
	var tf traffic.ITraffic
	var cfg *config.ListensMsg
	var pluginManager plugin.IPluginManager

	bean.Autowired(&tf)
	bean.Autowired(&cfg)
	bean.Autowired(&pluginManager)

	bean.AddInitializingBeanFunc(func() {
		log.Debug("init router manager")

		manager := NewManager(tf, cfg, pluginManager.CreateRequest("global", map[string]*plugin.Config{}))
		var m IManger = manager
		bean.Injection(&m)
	})
}
