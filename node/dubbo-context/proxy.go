package dubbo_context

import dubbo_context "github.com/eolinker/eosc/eocontext/dubbo-context"

var _ dubbo_context.IProxy = (*Proxy)(nil)

type Proxy struct {
	HeaderWriter  dubbo_context.IHeaderWriter
	serviceWriter dubbo_context.IServiceWriter
	body          interface{}
}

func (p *Proxy) Header() dubbo_context.IHeaderWriter {
	return p.HeaderWriter
}

func (p *Proxy) Service() dubbo_context.IServiceWriter {
	return p.serviceWriter
}

func (p *Proxy) SetBody(body interface{}) {
	p.body = body
}
