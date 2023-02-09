package manager

import (
	"net"

	"google.golang.org/grpc"

	"github.com/eolinker/eosc/eocontext"

	"github.com/eolinker/apinto/drivers/router"
	"github.com/eolinker/apinto/plugin"
	"github.com/eolinker/eosc/common/bean"
	"github.com/eolinker/eosc/log"
)

var (
	chainProxy eocontext.IChainPro
)

func init() {

	var routerManager = NewManager()
	serverHandler := func(port int, ln net.Listener) {
		opts := []grpc.ServerOption{
			grpc.UnknownServiceHandler(func(srv interface{}, stream grpc.ServerStream) error {
				routerManager.FastHandler(port, srv, stream)
				return nil
			}),
		}
		server := grpc.NewServer(opts...)
		server.Serve(ln)
	}
	router.Register(router.GRPC, serverHandler)

	var pluginManager plugin.IPluginManager
	bean.Autowired(&pluginManager)
	log.Debug("new router driver: ")

	var m IManger = routerManager
	bean.Injection(&m)
	bean.AddInitializingBeanFunc(func() {
		log.Debug("init grpc router manager")
		chainProxy = pluginManager.CreateRequest("global", map[string]*plugin.Config{})
		routerManager.SetGlobalFilters(&chainProxy)
	})

}
