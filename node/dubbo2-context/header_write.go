package dubbo2_context

import dubbo2_context "github.com/eolinker/eosc/eocontext/dubbo2-context"

var _ dubbo2_context.IHeaderWriter = (*RequestHeaderWrite)(nil)

type RequestHeaderWrite struct {
	id             int64
	serialID       byte
	_type          int
	bodyLen        int
	responseStatus byte
}

func (r *RequestHeaderWrite) SetID(id int64) {
	r.id = id
}

func (r *RequestHeaderWrite) SetSerialID(serialID byte) {
	r.serialID = serialID
}

func (r *RequestHeaderWrite) SetType(_type int) {
	r._type = _type
}

func (r *RequestHeaderWrite) SetBodyLen(len int) {
	r.bodyLen = len
}

func (r *RequestHeaderWrite) ID() int64 {
	return r.id
}

func (r *RequestHeaderWrite) SerialID() byte {
	return r.serialID
}

func (r *RequestHeaderWrite) Type() int {
	return r._type
}

func (r *RequestHeaderWrite) BodyLen() int {
	return r.bodyLen
}

func (r *RequestHeaderWrite) ResponseStatus() byte {
	return r.responseStatus
}
