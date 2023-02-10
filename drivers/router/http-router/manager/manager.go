package manager

import (
	"sync"
	"sync/atomic"

	http_complete "github.com/eolinker/apinto/drivers/router/http-router/http-complete"
	http_context "github.com/eolinker/apinto/node/http-context"
	"github.com/eolinker/apinto/router"
	eoscContext "github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
	"github.com/valyala/fasthttp"
)

var _ IManger = (*Manager)(nil)
var notFound = new(HttpNotFoundHandler)
var completeCaller = http_complete.NewHttpCompleteCaller()

type IManger interface {
	Set(id string, port int, hosts []string, method []string, path string, append []AppendRule, router router.IRouterHandler) error
	Delete(id string)
}
type Manager struct {
	lock    sync.RWMutex
	matcher router.IMatcher

	routersData   IRouterData
	globalFilters atomic.Pointer[eoscContext.IChainPro]
}

func (m *Manager) SetGlobalFilters(globalFilters *eoscContext.IChainPro) {
	m.globalFilters.Store(globalFilters)
}

// NewManager 创建路由管理器
func NewManager() *Manager {
	return &Manager{routersData: new(RouterData)}
}

func (m *Manager) Set(id string, port int, hosts []string, method []string, path string, append []AppendRule, router router.IRouterHandler) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	routersData := m.routersData.Set(id, port, hosts, method, path, append, router)
	matchers, err := routersData.Parse()
	if err != nil {
		log.Error("parse router data error: ", err)
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

func (m *Manager) FastHandler(port int, ctx *fasthttp.RequestCtx) {

	httpContext := http_context.NewContext(ctx, port)
	if m.matcher == nil {
		httpContext.SetFinish(notFound)
		httpContext.SetCompleteHandler(notFound)
		globalFilters := m.globalFilters.Load()
		if globalFilters != nil {
			(*globalFilters).Chain(httpContext, completeCaller)
		}
		return
	}
	log.Debug("port is ", port, " request: ", httpContext.Request())
	r, has := m.matcher.Match(port, httpContext.Request())
	if !has {
		httpContext.SetFinish(notFound)
		httpContext.SetCompleteHandler(notFound)
		globalFilters := m.globalFilters.Load()
		if globalFilters != nil {
			(*globalFilters).Chain(httpContext, completeCaller)

		}
	} else {
		log.Debug("match has:", port)
		r.ServeHTTP(httpContext)
	}
	finishHandler := httpContext.GetFinish()
	if finishHandler != nil {
		finishHandler.Finish(httpContext)
	}
}

type NotFoundHandler struct {
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
