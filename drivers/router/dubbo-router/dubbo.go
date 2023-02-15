package dubbo_router

import (
	"github.com/eolinker/apinto/drivers/router"
	"net"
)

func init() {
	router.Register(router.Dubbo, func(port int, listener net.Listener) {
		//todo start dubbo server
	})
}
