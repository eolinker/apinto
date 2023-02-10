package grpc_context

import (
	"context"
	"io"
	"time"

	grpc_context "github.com/eolinker/eosc/eocontext/grpc-context"
	"github.com/eolinker/eosc/log"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var _ grpc_context.IResponse = (*Response)(nil)

type Response struct {
	stream  grpc.ClientStream
	message *dynamic.Message
	headers metadata.MD
	trailer metadata.MD
}

func NewResponse(stream grpc.ClientStream) *Response {
	headers, err := stream.Header()
	if err != nil {
		log.Error("get grpc response header error: ", err)
		headers = metadata.New(map[string]string{})
	}

	return &Response{
		stream:  stream,
		headers: headers,
		trailer: stream.Trailer(),
	}
}

func (r *Response) Headers() metadata.MD {
	return r.headers
}

func (r *Response) Message(msgDesc *desc.MessageDescriptor) *dynamic.Message {
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
				if err != nil {
					errChan <- err
				}
			}
		}()
		for {
			select {
			case <-ctx.Done():
				// 超时关闭通道
				r.stream.CloseSend()
				break
			case err := <-errChan:
				if err == io.EOF {
					log.Debug("read message eof.")
				}
				break
			}
		}
	}
	r.message = msg

	return msg
}

func (r *Response) Trailer() metadata.MD {
	return r.trailer
}

func (r *Response) ClientStream() grpc.ClientStream {
	return r.stream
}
