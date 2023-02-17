package grpc_proxy_rewrite

import (
	"strings"

	grpc_context "github.com/eolinker/eosc/eocontext/grpc-context"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
)

var _ eocontext.IFilter = (*ProxyRewrite)(nil)
var _ grpc_context.GrpcFilter = (*ProxyRewrite)(nil)

var (
	regexpErrInfo   = `[plugin proxy-rewrite2 config err] Compile regexp fail. err regexp: %s `
	notMatchErrInfo = `[plugin proxy-rewrite2 err] Proxy path rewrite fail. Request path can't match any rewrite-path. request path: %s `
)

type ProxyRewrite struct {
	drivers.WorkerBase
	service   string
	method    string
	headers   map[string]string
	authority string
}

func (p *ProxyRewrite) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return grpc_context.DoGrpcFilter(p, ctx, next)
}

func (p *ProxyRewrite) DoGrpcFilter(ctx grpc_context.IGrpcContext, next eocontext.IChain) (err error) {

	if p.service != "" {
		ctx.Proxy().SetService(p.service)
	}
	if p.method != "" {
		ctx.Proxy().SetMethod(p.method)
	}
	if p.authority != "" {
		ctx.Proxy().SetHost(p.authority)
	}
	for key, value := range p.headers {
		ctx.Proxy().Headers().Set(key, value)
	}

	if next != nil {
		return next.DoChain(ctx)
	}
	return nil
}

func (p *ProxyRewrite) Start() error {
	return nil
}

func (p *ProxyRewrite) Reset(v interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	conf, err := check(v)
	if err != nil {
		return err
	}
	p.authority = strings.TrimSpace(conf.Authority)
	p.headers = conf.Headers
	p.service = conf.Service
	p.method = conf.Method
	return nil
}

func (p *ProxyRewrite) Stop() error {
	return nil
}

func (p *ProxyRewrite) Destroy() {

	p.headers = nil
}

func (p *ProxyRewrite) CheckSkill(skill string) bool {
	return grpc_context.FilterSkillName == skill
}
