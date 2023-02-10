package grpc_context

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/protobuf/types/known/anypb"

	"google.golang.org/grpc/metadata"

	"google.golang.org/grpc/credentials"

	"google.golang.org/grpc/credentials/insecure"

	"google.golang.org/grpc/peer"

	"github.com/eolinker/eosc/utils/config"

	"google.golang.org/grpc"

	"github.com/eolinker/eosc/eocontext"

	grpc_context "github.com/eolinker/eosc/eocontext/grpc-context"
)

var _ grpc_context.IGrpcContext = (*Context)(nil)

type Context struct {
	ctx                       context.Context
	cancel                    context.CancelFunc
	serverStream              grpc.ServerStream
	addr                      net.Addr
	srv                       interface{}
	acceptTime                time.Time
	requestId                 string
	request                   *Request
	proxy                     *Request
	response                  *Response
	completeHandler           eocontext.CompleteHandler
	finishHandler             eocontext.FinishHandler
	app                       eocontext.EoApp
	balance                   eocontext.BalanceHandler
	upstreamHostHandler       eocontext.UpstreamHostHandler
	labels                    map[string]string
	tls                       bool
	insecureCertificateVerify bool
	port                      int
}

func (c *Context) EnableTls(b bool) {
	c.tls = b
}

func (c *Context) InsecureCertificateVerify(b bool) {
	c.insecureCertificateVerify = b
}

func NewContext(srv interface{}, stream grpc.ServerStream) *Context {
	ctx, cancel := context.WithCancel(stream.Context())
	var addr net.Addr = zeroTCPAddr
	p, has := peer.FromContext(ctx)
	if has {
		addr = p.Addr
	}
	return &Context{
		ctx:     ctx,
		cancel:  cancel,
		addr:    addr,
		srv:     srv,
		request: NewRequest(stream),
	}
}

//func (c *Context) ServerStream() grpc.ServerStream {
//	return c.serverStream
//}

func (c *Context) RequestId() string {
	return c.requestId
}

func (c *Context) AcceptTime() time.Time {
	return c.acceptTime
}

func (c *Context) Context() context.Context {
	return c.ctx
}

func (c *Context) Value(key interface{}) interface{} {
	return c.ctx.Value(key)
}

func (c *Context) WithValue(key, val interface{}) {
	c.ctx = context.WithValue(c.ctx, key, val)
}

func (c *Context) Scheme() string {
	return "grpc"
}

func (c *Context) Assert(i interface{}) error {
	if v, ok := i.(*grpc_context.IGrpcContext); ok {
		*v = c
		return nil
	}
	return fmt.Errorf("not suport:%s", config.TypeNameOf(i))
}

func (c *Context) SetLabel(name, value string) {
	c.labels[name] = value
}

func (c *Context) GetLabel(name string) string {
	return c.labels[name]
}

func (c *Context) Labels() map[string]string {
	return c.labels
}

func (c *Context) GetComplete() eocontext.CompleteHandler {
	return c.completeHandler
}

func (c *Context) SetCompleteHandler(handler eocontext.CompleteHandler) {
	c.completeHandler = handler
}

func (c *Context) GetFinish() eocontext.FinishHandler {
	return c.finishHandler
}

func (c *Context) SetFinish(handler eocontext.FinishHandler) {
	c.finishHandler = handler
}

func (c *Context) GetApp() eocontext.EoApp {
	return c.app
}

func (c *Context) SetApp(app eocontext.EoApp) {
	c.app = app
}

func (c *Context) GetBalance() eocontext.BalanceHandler {
	return c.balance
}

func (c *Context) SetBalance(handler eocontext.BalanceHandler) {
	c.balance = handler
}

func (c *Context) GetUpstreamHostHandler() eocontext.UpstreamHostHandler {
	return c.upstreamHostHandler
}

func (c *Context) SetUpstreamHostHandler(handler eocontext.UpstreamHostHandler) {
	c.upstreamHostHandler = handler
}

func (c *Context) LocalIP() net.IP {
	return addrToIP(c.addr)
}

func (c *Context) LocalAddr() net.Addr {
	return c.addr
}

func (c *Context) LocalPort() int {
	return c.port
}

func (c *Context) Request() grpc_context.IRequest {
	return c.request
}

func (c *Context) Proxy() grpc_context.IRequest {
	return c.proxy
}

func (c *Context) Response() grpc_context.IResponse {
	return c.response
}

func (c *Context) SetResponse(response grpc_context.IResponse) {

}

func (c *Context) Invoke(address string, timeout time.Duration) error {
	outgoingCtx, clientConn, err := c.dial(address, timeout)
	if err != nil {
		return err
	}
	return handlerStream(outgoingCtx, c.serverStream, clientConn, c.proxy.FullMethodName())
}

func (c *Context) dial(address string, timeout time.Duration) (context.Context, *grpc.ClientConn, error) {
	opts := make([]grpc.DialOption, 0, 3)
	if !c.tls {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(
			credentials.NewTLS(
				&tls.Config{
					InsecureSkipVerify: c.insecureCertificateVerify,
				},
			)))
	}
	authorities := c.proxy.Headers().Get(":authority")
	if len(authorities) > 0 {
		opts = append(opts, grpc.WithAuthority(authorities[0]))
	}
	opts = append(opts, grpc.WithTimeout(timeout))

	conn, err := grpc.DialContext(c.ctx, address, opts...)
	return metadata.NewOutgoingContext(c.Context(), c.proxy.Headers().Copy()), conn, err
}

func handlerStream(outgoingCtx context.Context, serverStream grpc.ServerStream, clientConn *grpc.ClientConn, fullMethodName string) error {
	// We require that the director's returned context inherits from the serverStream.Context().

	clientCtx, clientCancel := context.WithCancel(outgoingCtx)
	defer clientCancel()
	// TODO(mwitkow): Add a `forwarded` header to metadata, https://en.wikipedia.org/wiki/X-Forwarded-For.
	clientStream, err := grpc.NewClientStream(clientCtx, clientStreamDescForProxying, clientConn, fullMethodName)
	if err != nil {
		return err
	}
	// Explicitly *do not close* s2cErrChan and c2sErrChan, otherwise the select below will not terminate.
	// Channels do not have to be closed, it is just a control flow mechanism, see
	// https://groups.google.com/forum/#!msg/golang-nuts/pZwdYRGxCIk/qpbHxRRPJdUJ
	s2cErrChan := forwardServerToClient(serverStream, clientStream)
	c2sErrChan := forwardClientToServer(clientStream, serverStream)
	// We don't know which side is going to stop sending first, so we need a select between the two.
	for i := 0; i < 2; i++ {
		select {
		case s2cErr := <-s2cErrChan:
			if s2cErr == io.EOF {
				// this is the happy case where the sender has encountered io.EOF, and won't be sending anymore./
				// the clientStream>serverStream may continue pumping though.
				clientStream.CloseSend()
			} else {
				// however, we may have gotten a receive error (stream disconnected, a read error etc) in which case we need
				// to cancel the clientStream to the backend, let all of its goroutines be freed up by the CancelFunc and
				// exit with an error to the stack
				clientCancel()
				return status.Errorf(codes.Internal, "failed proxying s2c: %v", s2cErr)
			}
		case c2sErr := <-c2sErrChan:
			// This happens when the clientStream has nothing else to offer (io.EOF), returned a gRPC error. In those two
			// cases we may have received Trailers as part of the call. In case of other errors (stream closed) the trailers
			// will be nil.
			serverStream.SetTrailer(clientStream.Trailer())
			// c2sErr will contain RPC error from client code. If not io.EOF return the RPC error as server stream error.
			if c2sErr != io.EOF {
				return c2sErr
			}
			return nil
		}
	}
	return status.Errorf(codes.Internal, "gRPC proxying should never reach this stage.")
}

func forwardClientToServer(src grpc.ClientStream, dst grpc.ServerStream) chan error {
	ret := make(chan error, 1)
	go func() {
		f := &anypb.Any{}
		for i := 0; ; i++ {
			if err := src.RecvMsg(f); err != nil {
				ret <- err // this can be io.EOF which is happy case
				break
			}
			if i == 0 {
				// This is a bit of a hack, but client to server headers are only readable after first client msg is
				// received but must be written to server stream before the first msg is flushed.
				// This is the only place to do it nicely.
				md, err := src.Header()
				if err != nil {
					ret <- err
					break
				}
				if err := dst.SendHeader(md); err != nil {
					ret <- err
					break
				}
			}
			if err := dst.SendMsg(f); err != nil {
				ret <- err
				break
			}
		}
	}()
	return ret
}

func forwardServerToClient(src grpc.ServerStream, dst grpc.ClientStream) chan error {
	ret := make(chan error, 1)
	go func() {
		f := &anypb.Any{}
		for i := 0; ; i++ {
			if err := src.RecvMsg(f); err != nil {
				ret <- err // this can be io.EOF which is happy case
				break
			}
			if err := dst.SendMsg(f); err != nil {
				ret <- err
				break
			}
		}
	}()
	return ret
}
