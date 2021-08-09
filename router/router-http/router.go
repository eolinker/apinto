package router_http

import (
	"sync"

	http_context "github.com/eolinker/goku-eosc/node/http-context"

	"github.com/valyala/fasthttp"

	"github.com/eolinker/eosc"
)

var _ IRouter = (*Router)(nil)

type IRouter interface {
	SetRouter(id string, config *Config) error
	Count() int
	Del(id string) int
	Handler() fasthttp.RequestHandler
}

type Router struct {
	locker  sync.Locker
	data    eosc.IUntyped
	match   IMatcher
	handler fasthttp.RequestHandler
}

func NewRouter() *Router {
	return &Router{
		locker: &sync.Mutex{},
		data:   eosc.NewUntyped(),
	}
}

func (r *Router) Count() int {
	return r.data.Count()
}

func (r *Router) Handler() fasthttp.RequestHandler {
	return func(requestCtx *fasthttp.RequestCtx) {
		ctx := http_context.NewContext(requestCtx)
		h, e, has := r.match.Match(ctx.Request())
		if !has {
			http_context.NotFound(ctx)
			return
		}
		h.Handle(ctx, NewEndPoint(e))
	}
}

//func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
//	h, e, has := r.match.Match(req)
//	if !has {
//		http.NotFound(w, req)
//		return
//	}
//	h.Handle(w, req, NewEndPoint(e))
//}

func (r *Router) SetRouter(id string, config *Config) error {
	r.locker.Lock()
	defer r.locker.Unlock()
	data := r.data.Clone()
	data.Set(id, config)
	list := data.List()
	cs := make([]*Config, 0, len(list))
	for _, i := range list {
		cs = append(cs, i.(*Config))
	}
	matcher, err := parse(cs)
	if err != nil {
		return err
	}
	r.match = matcher
	r.data = data
	return nil
}

func (r *Router) Del(id string) int {
	r.locker.Lock()
	defer r.locker.Unlock()

	data := r.data.Clone()
	data.Del(id)
	if data.Count() == 0 {
		r.match = nil
	} else {
		list := data.List()
		cs := make([]*Config, 0, len(list))
		for _, i := range list {
			cs = append(cs, i.(*Config))
		}
		m, err := parse(cs)
		if err != nil {
			return r.data.Count()
		}
		r.match = m
	}

	return r.data.Count()
}
