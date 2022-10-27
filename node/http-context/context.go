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
	uuid "github.com/satori/go.uuid"
	"github.com/valyala/fasthttp"
)

var _ http_service.IHttpContext = (*Context)(nil)

//Context fasthttpRequestCtx
type Context struct {
	fastHttpRequestCtx  *fasthttp.RequestCtx
	proxyRequest        *ProxyRequest
	proxyRequests       []http_service.IRequest
	requestID           string
	response            *Response
	requestReader       *RequestReader
	ctx                 context.Context
	completeHandler     eoscContext.CompleteHandler
	finishHandler       eoscContext.FinishHandler
	app                 eoscContext.EoApp
	balance             eoscContext.BalanceHandler
	upstreamHostHandler eoscContext.UpstreamHostHandler
	labels              map[string]string
	port                int
}

func (ctx *Context) GetUpstreamHostHandler() eoscContext.UpstreamHostHandler {
	return ctx.upstreamHostHandler
}

func (ctx *Context) SetUpstreamHostHandler(handler eoscContext.UpstreamHostHandler) {
	ctx.upstreamHostHandler = handler
}

func (ctx *Context) LocalIP() net.IP {
	return ctx.fastHttpRequestCtx.LocalIP()
}

func (ctx *Context) LocalAddr() net.Addr {
	return ctx.fastHttpRequestCtx.LocalAddr()
}

func (ctx *Context) LocalPort() int {
	return ctx.port
}

func (ctx *Context) GetApp() eoscContext.EoApp {
	return ctx.app
}

func (ctx *Context) SetApp(app eoscContext.EoApp) {
	ctx.app = app
}

func (ctx *Context) GetBalance() eoscContext.BalanceHandler {
	return ctx.balance
}

func (ctx *Context) SetBalance(handler eoscContext.BalanceHandler) {
	ctx.balance = handler
}

func (ctx *Context) SetLabel(name, value string) {
	ctx.labels[name] = value
}

func (ctx *Context) GetLabel(name string) string {
	return ctx.labels[name]
}

func (ctx *Context) Labels() map[string]string {
	return ctx.labels
}

func (ctx *Context) GetComplete() eoscContext.CompleteHandler {
	return ctx.completeHandler
}

func (ctx *Context) SetCompleteHandler(handler eoscContext.CompleteHandler) {
	ctx.completeHandler = handler
}

func (ctx *Context) GetFinish() eoscContext.FinishHandler {
	return ctx.finishHandler
}

func (ctx *Context) SetFinish(handler eoscContext.FinishHandler) {
	ctx.finishHandler = handler
}

func (ctx *Context) Scheme() string {
	return string(ctx.fastHttpRequestCtx.Request.URI().Scheme())
}

func (ctx *Context) Assert(i interface{}) error {
	if v, ok := i.(*http_service.IHttpContext); ok {
		*v = ctx
		return nil
	}
	return fmt.Errorf("not suport:%s", config.TypeNameOf(i))
}

func (ctx *Context) Proxies() []http_service.IRequest {
	return ctx.proxyRequests
}

func (ctx *Context) Response() http_service.IResponse {
	return ctx.response
}

func (ctx *Context) SendTo(address string, timeout time.Duration) error {

	_, host := readAddress(address)
	request := ctx.proxyRequest.Request()
	ctx.proxyRequests = append(ctx.proxyRequests, newRequestAgent(ctx.proxyRequest, host))

	passHost, targethost := ctx.GetUpstreamHostHandler().PassHost()
	switch passHost {
	case eoscContext.PassHost:
	case eoscContext.NodeHost:
		request.URI().SetHost(host)
	case eoscContext.ReWriteHost:
		request.URI().SetHost(targethost)
	}

	ctx.response.responseError = fasthttp_client.ProxyTimeout(address, request, &ctx.fastHttpRequestCtx.Response, timeout)

	return ctx.response.responseError

}

func (ctx *Context) Context() context.Context {
	if ctx.ctx == nil {
		ctx.ctx = context.Background()
	}
	return ctx.ctx
}

func (ctx *Context) Value(key interface{}) interface{} {
	return ctx.Context().Value(key)
}

func (ctx *Context) WithValue(key, val interface{}) {
	ctx.ctx = context.WithValue(ctx.Context(), key, val)
}

func (ctx *Context) Proxy() http_service.IRequest {
	return ctx.proxyRequest
}

func (ctx *Context) Request() http_service.IRequestReader {

	return ctx.requestReader
}

//NewContext 创建Context
func NewContext(ctx *fasthttp.RequestCtx, port int) *Context {

	id := uuid.NewV4()
	requestID := id.String()
	newCtx := &Context{
		fastHttpRequestCtx: ctx,
		requestID:          requestID,
		requestReader:      NewRequestReader(&ctx.Request, ctx.RemoteAddr().String()),
		proxyRequest:       NewProxyRequest(&ctx.Request, ctx.RemoteAddr().String()),
		proxyRequests:      make([]http_service.IRequest, 0, 5),
		response:           NewResponse(ctx),
		labels:             make(map[string]string),
		port:               port,
	}
	//记录请求时间
	newCtx.WithValue("request_time", ctx.Time())
	return newCtx
}

//RequestId 请求ID
func (ctx *Context) RequestId() string {
	return ctx.requestID
}

//Finish finish
func (ctx *Context) FastFinish() {
	if ctx.response.responseError != nil {
		ctx.fastHttpRequestCtx.SetStatusCode(504)
		ctx.fastHttpRequestCtx.SetBodyString(ctx.response.responseError.Error())
		return
	}

	ctx.requestReader.Finish()
	ctx.proxyRequest.Finish()

	return
}

func NotFound(ctx *Context) {
	ctx.fastHttpRequestCtx.SetStatusCode(404)
	ctx.fastHttpRequestCtx.SetBody([]byte("404 Not Found"))
}

func readAddress(addr string) (scheme, host string) {
	if i := strings.Index(addr, "://"); i > 0 {
		return strings.ToLower(addr[:i]), addr[i+3:]
	}
	return "http", addr
}
