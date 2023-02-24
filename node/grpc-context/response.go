package grpc_context

import (
	grpc_context "github.com/eolinker/eosc/eocontext/grpc-context"
	"github.com/jhump/protoreflect/dynamic"
	"google.golang.org/grpc/metadata"
)

var _ grpc_context.IResponse = (*Response)(nil)

type Response struct {
	header  metadata.MD
	trailer metadata.MD
	msg     *dynamic.Message
	err     error
}

func (r *Response) SetErr(err error) {
	r.err = err
}

func (r *Response) Error() error {
	return r.err
}

func (r *Response) Write(msg *dynamic.Message) {
	r.msg = msg
}

func NewResponse() *Response {
	return &Response{
		header:  metadata.New(map[string]string{}),
		trailer: metadata.New(map[string]string{}),
		msg:     nil,
	}
}

func (r *Response) Headers() metadata.MD {
	return r.header
}

func (r *Response) Message() *dynamic.Message {
	return r.msg
}

func (r *Response) Trailer() metadata.MD {
	return r.trailer
}
