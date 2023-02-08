package http_context

import (
	"errors"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

var _ http_context.IDubboContext = (*DubboContext)(nil)

type DubboContext struct {
	*HttpContext
	methodName  string
	serviceName string
	attachment  map[string]interface{}
}

func NewDubboContext(ctx http_context.IHttpContext) (*DubboContext, error) {
	httpCtx, ok := ctx.(*HttpContext)
	if !ok {
		return nil, errors.New("unsupported context type")
	}
	return &DubboContext{
		HttpContext: httpCtx,
		methodName:  "",
		serviceName: "",
		attachment:  make(map[string]interface{}),
	}, nil
}

func (d *DubboContext) MethodName() string {
	return d.methodName
}

func (d *DubboContext) Interface() string {
	return d.serviceName
}

func (d *DubboContext) Serialization() string {
	//TODO implement me
	panic("implement me")
}

func (d *DubboContext) SetAttachment(key string, value interface{}) {
	d.attachment[key] = value
}

func (d *DubboContext) Attachments() map[string]interface{} {
	return d.attachment
}

func (d *DubboContext) Attachment(s string) (interface{}, bool) {
	v, ok := d.attachment[s]
	return v, ok
}
