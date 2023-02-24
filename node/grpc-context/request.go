package grpc_context

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/jhump/protoreflect/dynamic"

	"github.com/jhump/protoreflect/desc"

	"github.com/eolinker/eosc/log"

	grpc_context "github.com/eolinker/eosc/eocontext/grpc-context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var _ grpc_context.IRequest = (*Request)(nil)

type Request struct {
	headers metadata.MD
	host    string
	service string
	method  string
	message *dynamic.Message
	stream  grpc.ServerStream
	realIP  string
}

func (r *Request) SetHost(s string) {
	r.host = s
}

func (r *Request) SetService(service string) {
	r.service = service
}

func (r *Request) SetMethod(method string) {
	r.method = method
}

func NewRequest(stream grpc.ServerStream) *Request {
	fullService, has := grpc.MethodFromServerStream(stream)
	var service, method string
	if has {
		names := strings.Split(strings.TrimPrefix(fullService, "/"), "/")
		service = names[0]
		if len(names) > 1 {
			method = names[1]
		}
	}
	md, has := metadata.FromIncomingContext(stream.Context())
	if !has {
		md = metadata.New(map[string]string{})
	}
	hosts := md.Get(":authority")
	return &Request{
		stream:  stream,
		service: service,
		method:  method,
		headers: md,
		host:    strings.Join(hosts, ";"),
	}
}

func (r *Request) Headers() metadata.MD {
	return r.headers
}

func (r *Request) Host() string {
	return r.host
}

func (r *Request) Service() string {
	return r.service
}

func (r *Request) Method() string {
	return r.method
}

func (r *Request) FullMethodName() string {
	return fmt.Sprintf("/%s/%s", r.service, r.method)
}

func (r *Request) RealIP() string {
	if r.realIP == "" {
		r.realIP = strings.Join(r.headers.Get("x-real-ip"), ";")
	}
	return r.realIP
}

func (r *Request) ForwardIP() string {
	return strings.Join(r.headers.Get("x-forwarded-for"), ";")
}

func (r *Request) Message(msgDesc *desc.MessageDescriptor) *dynamic.Message {
	if r.message != nil {
		return r.message
	}
	msg := dynamic.NewMessage(msgDesc)
	if r.stream != nil {
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		errChan := make(chan error)
		defer close(errChan)
		go func() {
			var err error
			for {
				err = r.stream.RecvMsg(msg)
				r.message = msg
				if err != nil {
					errChan <- err
				}
			}
		}()
		for {
			select {
			case <-ctx.Done():
				r.message = msg
				return r.message
			case err, ok := <-errChan:
				if !ok {
					return r.message
				}
				if err == io.EOF {
					log.Debug("read message eof.")
				}
				return r.message
			}
		}
	}

	return msg
}
