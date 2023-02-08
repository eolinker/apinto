package manager

import (
	"net"

	"github.com/eolinker/apinto/drivers/router"
	"github.com/eolinker/apinto/plugin"
	"github.com/eolinker/eosc/common/bean"
	"github.com/eolinker/eosc/eocontext"
	"github.com/eolinker/eosc/log"
	"github.com/valyala/fasthttp"
)

var (
	chainProxy eocontext.IChainPro
)

func init() {

	var routerManager = NewManager()
	serverHandler := func(port int, ln net.Listener) {
		server := fasthttp.Server{
			StreamRequestBody:            true,
			DisablePreParseMultipartForm: true,
			MaxRequestBodySize:           100 * 1024 * 1024,
			Handler: func(ctx *fasthttp.RequestCtx) {
				routerManager.FastHandler(port, ctx)
			}}
		server.Serve(ln)
	}
	router.Register(router.Http, serverHandler)

	var pluginManager plugin.IPluginManager
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
