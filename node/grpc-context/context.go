package grpc_context

import (
	"context"
	"net"
	"time"

	"github.com/eolinker/eosc/eocontext"

	grpc_context "github.com/eolinker/eosc/eocontext/grpc-context"
)

var _ grpc_context.IGrpcContext = (*Context)(nil)

type Context struct {
}

func (c *Context) RequestId() string {
	//TODO implement me
	panic("implement me")
}

func (c *Context) AcceptTime() time.Time {
	//TODO implement me
	panic("implement me")
}

func (c *Context) Context() context.Context {
	//TODO implement me
	panic("implement me")
}

func (c *Context) Value(key interface{}) interface{} {
	//TODO implement me
	panic("implement me")
}

func (c *Context) WithValue(key, val interface{}) {
	//TODO implement me
	panic("implement me")
}

func (c *Context) Scheme() string {
	//TODO implement me
	panic("implement me")
}

func (c *Context) Assert(i interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (c *Context) SetLabel(name, value string) {
	//TODO implement me
	panic("implement me")
}

func (c *Context) GetLabel(name string) string {
	//TODO implement me
	panic("implement me")
}

func (c *Context) Labels() map[string]string {
	//TODO implement me
	panic("implement me")
}

func (c *Context) GetComplete() eocontext.CompleteHandler {
	//TODO implement me
	panic("implement me")
}

func (c *Context) SetCompleteHandler(handler eocontext.CompleteHandler) {
	//TODO implement me
	panic("implement me")
}

func (c *Context) GetFinish() eocontext.FinishHandler {
	//TODO implement me
	panic("implement me")
}

func (c *Context) SetFinish(handler eocontext.FinishHandler) {
	//TODO implement me
	panic("implement me")
}

func (c *Context) GetApp() eocontext.EoApp {
	//TODO implement me
	panic("implement me")
}

func (c *Context) SetApp(app eocontext.EoApp) {
	//TODO implement me
	panic("implement me")
}

func (c *Context) GetBalance() eocontext.BalanceHandler {
	//TODO implement me
	panic("implement me")
}

func (c *Context) SetBalance(handler eocontext.BalanceHandler) {
	//TODO implement me
	panic("implement me")
}

func (c *Context) GetUpstreamHostHandler() eocontext.UpstreamHostHandler {
	//TODO implement me
	panic("implement me")
}

func (c *Context) SetUpstreamHostHandler(handler eocontext.UpstreamHostHandler) {
	//TODO implement me
	panic("implement me")
}

func (c *Context) LocalIP() net.IP {
	//TODO implement me
	panic("implement me")
}

func (c *Context) LocalAddr() net.Addr {
	//TODO implement me
	panic("implement me")
}

func (c *Context) LocalPort() int {
	//TODO implement me
	panic("implement me")
}

func (c *Context) Request() grpc_context.IRequest {
	//TODO implement me
	panic("implement me")
}

func (c *Context) Proxy() grpc_context.IRequest {
	//TODO implement me
	panic("implement me")
}

func (c *Context) Response() grpc_context.IResponse {
	//TODO implement me
	panic("implement me")
}
