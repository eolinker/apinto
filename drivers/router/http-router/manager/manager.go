package manager

import (
	"crypto/tls"
	"errors"
	"github.com/eolinker/apinto/certs"
	"net"
	"sync"

	http_complete "github.com/eolinker/apinto/drivers/router/http-router/http-complete"
	http_context "github.com/eolinker/apinto/node/http-context"
	http_router "github.com/eolinker/apinto/router/http-router"
	"github.com/eolinker/eosc/config"
	eoscContext "github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
	"github.com/eolinker/eosc/traffic"
	"github.com/valyala/fasthttp"
)

var _ IManger = (*Manager)(nil)
var notFound = new(HttpNotFoundHandler)
var completeCaller = http_complete.NewHttpCompleteCaller()

type IManger interface {
	Set(id string, port int, hosts []string, method []string, path string, append []AppendRule, router http_router.IRouterHandler) error
	Delete(id string)
}

type Manager struct {
	lock    sync.RWMutex
	matcher http_router.IMatcher

	routersData   IRouterData
	globalFilters eoscContext.IChainPro
}

func (m *Manager) Set(id string, port int, hosts []string, method []string, path string, append []AppendRule, router http_router.IRouterHandler) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	routersData := m.routersData.Set(id, port, hosts, method, path, append, router)
	matchers, err := routersData.Parse()
	if err != nil {
		return err
	}
	m.matcher = matchers
	m.routersData = routersData
	return nil
}

func (m *Manager) Delete(id string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	routersData := m.routersData.Delete(id)
	matchers, err := routersData.Parse()
	if err != nil {
		log.Errorf("delete router:%s %s", id, err.Error())
		return
	}
	m.matcher = matchers
	m.routersData = routersData
	return
}

var errNoCertificates = errors.New("tls: no certificates configured")

// NewManager 创建路由管理器
func NewManager(tf traffic.ITraffic, listenCfg *config.ListensMsg, globalFilters eoscContext.IChainPro) *Manager {
	log.Debug("new router manager")
	m := &Manager{
		globalFilters: globalFilters,
		routersData:   new(RouterData),
	}

	if tf.IsStop() {
		return m
	}

	wg := sync.WaitGroup{}

	for _, cfg := range listenCfg.Listens {
		port := int(cfg.Port)
		var ln net.Listener
		log.Debug("read listener:", cfg.Scheme, ":", port)
		if cfg.Scheme == "https" {
			ln = tf.ListenTcp(port, traffic.Https)
			if ln == nil {
				continue
			}

			iCert := certs.NewCert(cfg.Certificate, listenCfg.Dir)
			ln = tls.NewListener(ln, &tls.Config{GetCertificate: iCert.GetCertificate})

			//cert, err := config.NewCert(cfg.Certificate, listenCfg.Dir)
			//if err == nil {
			//	ln = tls.NewListener(ln, &tls.Config{GetCertificate: cert.GetCertificate})
			//} else {
			//
			//	//ln = tls.NewListener(ln, &tls.Config{GetCertificate: certs.GetCertificate})
			//
			//	//ln = tls.NewListener(ln, &tls.Config{GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			//	//	return nil, errNoCertificates
			//	//}})
			//	log.Warn("worker create certificate error:", err)
			//}
		} else {
			ln = tf.ListenTcp(port, traffic.Http1)
			if ln == nil {
				continue
			}
		}

		wg.Add(1)
		go func(ln net.Listener, port int) {
			log.Debug("fast server:", port, ln.Addr())
			wg.Done()
			server := fasthttp.Server{DisablePreParseMultipartForm: true, Handler: func(ctx *fasthttp.RequestCtx) {
				m.FastHandler(port, ctx)
			}}
			server.Serve(ln)
		}(ln, port)
	}
	wg.Wait()
	return m
}
func (m *Manager) FastHandler(port int, ctx *fasthttp.RequestCtx) {
	httpContext := http_context.NewContext(ctx, port)
	r, has := m.matcher.Match(port, httpContext.Request())
	if !has {
		httpContext.SetFinish(notFound)
		httpContext.SetCompleteHandler(notFound)
		m.globalFilters.Chain(httpContext, completeCaller)
	} else {
		log.Debug("match has:", port)
		r.ServeHTTP(httpContext)
	}
	//}

}

type HttpNotFoundHandler struct {
}

func (m *HttpNotFoundHandler) Complete(ctx eoscContext.EoContext) error {

	httpContext, err := http_service.Assert(ctx)
	if err != nil {
		return nil
	}
	httpContext.Response().SetStatus(404, "404")
	httpContext.Response().SetBody([]byte("404 Not Found"))
	return nil
}

func (m *HttpNotFoundHandler) Finish(ctx eoscContext.EoContext) error {
	httpContext, err := http_service.Assert(ctx)
	if err != nil {
		return err
	}
	httpContext.FastFinish()
	return nil
}
