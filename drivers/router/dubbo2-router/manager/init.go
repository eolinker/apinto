package manager

import (
	"github.com/eolinker/apinto/drivers/router"
	getty "github.com/eolinker/apinto/dubbo-getty/server"
	"github.com/eolinker/apinto/plugin"
	"github.com/eolinker/eosc/common/bean"
	"github.com/eolinker/eosc/eocontext"
	"github.com/eolinker/eosc/log"
	"net"
)

var (
	chainProxy eocontext.IChainPro
	manager    = NewManager()
)

func init() {

	serverHandler := func(port int, listener net.Listener) {

		server := getty.NewServer(manager.Handler, getty.WithListenerServer(listener))
		server.Start()
	}
	router.Register(router.Dubbo2, serverHandler)

	var pluginManager plugin.IPluginManager
	bean.Autowired(&pluginManager)
	log.Debug("new router driver: ")

	var m IManger = manager
	bean.Injection(&m)

	bean.AddInitializingBeanFunc(func() {
		log.Debug("init router manager")
		chainProxy = pluginManager.Global()
		manager.SetGlobalFilters(&chainProxy)
	})
}
