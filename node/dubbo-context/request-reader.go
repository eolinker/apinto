package dubbo_context

import (
	dubbo_context "github.com/eolinker/eosc/eocontext/dubbo-context"
)

var _ dubbo_context.IRequestReader = (*RequestReader)(nil)

type RequestReader struct {
	headerReader  dubbo_context.IHeaderReader
	serviceReader dubbo_context.IServiceReader
	body          interface{}
	attachments   map[string]interface{}
}

func (r *RequestReader) Attachments() map[string]interface{} {
	return r.attachments
}

func (r *RequestReader) Attachment(s string) (interface{}, bool) {
	v, ok := r.attachments[s]
	return v, ok
}

func (r *RequestReader) Header() dubbo_context.IHeaderReader {
	return r.headerReader
}

func (r *RequestReader) Service() dubbo_context.IServiceReader {
	return r.serviceReader
}

func (r *RequestReader) Body() interface{} {
	return r.body
}
