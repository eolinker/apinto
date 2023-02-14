package grpc_context

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

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
	response                  grpc_context.IResponse
	completeHandler           eocontext.CompleteHandler
	finishHandler             eocontext.FinishHandler
	app                       eocontext.EoApp
	balance                   eocontext.BalanceHandler
	upstreamHostHandler       eocontext.UpstreamHostHandler
	labels                    map[string]string
	tls                       bool
	insecureCertificateVerify bool
	port                      int
	finish                    bool
	errChan                   chan error
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
		ctx:          ctx,
		cancel:       cancel,
		addr:         addr,
		srv:          srv,
		serverStream: stream,
		request:      NewRequest(stream),
		proxy:        NewRequest(stream),
		response:     NewResponse(),
		labels:       map[string]string{},
		errChan:      make(chan error),
	}
}

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
	c.response = response
}

func (c *Context) Invoke(address string, timeout time.Duration) error {
	outgoingCtx, clientConn, err := c.dial(address, timeout)
	if err != nil {
		return err
	}
	// We require that the director's returned context inherits from the serverStream.Context().

	clientCtx, _ := context.WithCancel(outgoingCtx)
	clientStream, err := grpc.NewClientStream(clientCtx, clientStreamDescForProxying, clientConn, c.proxy.FullMethodName())
	if err != nil {
		return err
	}
	c.finish = true
	go c.readError(c.serverStream, clientStream, c.response)
	return nil
}

func (c *Context) FastFinish() error {
	if c.finish {
		err, ok := <-c.errChan
		if !ok {
			return nil
		}
		return err
	}
	c.serverStream.SendHeader(c.response.Headers())
	c.serverStream.SendMsg(c.response.Message())
	c.serverStream.SetTrailer(c.response.Trailer())

	c.port = 0
	c.ctx = nil
	c.app = nil
	c.balance = nil
	c.upstreamHostHandler = nil
	c.finishHandler = nil
	c.completeHandler = nil

	pool.Put(c)
	return nil
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
	//authorities := c.proxy.Headers().Get(":authority")
	//if len(authorities) > 0 {
	//	opts = append(opts, grpc.WithAuthority(authorities[0]))
	//}
	//opts = append(opts, grpc.WithTimeout(timeout))
	conn, err := grpc.DialContext(c.ctx, address, opts...)
	return metadata.NewOutgoingContext(c.Context(), c.proxy.Headers().Copy()), conn, err
}
