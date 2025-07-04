package router

import (
	"crypto/tls"
	"net"
	"net/http"
	"sync"

	"github.com/tjfoc/gmsm/gmtls"

	"github.com/eolinker/apinto/certs"
	"github.com/eolinker/eosc/common/bean"
	"github.com/eolinker/eosc/config"
	"github.com/eolinker/eosc/log"
	"github.com/eolinker/eosc/traffic"
	"github.com/eolinker/eosc/traffic/mixl"
	"github.com/soheilhy/cmux"
)

func init() {
	matchWriters[AnyTCP] = matchersToMatchWriters(cmux.Any())

	// 基本TLS
	matchWriters[TlsTCP] = matchersToMatchWriters(cmux.TLS(
		gmtls.VersionSSL30,
		gmtls.VersionTLS10,
		gmtls.VersionTLS11,
		gmtls.VersionTLS12,
		tls.VersionTLS13,
		gmtls.VersionGMSSL,
	))

	matchWriters[Http] = matchersToMatchWriters(cmux.HTTP1Fast(http.MethodPatch))
	matchWriters[Dubbo2] = matchersToMatchWriters(cmux.PrefixMatcher(string([]byte{0xda, 0xbb})))
	matchWriters[GRPC] = []cmux.MatchWriter{cmux.HTTP2MatchHeaderFieldPrefixSendSettings("content-type", "application/grpc")}
	var tf traffic.ITraffic
	var listenCfg *config.ListenUrl
	bean.Autowired(&tf, &listenCfg)

	bean.AddInitializingBeanFunc(func() {
		initListener(tf, listenCfg)
	})
}

func initListener(tf traffic.ITraffic, listenCfg *config.ListenUrl) {

	if tf.IsStop() {
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
		support := gmtls.NewGMSupport()
		support.EnableMixMode()
		gmTlsConfig := &gmtls.Config{
			GetCertificate:   certs.GetAutoCertificateFunc(),
			GetKECertificate: certs.GetKECertificate(),
			GMSupport:        support,
			MinVersion:       gmtls.VersionGMSSL,
			MaxVersion:       tls.VersionTLS13,
		}

		for _, l := range ssl {
			log.Debug("ssl listen: ", l.Addr().String())
			port := readPort(l.Addr())
			listenerByPort[port] = append(listenerByPort[port], gmtls.NewListener(l, gmTlsConfig))
		}
	}
	for port, lns := range listenerByPort {

		var ln = mixl.NewMixListener(port, lns...)

		wg.Add(1)
		go func(ln net.Listener, p int) {
			wg.Done()
			m := cmux.New(ln)
			for i, handler := range handlers {
				log.Debug("i is ", i, " handler is ", handler)
				if handler != nil {
					go handler(p, m.MatchWithWriters(matchWriters[i]...))
				}
			}

			m.Serve()
		}(ln, port)

	}
	wg.Wait()
	return
}
