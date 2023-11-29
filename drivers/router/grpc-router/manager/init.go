package manager

import (
	"io"
	"net"

	"google.golang.org/grpc"

	"github.com/eolinker/eosc/eocontext"

	"github.com/eolinker/eosc/common/bean"
	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/drivers/router"
	"github.com/eolinker/apinto/plugin"
)

var (
	chainProxy eocontext.IChainPro
)

func init() {

	var routerManager = NewManager()
	serverHandler := func(port int, ln net.Listener) {
		opts := []grpc.ServerOption{
			grpc.UnknownServiceHandler(func(srv interface{}, stream grpc.ServerStream) error {
				err := routerManager.FastHandler(port, srv, stream)
				if err == io.EOF {
					return nil
				}
				return err
			}),
			grpc.MaxRecvMsgSize(64 * 1024 * 1024),
			grpc.MaxSendMsgSize(64 * 1024 * 1024),
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
		chainProxy = pluginManager.Global()
		routerManager.SetGlobalFilters(&chainProxy)
	})

}
