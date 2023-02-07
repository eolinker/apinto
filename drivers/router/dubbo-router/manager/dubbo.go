package manager

import (
	"github.com/eolinker/apinto/drivers/router"
	"github.com/eolinker/eosc/common/bean"
	"github.com/eolinker/eosc/log"
	"net"
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

	bean.Injection(&manager)
}
