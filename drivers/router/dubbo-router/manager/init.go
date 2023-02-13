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
)

func init() {
	manager := NewManager()

	serverHandler := func(port int, listener net.Listener) {

		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Errorf("dubbo-manger listener.Accept err=%v", err)
			}
			go manager.FastHandler(port, conn)
		}

	}
	router.Register(router.Dubbo, serverHandler)

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
