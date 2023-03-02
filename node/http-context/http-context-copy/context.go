package http_context_copy

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/eolinker/eosc/utils/config"

	fasthttp_client "github.com/eolinker/apinto/node/fasthttp-client"

	eoscContext "github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/valyala/fasthttp"
)

var _ http_service.IHttpContext = (*HttpContextCopy)(nil)

// HttpContextCopy fasthttpRequestCtx
type HttpContextCopy struct {
	proxyRequest        ProxyRequest
	proxyRequests       []http_service.IProxy
	requestID           string
	response            Response
	requestReader       RequestReader
	ctx                 context.Context
	completeHandler     eoscContext.CompleteHandler
	finishHandler       eoscContext.FinishHandler
	app                 eoscContext.EoApp
	balance             eoscContext.BalanceHandler
	upstreamHostHandler eoscContext.UpstreamHostHandler
	labels              map[string]string
	port                int

	localIP    net.IP
	netAddr    net.Addr
	acceptTime time.Time
}

func (ctx *HttpContextCopy) GetUpstreamHostHandler() eoscContext.UpstreamHostHandler {
	return ctx.upstreamHostHandler
}

func (ctx *HttpContextCopy) SetUpstreamHostHandler(handler eoscContext.UpstreamHostHandler) {
	ctx.upstreamHostHandler = handler
}

func (ctx *HttpContextCopy) LocalIP() net.IP {
	return ctx.localIP
}

func (ctx *HttpContextCopy) LocalAddr() net.Addr {
	return ctx.netAddr
}

func (ctx *HttpContextCopy) LocalPort() int {
	return ctx.port
}

func (ctx *HttpContextCopy) GetApp() eoscContext.EoApp {
	return ctx.app
}

func (ctx *HttpContextCopy) SetApp(app eoscContext.EoApp) {
	ctx.app = app
}

func (ctx *HttpContextCopy) GetBalance() eoscContext.BalanceHandler {
	return ctx.balance
}

func (ctx *HttpContextCopy) SetBalance(handler eoscContext.BalanceHandler) {
	ctx.balance = handler
}

func (ctx *HttpContextCopy) SetLabel(name, value string) {
	ctx.labels[name] = value
}

func (ctx *HttpContextCopy) GetLabel(name string) string {
	return ctx.labels[name]
}

func (ctx *HttpContextCopy) Labels() map[string]string {
	return ctx.labels
}

func (ctx *HttpContextCopy) GetComplete() eoscContext.CompleteHandler {
	return ctx.completeHandler
}

func (ctx *HttpContextCopy) SetCompleteHandler(handler eoscContext.CompleteHandler) {
	ctx.completeHandler = handler
}

func (ctx *HttpContextCopy) GetFinish() eoscContext.FinishHandler {
	return ctx.finishHandler
}

func (ctx *HttpContextCopy) SetFinish(handler eoscContext.FinishHandler) {
	ctx.finishHandler = handler
}

func (ctx *HttpContextCopy) Scheme() string {
	return string(ctx.requestReader.req.URI().Scheme())
}

func (ctx *HttpContextCopy) Assert(i interface{}) error {
	if v, ok := i.(*http_service.IHttpContext); ok {
		*v = ctx
		return nil
	}
	return fmt.Errorf("not suport:%s", config.TypeNameOf(i))
}

func (ctx *HttpContextCopy) Proxies() []http_service.IProxy {
	return ctx.proxyRequests
}

func (ctx *HttpContextCopy) Response() http_service.IResponse {
	return &ctx.response
}

func (ctx *HttpContextCopy) SendTo(address string, timeout time.Duration) error {

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

	beginTime := time.Now()
	ctx.response.responseError = fasthttp_client.ProxyTimeout(address, request, ctx.response.Response, timeout)
	agent := newRequestAgent(&ctx.proxyRequest, host, scheme, beginTime, time.Now())
	if ctx.response.responseError != nil {
		agent.setStatusCode(504)
	} else {
		agent.setStatusCode(ctx.response.Response.StatusCode())
	}

	agent.setResponseLength(ctx.response.Response.Header.ContentLength())

	ctx.proxyRequests = append(ctx.proxyRequests, agent)
	return ctx.response.responseError

}

func (ctx *HttpContextCopy) Context() context.Context {
	if ctx.ctx == nil {

		ctx.ctx = context.Background()
	}
	return ctx.ctx
}

func (ctx *HttpContextCopy) AcceptTime() time.Time {
	return ctx.acceptTime
}

func (ctx *HttpContextCopy) Value(key interface{}) interface{} {
	return ctx.Context().Value(key)
}

func (ctx *HttpContextCopy) WithValue(key, val interface{}) {
	ctx.ctx = context.WithValue(ctx.Context(), key, val)
}

func (ctx *HttpContextCopy) Proxy() http_service.IRequest {
	return &ctx.proxyRequest
}

func (ctx *HttpContextCopy) Request() http_service.IRequestReader {

	return &ctx.requestReader
}

func (ctx *HttpContextCopy) IsCloneable() bool {
	return false
}

func (ctx *HttpContextCopy) Clone() (eoscContext.EoContext, error) {
	return nil, fmt.Errorf("%s %w", "HttpContextCopy", eoscContext.ErrEoCtxUnCloneable)
}

// NewContextCopy 创建Context-Copy
func NewContextCopy(requestCtx *fasthttp.RequestCtx, requestID string, port int, labels map[string]string) *HttpContextCopy {
	ctxCopy := pool.Get().(*HttpContextCopy)
	remoteAddr := requestCtx.RemoteAddr().String()

	cloneReq := fasthttp.AcquireRequest()
	requestCtx.Request.CopyTo(cloneReq)
	cloneResp := fasthttp.AcquireResponse()
	requestCtx.Response.CopyTo(cloneResp)

	ctxCopy.requestReader.reset(cloneReq, remoteAddr)
	ctxCopy.proxyRequest.reset(cloneReq, remoteAddr)
	ctxCopy.proxyRequests = ctxCopy.proxyRequests[:0]
	ctxCopy.response.reset(cloneResp)

	ctxCopy.localIP = requestCtx.LocalIP()
	ctxCopy.netAddr = requestCtx.LocalAddr()
	ctxCopy.acceptTime = requestCtx.Time()

	ctxCopy.requestID = requestID
	ctxCopy.port = port

	ctxCopy.ctx = context.Background()
	ctxCopy.WithValue("request_time", ctxCopy.acceptTime)

	cLabels := make(map[string]string, len(labels))
	for k, v := range labels {
		cLabels[k] = v
	}
	ctxCopy.labels = cLabels

	return ctxCopy

}

// RequestId 请求ID
func (ctx *HttpContextCopy) RequestId() string {
	return ctx.requestID
}

// FastFinish finish
func (ctx *HttpContextCopy) FastFinish() {
	if ctx.response.responseError != nil {
		ctx.response.Response.SetStatusCode(504)
		ctx.response.Response.SetBodyString(ctx.response.responseError.Error())
		return
	}

	ctx.port = 0
	ctx.ctx = nil
	ctx.app = nil
	ctx.balance = nil
	ctx.upstreamHostHandler = nil
	ctx.finishHandler = nil
	ctx.completeHandler = nil

	ctx.requestReader.Finish()
	ctx.proxyRequest.Finish()
	ctx.response.Finish()
	pool.Put(ctx)
	return
}

func NotFound(ctx *HttpContextCopy) {
	ctx.response.Response.SetStatusCode(404)
	ctx.response.Response.SetBody([]byte("404 Not Found"))
}

func readAddress(addr string) (scheme, host string) {
	if i := strings.Index(addr, "://"); i > 0 {
		return strings.ToLower(addr[:i]), addr[i+3:]
	}
	return "http", addr
}
