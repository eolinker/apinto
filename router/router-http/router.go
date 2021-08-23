package router_http

import (
	"sync"

	http_context "github.com/eolinker/goku/node/http-context"

	"github.com/valyala/fasthttp"

	"github.com/eolinker/eosc"
)

var _ IRouter = (*Router)(nil)

//IRouter 路由树的接口
type IRouter interface {
	SetRouter(id string, config *Config) error
	Count() int
	Del(id string) int
	Handler() fasthttp.RequestHandler
}

//Router 实现了路由树接口
type Router struct {
	locker  sync.Locker
	data    eosc.IUntyped
	match   IMatcher
	handler fasthttp.RequestHandler
}

//NewRouter 新建路由树
func NewRouter() *Router {
	return &Router{
		locker: &sync.Mutex{},
		data:   eosc.NewUntyped(),
	}
}

//Count 返回路由树中配置实例的数量
func (r *Router) Count() int {
	return r.data.Count()
}

//Handler
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

//SetRouter 将路由配置加入到路由树中
func (r *Router) SetRouter(id string, config *Config) error {
	r.locker.Lock()
	defer r.locker.Unlock()
	data := r.data.Clone()
	data.Set(id, config)
	//重新生成路由树
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

//Del 将某个路由从路由树中删去
func (r *Router) Del(id string) int {
	r.locker.Lock()
	defer r.locker.Unlock()

	data := r.data.Clone()
	data.Del(id)
	if data.Count() == 0 {
		r.match = nil
	} else {
		//重新生成路由树
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
