package http_context

import (
	"context"
	"fmt"
	"net"
	"time"

	http_entry "github.com/eolinker/apinto/entries/http-entry"

	"github.com/eolinker/eosc"

	"github.com/valyala/fasthttp"

	"github.com/eolinker/eosc/log"
	"github.com/eolinker/eosc/utils/config"

	fasthttp_client "github.com/eolinker/apinto/node/fasthttp-client"

	eoscContext "github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var _ http_service.IHttpContext = (*cloneContext)(nil)

// HttpContext fasthttpRequestCtx
type cloneContext struct {
	org                 *HttpContext
	proxyRequest        ProxyRequest
	response            Response
	proxyRequests       []http_service.IProxy
	ctx                 context.Context
	completeHandler     eoscContext.CompleteHandler
	finishHandler       eoscContext.FinishHandler
	app                 eoscContext.EoApp
	balance             eoscContext.BalanceHandler
	upstreamHostHandler eoscContext.UpstreamHostHandler
	labels              map[string]string
	entry               eosc.IEntry
	responseError       error
}

func (ctx *cloneContext) ProxyClone() http_service.IRequest {
	// 创建一个新的 ProxyRequest 实例
	req := fasthttp.AcquireRequest()
	// 将原始请求内容拷贝到新请求
	ctx.proxyRequest.Request().CopyTo(req)

	// 使用 ProxyRequest.reset 初始化克隆的 ProxyRequest
	cloneProxy := &ProxyRequest{}
	cloneProxy.reset(req, ctx.proxyRequest.remoteAddr)

	return cloneProxy
}

func (ctx *cloneContext) SetProxy(proxy http_service.IRequest) {
	if p, ok := proxy.(*ProxyRequest); ok {
		// 替换当前的 proxyRequest
		ctx.proxyRequest = *p
		//记录到 proxyRequests 数组中
		//ctx.proxyRequests = append(ctx.proxyRequests, p)
	} else {
		log.Warn("SetProxy failed: incompatible type")
	}
}

func (ctx *cloneContext) GetEntry() eosc.IEntry {
	if ctx.entry == nil {
		ctx.entry = http_entry.NewEntry(ctx)
	}
	return ctx.entry
}

func (ctx *cloneContext) RealIP() string {
	return ctx.org.RealIP()
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
	return &ctx.response
}

func (ctx *cloneContext) SendTo(scheme string, node eoscContext.INode, timeout time.Duration) error {

	host := node.Addr()
	request := ctx.proxyRequest.Request()
	rewriteHost := string(request.Host())
	passHost, targetHost := ctx.GetUpstreamHostHandler().PassHost()
	switch passHost {
	case eoscContext.PassHost:
	case eoscContext.NodeHost:
		rewriteHost = host
		request.URI().SetHost(host)
	case eoscContext.ReWriteHost:
		rewriteHost = targetHost
		request.URI().SetHost(targetHost)
	}
	beginTime := time.Now()
	ctx.responseError = fasthttp_client.ProxyTimeout(scheme, rewriteHost, node, request, ctx.response.Response, timeout)
	var responseHeader fasthttp.ResponseHeader
	if ctx.response.Response != nil {
		responseHeader = ctx.response.Response.Header
	}
	agent := newRequestAgent(&ctx.proxyRequest, host, scheme, responseHeader, beginTime, time.Now())
	if ctx.responseError != nil {
		agent.setStatusCode(504)
	} else {
		agent.setStatusCode(ctx.response.Response.StatusCode())
		ip, port := parseAddr(ctx.response.Response.RemoteAddr().String())
		agent.setRemoteIP(ip)
		agent.setRemotePort(port)
		ctx.response.remoteIP = ip
		ctx.response.remotePort = port
	}
	agent.responseBody.Write(ctx.response.Response.Body())

	agent.setResponseLength(ctx.response.Response.Header.ContentLength())

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
	return ctx.ctx.Value(key)
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
	fasthttp.ReleaseRequest(ctx.proxyRequest.req)
	fasthttp.ReleaseResponse(ctx.response.Response)
	ctx.response.Finish()
	ctx.proxyRequest.Finish()

}
