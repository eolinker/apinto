package router

import (
	"crypto/tls"
	"github.com/eolinker/apinto/certs"
	"github.com/eolinker/apinto/drivers/router/http-router/manager"
	"github.com/eolinker/eosc/common/bean"
	"github.com/eolinker/eosc/config"
	"github.com/eolinker/eosc/log"
	"github.com/eolinker/eosc/traffic"
	"github.com/eolinker/eosc/traffic/mixl"
	"github.com/valyala/fasthttp"
	"net"
	"strconv"
	"strings"
	"sync"
)

var (
	httpRouterManger manager.IManger
)

func init() {
	var tf traffic.ITraffic
	var listenCfg *config.ListenUrl
	bean.Autowired(&httpRouterManger, &tf, &listenCfg)

	bean.AddInitializingBeanFunc(func() {
		initListener(tf, listenCfg)
	})
}
func initListener(tf traffic.ITraffic, listenCfg *config.ListenUrl) {

	if tf.IsStop() {
		return
	}

	m, ok := httpRouterManger.(*manager.Manager)
	if !ok {
		return
	}
	wg := sync.WaitGroup{}
	tcp, ssl := tf.Listen(listenCfg.ListenUrls...)

	listenerByPort := make(map[int][]net.Listener)
	for _, l := range tcp {
		port := readPort(l.Addr())
		listenerByPort[port] = append(listenerByPort[port], l)
	}
	if len(ssl) > 0 {
		tlsConfig := &tls.Config{GetCertificate: certs.GetCertificateFunc()}

		for _, l := range ssl {
			port := readPort(l.Addr())
			listenerByPort[port] = append(listenerByPort[port], tls.NewListener(l, tlsConfig))
		}
	}
	for port, lns := range listenerByPort {

		var ln net.Listener = mixl.NewMixListener(port, lns...)

		wg.Add(1)
		go func(ln net.Listener, port int) {
			log.Debug("fast server:", port, ln.Addr())
			wg.Done()
			server := fasthttp.Server{
				StreamRequestBody:            true,
				DisablePreParseMultipartForm: true,
				MaxRequestBodySize:           100 * 1024 * 1024,
				Handler: func(ctx *fasthttp.RequestCtx) {
					m.FastHandler(port, ctx)
				}}
			server.Serve(ln)
		}(ln, port)
	}
	wg.Wait()
	return
}
func readPort(addr net.Addr) int {
	ipPort := addr.String()
	i := strings.LastIndex(ipPort, ":")
	port := ipPort[i+1:]
	pv, _ := strconv.Atoi(port)
	return pv
}
