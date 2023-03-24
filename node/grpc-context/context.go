package grpc_context

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"

	"google.golang.org/grpc/metadata"

	"google.golang.org/grpc/peer"

	"github.com/eolinker/eosc/utils/config"

	"google.golang.org/grpc"

	"github.com/eolinker/eosc/eocontext"

	grpc_context "github.com/eolinker/eosc/eocontext/grpc-context"
)

var _ grpc_context.IGrpcContext = (*Context)(nil)

var (
	pool = sync.Pool{
		New: newContext,
	}
)

func newContext() interface{} {
	h := new(Context)
	return h
}

type Context struct {
	ctx                       context.Context
	cancel                    context.CancelFunc
	serverStream              grpc.ServerStream
	addr                      net.Addr
	srv                       interface{}
	acceptTime                time.Time
	requestId                 string
	request                   grpc_context.IRequest
	proxy                     grpc_context.IRequest
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
	now := time.Now()
	ctx, cancel := context.WithCancel(stream.Context())
	var addr net.Addr = zeroTCPAddr
	p, has := peer.FromContext(ctx)
	if has {
		addr = p.Addr
	}

	newCtx := &Context{
		requestId:    uuid.New().String(),
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
	newCtx.WithValue("request_time", now)
	return newCtx
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

func (c *Context) Invoke(node eocontext.INode, timeout time.Duration) error {

	err := c.doInvoke(node.Addr(), timeout)
	if err != nil {
		node.Down()
		return err
	}
	return nil
}
func (c *Context) doInvoke(address string, timeout time.Duration) error {
	passHost, targetHost := c.GetUpstreamHostHandler().PassHost()
	switch passHost {
	case eocontext.NodeHost:
		c.proxy.SetHost(address)
	case eocontext.ReWriteHost:
		c.proxy.SetHost(targetHost)
	}
	clientConn, err := clientPool.Get(address, c.tls, c.proxy.Host()).Get()
	if err != nil {
		return err
	}

	c.proxy.Headers().Set("grpc-timeout", fmt.Sprintf("%dn", timeout))
	clientCtx, _ := context.WithCancel(metadata.NewOutgoingContext(c.Context(), c.proxy.Headers().Copy()))
	clientStream, err := grpc.NewClientStream(clientCtx, clientStreamDescForProxying, clientConn, c.proxy.FullMethodName())
	if err != nil {
		return err
	}
	c.finish = true
	go c.readError(c.serverStream, clientStream, c.response)
	return nil
}

func (c *Context) FastFinish() error {
	defer c.reset()
	if c.finish {
		err, ok := <-c.errChan
		if !ok {
			return nil
		}
		return err
	}
	err := c.response.Error()
	if err != nil {
		return err
	}
	c.serverStream.SendHeader(c.response.Headers())
	c.serverStream.SendMsg(c.response.Message())
	c.serverStream.SetTrailer(c.response.Trailer())

	return nil
}

func (c *Context) reset() {
	c.port = 0
	c.ctx = nil
	c.app = nil
	c.balance = nil
	c.upstreamHostHandler = nil
	c.finishHandler = nil
	c.completeHandler = nil

	pool.Put(c)
}

func (c *Context) IsCloneable() bool {
	return false
}

func (c *Context) Clone() (eocontext.EoContext, error) {
	//TODO
	return nil, fmt.Errorf("%s %w", "GrpcContext", eocontext.ErrEoCtxUnCloneable)
}
