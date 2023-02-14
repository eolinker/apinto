package router

import (
	"crypto/tls"
	"errors"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/certs"
	"github.com/eolinker/eosc/common/bean"
	"github.com/eolinker/eosc/config"
	"github.com/eolinker/eosc/traffic"
	"github.com/eolinker/eosc/traffic/mixl"
	"github.com/soheilhy/cmux"
)

type RouterType int

const (
	GRPC RouterType = iota
	Http
	Dubbo
	TslTCP
	AnyTCP
	depth
)

var (
	handlers = make([]RouterServerHandler, depth)
	//matchers                 = make([][]cmux.Matcher, depth)
	matchWriters             = make([][]cmux.MatchWriter, depth)
	ErrorDuplicateRouterType = errors.New("duplicate")
)

func Register(tp RouterType, handler RouterServerHandler) error {
	if handlers[tp] != nil {
		return ErrorDuplicateRouterType
	}
	handlers[tp] = handler
	return nil
}

type RouterServerHandler func(port int, listener net.Listener)

func init() {
	matchWriters[AnyTCP] = matchersToMatchWriters([]cmux.Matcher{cmux.Any()})
	matchWriters[TslTCP] = matchersToMatchWriters([]cmux.Matcher{cmux.TLS()})
	matchWriters[Http] = matchersToMatchWriters([]cmux.Matcher{cmux.HTTP1Fast()})
	matchWriters[Dubbo] = matchersToMatchWriters([]cmux.Matcher{cmux.PrefixMatcher(string([]byte{0xda, 0xbb}))})
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
		tlsConfig := &tls.Config{GetCertificate: certs.GetCertificateFunc()}

		for _, l := range ssl {
			port := readPort(l.Addr())
			listenerByPort[port] = append(listenerByPort[port], tls.NewListener(l, tlsConfig))
		}
	}
	for port, lns := range listenerByPort {

		var ln net.Listener = mixl.NewMixListener(port, lns...)

		wg.Add(1)
		go func(ln net.Listener, p int) {
			wg.Done()
			cMux := cmux.New(ln)
			for i, handler := range handlers {
				log.Debug("i is ", i, " handler is ", handler)
				if handler != nil {
					go handler(p, cMux.MatchWithWriters(matchWriters[i]...))
				}
			}

			cMux.Serve()
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

func matchersToMatchWriters(matchers []cmux.Matcher) []cmux.MatchWriter {
	mws := make([]cmux.MatchWriter, 0, len(matchers))
	for _, m := range matchers {
		cm := m
		mws = append(mws, func(w io.Writer, r io.Reader) bool {
			return cm(r)
		})
	}
	return mws
}
