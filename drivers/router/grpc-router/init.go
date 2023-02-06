package grpc_router

import (
	"github.com/eolinker/apinto/drivers/router"
	"net"
)

func init() {
	router.Register(router.GRPC, func(port int, listener net.Listener) {
		// star grpc server
	})
}
