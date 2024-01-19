package manager

import (
	"sync"

	"github.com/eolinker/apinto/router"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
)

type IPreRouterData interface {
	router.IRouterPreHandler
	AddPreRouter(id string, method []string, path string, handler router.IRouterPreHandler)
	DeletePreRouter(id string)
}

type preRouterItem struct {
	id      string
	method  []string
	path    string
	handler router.IRouterPreHandler
}

var (
	_ router.IRouterPreHandler = (*iPreRouterHandler)(nil)
)

type iPreRouterHandler struct {
	routers map[string]map[string][]router.IRouterPreHandler
}

func (i *iPreRouterHandler) Server(ctx eocontext.EoContext) (isContinue bool) {
	if i == nil {
		return true
	}
	httpCtx, err := http_context.Assert(ctx)
	if err != nil {
		return true
	}
	method := httpCtx.Request().Method()
	path := httpCtx.Request().URI().Path()

	ms, has := i.routers[path]

	if !has {
		return true
	}

	handlers, has := ms[method]
	if !has {
		handlers, has = ms["*"]
		if !has {
			return true
		}
	}
	for _, handler := range handlers {
		if !handler.Server(ctx) {
			return false
		}
	}
	return true

}

type imlPreRouterData struct {
	lock sync.RWMutex

	handler router.IRouterPreHandler
	items   map[string]*preRouterItem
}

func (p *imlPreRouterData) Server(ctx eocontext.EoContext) (isContinue bool) {
	if p == nil || p.handler == nil {
		return true
	}
	log.Debug("pre router hander:", p.handler)
	return p.handler.Server(ctx)
}

func newImlPreRouterData() IPreRouterData {
	return &imlPreRouterData{
		items: make(map[string]*preRouterItem),
	}
}

func (p *imlPreRouterData) AddPreRouter(id string, method []string, path string, handler router.IRouterPreHandler) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.items[id] = &preRouterItem{
		id:      id,
		method:  method,
		path:    path,
		handler: handler,
	}
	log.Debug("add pre router:", p.items)
	p.handler = p.parse()
}

func (p *imlPreRouterData) DeletePreRouter(id string) {
	p.lock.Lock()
	defer p.lock.Unlock()
	delete(p.items, id)
	p.handler = p.parse()
}
func (p *imlPreRouterData) parse() router.IRouterPreHandler {
	if len(p.items) == 0 {
		return nil
	}
	routers := make(map[string]map[string][]router.IRouterPreHandler)
	for _, v := range p.items {
		if _, has := routers[v.path]; !has {
			routers[v.path] = make(map[string][]router.IRouterPreHandler)
		}
		if len(v.method) == 0 {
			v.method = []string{"*"}
		}
		for _, method := range v.method {
			routers[v.path][method] = append(routers[v.path][method], v.handler)
		}
	}
	return &iPreRouterHandler{routers: routers}
}
