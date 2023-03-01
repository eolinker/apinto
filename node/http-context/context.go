package http_context

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
	uuid "github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

var _ http_service.IHttpContext = (*HttpContext)(nil)

// HttpContext fasthttpRequestCtx
type HttpContext struct {
	fastHttpRequestCtx  *fasthttp.RequestCtx
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
}

func (ctx *HttpContext) GetUpstreamHostHandler() eoscContext.UpstreamHostHandler {
	return ctx.upstreamHostHandler
}

func (ctx *HttpContext) SetUpstreamHostHandler(handler eoscContext.UpstreamHostHandler) {
	ctx.upstreamHostHandler = handler
}

func (ctx *HttpContext) LocalIP() net.IP {
	return ctx.fastHttpRequestCtx.LocalIP()
}

func (ctx *HttpContext) LocalAddr() net.Addr {
	return ctx.fastHttpRequestCtx.LocalAddr()
}

func (ctx *HttpContext) LocalPort() int {
	return ctx.port
}

func (ctx *HttpContext) GetApp() eoscContext.EoApp {
	return ctx.app
}

func (ctx *HttpContext) SetApp(app eoscContext.EoApp) {
	ctx.app = app
}

func (ctx *HttpContext) GetBalance() eoscContext.BalanceHandler {
	return ctx.balance
}

func (ctx *HttpContext) SetBalance(handler eoscContext.BalanceHandler) {
	ctx.balance = handler
}

func (ctx *HttpContext) SetLabel(name, value string) {
	ctx.labels[name] = value
}

func (ctx *HttpContext) GetLabel(name string) string {
	return ctx.labels[name]
}

func (ctx *HttpContext) Labels() map[string]string {
	return ctx.labels
}

func (ctx *HttpContext) GetComplete() eoscContext.CompleteHandler {
	return ctx.completeHandler
}

func (ctx *HttpContext) SetCompleteHandler(handler eoscContext.CompleteHandler) {
	ctx.completeHandler = handler
}

func (ctx *HttpContext) GetFinish() eoscContext.FinishHandler {
	return ctx.finishHandler
}

func (ctx *HttpContext) SetFinish(handler eoscContext.FinishHandler) {
	ctx.finishHandler = handler
}

func (ctx *HttpContext) Scheme() string {
	return string(ctx.fastHttpRequestCtx.Request.URI().Scheme())
}

func (ctx *HttpContext) Assert(i interface{}) error {
	if v, ok := i.(*http_service.IHttpContext); ok {
		*v = ctx
		return nil
	}
	return fmt.Errorf("not suport:%s", config.TypeNameOf(i))
}

func (ctx *HttpContext) Proxies() []http_service.IProxy {
	return ctx.proxyRequests
}

func (ctx *HttpContext) Response() http_service.IResponse {
	return &ctx.response
}

func (ctx *HttpContext) SendTo(address string, timeout time.Duration) error {

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
	ctx.response.responseError = fasthttp_client.ProxyTimeout(address, request, &ctx.fastHttpRequestCtx.Response, timeout)
	agent := newRequestAgent(&ctx.proxyRequest, host, scheme, beginTime, time.Now())
	if ctx.response.responseError != nil {
		agent.setStatusCode(504)
	} else {
		agent.setStatusCode(ctx.fastHttpRequestCtx.Response.StatusCode())
	}

	agent.setResponseLength(ctx.fastHttpRequestCtx.Response.Header.ContentLength())

	ctx.proxyRequests = append(ctx.proxyRequests, agent)
	return ctx.response.responseError

}

func (ctx *HttpContext) Context() context.Context {
	if ctx.ctx == nil {
		ctx.ctx = context.Background()
	}
	return ctx.ctx
}

func (ctx *HttpContext) AcceptTime() time.Time {
	return ctx.fastHttpRequestCtx.Time()
}

func (ctx *HttpContext) Value(key interface{}) interface{} {
	return ctx.Context().Value(key)
}

func (ctx *HttpContext) WithValue(key, val interface{}) {
	ctx.ctx = context.WithValue(ctx.Context(), key, val)
}

func (ctx *HttpContext) Proxy() http_service.IRequest {
	return &ctx.proxyRequest
}

func (ctx *HttpContext) Request() http_service.IRequestReader {

	return &ctx.requestReader
}

func (ctx *HttpContext) IsCloneable() bool {
	return true
}

func (ctx *HttpContext) Clone() (eoscContext.EoContext, error) {
	if !ctx.IsCloneable() {
		return nil, fmt.Errorf("%s %w", "HttpContext", eoscContext.ErrEoCtxUnCloneable)
	}

	cloneCtx := pool.Get().(*HttpContext)
	remoteAddr := ctx.fastHttpRequestCtx.RemoteAddr().String()

	cloneReq := fasthttp.AcquireRequest()
	ctx.proxyRequest.Request().CopyTo(cloneReq)
	cloneResp := fasthttp.AcquireResponse()
	ctx.response.Response.CopyTo(cloneResp)

	cloneCtx.fastHttpRequestCtx = ctx.fastHttpRequestCtx //TODO
	cloneCtx.requestID = ctx.requestID                   //TODO
	cloneCtx.requestReader.reset(cloneReq, remoteAddr)
	cloneCtx.proxyRequest.reset(cloneReq, remoteAddr)
	cloneCtx.proxyRequests = cloneCtx.proxyRequests[:0]
	cloneCtx.response.reset(cloneResp)

	cloneCtx.port = ctx.port
	cloneCtx.ctx = ctx.ctx

	cloneCtx.completeHandler = ctx.completeHandler
	cloneCtx.finishHandler = ctx.finishHandler
	cloneCtx.upstreamHostHandler = ctx.upstreamHostHandler
	cloneCtx.app = ctx.app
	cloneCtx.balance = ctx.balance

	cLabels := make(map[string]string, len(ctx.labels))
	for k, v := range ctx.labels {
		cLabels[k] = v
	}
	cloneCtx.labels = cLabels

	return cloneCtx, nil
}

// NewContext 创建Context
func NewContext(ctx *fasthttp.RequestCtx, port int) *HttpContext {

	remoteAddr := ctx.RemoteAddr().String()

	httpContext := pool.Get().(*HttpContext)

	httpContext.fastHttpRequestCtx = ctx
	httpContext.requestID = uuid.New().String()

	httpContext.requestReader.reset(&ctx.Request, remoteAddr)
	httpContext.proxyRequest.reset(&ctx.Request, remoteAddr)
	httpContext.proxyRequests = httpContext.proxyRequests[:0]
	httpContext.response.reset(&ctx.Response)
	httpContext.labels = make(map[string]string)
	httpContext.port = port
	//记录请求时间
	httpContext.ctx = context.Background()
	httpContext.WithValue("request_time", ctx.Time())

	return httpContext

}

// RequestId 请求ID
func (ctx *HttpContext) RequestId() string {
	return ctx.requestID
}

// Finish finish
func (ctx *HttpContext) FastFinish() {
	if ctx.response.responseError != nil {
		ctx.fastHttpRequestCtx.SetStatusCode(504)
		ctx.fastHttpRequestCtx.SetBodyString(ctx.response.responseError.Error())
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
	ctx.fastHttpRequestCtx = nil
	pool.Put(ctx)
	return
}

func NotFound(ctx *HttpContext) {
	ctx.fastHttpRequestCtx.SetStatusCode(404)
	ctx.fastHttpRequestCtx.SetBody([]byte("404 Not Found"))
}

func readAddress(addr string) (scheme, host string) {
	if i := strings.Index(addr, "://"); i > 0 {
		return strings.ToLower(addr[:i]), addr[i+3:]
	}
	return "http", addr
}
