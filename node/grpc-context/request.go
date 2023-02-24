package grpc_context

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/jhump/protoreflect/dynamic"

	"github.com/jhump/protoreflect/desc"

	grpc_context "github.com/eolinker/eosc/eocontext/grpc-context"
	"github.com/eolinker/eosc/log"
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
	r.message = dynamic.NewMessage(msgDesc)
	if r.stream == nil {
		return r.message
	}

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	errChan := make(chan error)

	go func() {
		var err error
		for {
			err = r.stream.RecvMsg(r.message)
			if err != nil {
				errChan <- err
				close(errChan)
				return
			}
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return r.message
		case err, ok := <-errChan:
			if !ok {
				return r.message
			}
			if err != nil {
				if err == io.EOF {
					log.Debug("read message eof.")
				} else {
					log.Debug("read message error: ", err)
				}
			}
			return r.message
		}
	}
}
