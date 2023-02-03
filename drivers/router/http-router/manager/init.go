package manager

import (
	"github.com/eolinker/apinto/plugin"
	"github.com/eolinker/eosc/common/bean"
	"github.com/eolinker/eosc/eocontext"
	"github.com/eolinker/eosc/log"
)

var (
	chainProxy eocontext.IChainPro
)

func init() {

	var pluginManager plugin.IPluginManager
	var routerManager = NewManager()

	bean.Autowired(&pluginManager)
	log.Debug("new router driver: ")

	var m IManger = routerManager
	bean.Injection(&m)
	bean.AddInitializingBeanFunc(func() {
		log.Debug("init router manager")
		chainProxy = pluginManager.CreateRequest("global", map[string]*plugin.Config{})
		routerManager.SetGlobalFilters(&chainProxy)

	})
}
