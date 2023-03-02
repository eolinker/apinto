package http_context

import (
	"context"
	"fmt"
	"github.com/valyala/fasthttp"
	"net"
	"time"

	"github.com/eolinker/eosc/utils/config"

	fasthttp_client "github.com/eolinker/apinto/node/fasthttp-client"

	eoscContext "github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var _ http_service.IHttpContext = (*cloneContext)(nil)

// HttpContext fasthttpRequestCtx
type cloneContext struct {
	org           *HttpContext
	proxyRequest  ProxyRequest
	proxyRequests []http_service.IProxy

	ctx                 context.Context
	completeHandler     eoscContext.CompleteHandler
	finishHandler       eoscContext.FinishHandler
	app                 eoscContext.EoApp
	balance             eoscContext.BalanceHandler
	upstreamHostHandler eoscContext.UpstreamHostHandler
	labels              map[string]string

	responseError error
}

func (ctx *cloneContext) GetUpstreamHostHandler() eoscContext.UpstreamHostHandler {
	return ctx.upstreamHostHandler
}

func (ctx *cloneContext) SetUpstreamHostHandler(handler eoscContext.UpstreamHostHandler) {
	ctx.upstreamHostHandler = handler
}

func (ctx *cloneContext) LocalIP() net.IP {
	return ctx.org.LocalIP()
}

func (ctx *cloneContext) LocalAddr() net.Addr {
	return ctx.org.LocalAddr()
}

func (ctx *cloneContext) LocalPort() int {
	return ctx.org.LocalPort()
}

func (ctx *cloneContext) GetApp() eoscContext.EoApp {
	return ctx.app
}

func (ctx *cloneContext) SetApp(app eoscContext.EoApp) {
	ctx.app = app
}

func (ctx *cloneContext) GetBalance() eoscContext.BalanceHandler {
	return ctx.balance
}

func (ctx *cloneContext) SetBalance(handler eoscContext.BalanceHandler) {
	ctx.balance = handler
}

func (ctx *cloneContext) SetLabel(name, value string) {
	ctx.labels[name] = value
}

func (ctx *cloneContext) GetLabel(name string) string {
	return ctx.labels[name]
}

func (ctx *cloneContext) Labels() map[string]string {
	return ctx.labels
}

func (ctx *cloneContext) GetComplete() eoscContext.CompleteHandler {
	return ctx.completeHandler
}

func (ctx *cloneContext) SetCompleteHandler(handler eoscContext.CompleteHandler) {
	ctx.completeHandler = handler
}

func (ctx *cloneContext) GetFinish() eoscContext.FinishHandler {
	return ctx.finishHandler
}

func (ctx *cloneContext) SetFinish(handler eoscContext.FinishHandler) {
	ctx.finishHandler = handler
}

func (ctx *cloneContext) Scheme() string {
	return ctx.org.Scheme()
}

func (ctx *cloneContext) Assert(i interface{}) error {
	if v, ok := i.(*http_service.IHttpContext); ok {
		*v = ctx
		return nil
	}
	return fmt.Errorf("not suport:%s", config.TypeNameOf(i))
}

func (ctx *cloneContext) Proxies() []http_service.IProxy {
	return ctx.proxyRequests
}

func (ctx *cloneContext) Response() http_service.IResponse {
	return nil
}

func (ctx *cloneContext) SendTo(address string, timeout time.Duration) error {

	scheme, host := readAddress(address)
	request := ctx.proxyRequest.Request()

	passHost, targetHost := ctx.GetUpstreamHostHandler().PassHost()
	switch passHost {
	case eoscContext.PassHost:
	case eoscContext.NodeHost:
		request.URI().SetHost(host)
	case eoscContext.ReWriteHost:
		request.URI().SetHost(targetHost)
	}
	response := fasthttp.AcquireResponse()
	beginTime := time.Now()
	ctx.responseError = fasthttp_client.ProxyTimeout(address, request, response, timeout)
	agent := newRequestAgent(&ctx.proxyRequest, host, scheme, beginTime, time.Now())
	if ctx.responseError != nil {
		agent.setStatusCode(504)
	} else {
		agent.setStatusCode(response.StatusCode())
	}

	agent.setResponseLength(response.Header.ContentLength())

	ctx.proxyRequests = append(ctx.proxyRequests, agent)
	return ctx.responseError

}

func (ctx *cloneContext) Context() context.Context {

	return ctx.ctx
}

func (ctx *cloneContext) AcceptTime() time.Time {
	return ctx.org.AcceptTime()
}

func (ctx *cloneContext) Value(key interface{}) interface{} {
	return ctx.org.Value(key)
}

func (ctx *cloneContext) WithValue(key, val interface{}) {
	ctx.ctx = context.WithValue(ctx.Context(), key, val)
}

func (ctx *cloneContext) Proxy() http_service.IRequest {
	return &ctx.proxyRequest
}

func (ctx *cloneContext) Request() http_service.IRequestReader {

	return ctx.org.Request()
}

func (ctx *cloneContext) IsCloneable() bool {
	return false
}

func (ctx *cloneContext) Clone() (eoscContext.EoContext, error) {
	return nil, fmt.Errorf("%s %w", "HttpContext", eoscContext.ErrEoCtxUnCloneable)
}

var copyKey = struct{}{}

// RequestId 请求ID
func (ctx *cloneContext) RequestId() string {
	return ctx.org.requestID
}

// Finish finish
func (ctx *cloneContext) FastFinish() {

	ctx.ctx = nil
	ctx.app = nil
	ctx.balance = nil
	ctx.upstreamHostHandler = nil
	ctx.finishHandler = nil
	ctx.completeHandler = nil

	ctx.proxyRequest.Finish()

}
