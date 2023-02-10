package dubbo_context

import dubbo_context "github.com/eolinker/eosc/eocontext/dubbo-context"

var _ dubbo_context.IProxy = (*Proxy)(nil)

type Proxy struct {
	HeaderWriter  dubbo_context.IHeaderWriter
	serviceWriter dubbo_context.IServiceWriter
	param         interface{}
	attachments   map[string]interface{}
}

func (p *Proxy) GetParam() interface{} {
	return p.param
}

func (p *Proxy) SetParam(param interface{}) {
	p.param = param
}

func (p *Proxy) Attachments() map[string]interface{} {
	return p.attachments
}

func (p *Proxy) SetAttachment(k string, v interface{}) {
	p.attachments[k] = v
}

func (p *Proxy) Header() dubbo_context.IHeaderWriter {
	return p.HeaderWriter
}

func (p *Proxy) Service() dubbo_context.IServiceWriter {
	return p.serviceWriter
}
