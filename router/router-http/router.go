package router_http

import (
	"sync"

	"github.com/eolinker/eosc/log"

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
	Handler(ctx *fasthttp.RequestCtx)
}

//Router 实现了路由树接口
type Router struct {
	locker  sync.Locker
	data    eosc.IUntyped
	match   IMatcher
	handler fasthttp.RequestHandler
	//chain   []IRouterFilter
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

//Handler 路由树的handler方法
func (r *Router) Handler(requestCtx *fasthttp.RequestCtx) {
	match := r.match
	if r.match == nil {
		requestCtx.NotFound()
		return
	}
	log.Debug("router handler", requestCtx.Request.String())
	ctx := http_context.NewContext(requestCtx)
	// TODO: 执行全局的Filter
	h, e, has := match.Match(ctx.Request())
	if !has {
		requestCtx.NotFound()
		return
	}
	h.Handle(ctx, NewEndPoint(e))

}

//SetRouter 将路由配置加入到路由树中
func (r *Router) SetRouter(id string, config *Config) error {
	r.locker.Lock()
	defer r.locker.Unlock()
	data := r.data.Clone()
	data.Set(id, config)
	//重新生成路由树
	m, err := parseData(data)
	if err != nil {
		return err
	}

	r.match = m
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
		r.data = data
		r.match = nil
	} else {
		//重新生成路由树
		m, err := parseData(data)
		if err != nil {
			// 路由树生成失败， 则放弃
			return r.data.Count()
		}
		// 路由树生成成功，则替换
		r.data = data
		r.match = m
	}

	return r.data.Count()
}

func parseData(data eosc.IUntyped) (IMatcher, error) {
	list := data.List()
	cs := make([]*Config, 0, len(list))
	for _, i := range list {
		cs = append(cs, i.(*Config))
	}
	return parse(cs)
}
