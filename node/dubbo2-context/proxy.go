package dubbo2_context

import dubbo2_context "github.com/eolinker/eosc/eocontext/dubbo2-context"

var _ dubbo2_context.IProxy = (*Proxy)(nil)

type Proxy struct {
	HeaderWriter  dubbo2_context.IHeaderWriter
	serviceWriter dubbo2_context.IServiceWriter
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

func (p *Proxy) Header() dubbo2_context.IHeaderWriter {
	return p.HeaderWriter
}

func (p *Proxy) Service() dubbo2_context.IServiceWriter {
	return p.serviceWriter
}
