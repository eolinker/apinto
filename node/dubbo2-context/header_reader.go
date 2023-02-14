package dubbo2_context

import (
	dubbo2_context "github.com/eolinker/eosc/eocontext/dubbo2-context"
)

var _ dubbo2_context.IHeaderReader = (*RequestHeaderReader)(nil)

type RequestHeaderReader struct {
	id             int64
	serialID       byte
	_type          int
	bodyLen        int
	responseStatus byte
}

func (r *RequestHeaderReader) ID() int64 {
	return r.id
}

func (r *RequestHeaderReader) SerialID() byte {
	return r.serialID
}

func (r *RequestHeaderReader) Type() int {
	return r._type
}

func (r *RequestHeaderReader) BodyLen() int {
	return r.bodyLen
}

func (r *RequestHeaderReader) ResponseStatus() byte {
	return r.responseStatus
}
