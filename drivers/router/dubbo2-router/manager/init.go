package manager

import (
	"github.com/eolinker/apinto/drivers/router"
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

		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Errorf("dubbo-manger listener.Accept err=%v", err)
			}
			go manager.connHandler.Handler(port, conn)
		}

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
