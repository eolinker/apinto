package grpc_proxy_rewrite

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	dubbo2_context "github.com/eolinker/eosc/eocontext/dubbo2-context"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var _ eocontext.IFilter = (*ProxyRewrite)(nil)
var _ dubbo2_context.DubboFilter = (*ProxyRewrite)(nil)

var (
	regexpErrInfo   = `[plugin proxy-rewrite2 config err] Compile regexp fail. err regexp: %s `
	notMatchErrInfo = `[plugin proxy-rewrite2 err] Proxy path rewrite fail. Request path can't match any rewrite-path. request path: %s `
)

type ProxyRewrite struct {
	drivers.WorkerBase
	service         string
	method          string
	headers         map[string]string
	tls             bool
	skipCertificate bool
}

func (p *ProxyRewrite) DoDubboFilter(ctx dubbo2_context.IDubbo2Context, next eocontext.IChain) error {
	if p.service != "" {
		ctx.Proxy().Service().SetInterface(p.service)
	}
	if p.method != "" {
		ctx.Proxy().Service().SetMethod(p.method)
	}
	for key, value := range p.headers {
		ctx.Proxy().SetAttachment(key, value)
	}
	if next != nil {
		return next.DoChain(ctx)
	}
	return nil
}

func (p *ProxyRewrite) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return dubbo2_context.DoDubboFilter(p, ctx, next)
}

func (p *ProxyRewrite) Start() error {
	return nil
}

func (p *ProxyRewrite) Reset(v interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	conf, err := check(v)
	if err != nil {
		return err
	}
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
	return http_service.FilterSkillName == skill
}
