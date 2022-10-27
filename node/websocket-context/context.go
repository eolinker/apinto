package websocket_context

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	http_service "github.com/eolinker/eosc/eocontext/http-context"

	"github.com/fasthttp/websocket"

	websocket_context "github.com/eolinker/eosc/eocontext/websocket-context"

	"github.com/eolinker/eosc/utils/config"

	eoscContext "github.com/eolinker/eosc/eocontext"
	uuid "github.com/satori/go.uuid"
	"github.com/valyala/fasthttp"
)

var _ websocket_context.IWebsocketContext = (*Context)(nil)

//Context fasthttpRequestCtx
type Context struct {
	fastHttpRequestCtx  *fasthttp.RequestCtx
	requestID           string
	ctx                 context.Context
	requestReader       http_service.IRequestReader
	completeHandler     eoscContext.CompleteHandler
	finishHandler       eoscContext.FinishHandler
	app                 eoscContext.EoApp
	balance             eoscContext.BalanceHandler
	upstreamHostHandler eoscContext.UpstreamHostHandler
	labels              map[string]string
	port                int
	response            *Response
	clientConn          *websocket.Conn
	upstreamConn        *websocket.Conn
	finishChan          chan struct{}
}

func (ctx *Context) Response() http_service.IResponse {
	return ctx.response
}

func (ctx *Context) Request() http_service.IRequestReader {

	return ctx.requestReader
}

func (ctx *Context) Dial(address string, timeout time.Duration) error {
	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: timeout,
	}

	conn, _, err := dialer.Dial(address, ctx.requestReader.Header().Headers())
	if err != nil {
		return err
	}
	ctx.upstreamConn = conn
	return nil
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
	if v, ok := i.(*websocket_context.IWebsocketContext); ok {
		*v = ctx
		return nil
	}
	return fmt.Errorf("not suport:%s", config.TypeNameOf(i))
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

var upgrader = websocket.FastHTTPUpgrader{}

//Upgrade Upgrade
func Upgrade(ctx *fasthttp.RequestCtx, port int) (*Context, error) {

	ch := make(chan error)
	finishChan := make(chan struct{})
	var clientConn *websocket.Conn
	go func() {
		err := upgrader.Upgrade(ctx, func(conn *websocket.Conn) {
			clientConn = conn
			close(ch)
			<-finishChan

		})
		if err != nil {
			ch <- err
			close(ch)
		}
	}()
	err, ok := <-ch
	if ok {
		return nil, err
	}
	newCtx := &Context{
		fastHttpRequestCtx: ctx,
		requestReader:      NewRequestReader(&ctx.Request, ctx.RemoteAddr().String()),
		requestID:          uuid.NewV4().String(),
		port:               port,
		response:           NewResponse(ctx),
		clientConn:         clientConn,
		finishChan:         finishChan,
	}
	//记录请求时间
	newCtx.WithValue("request_time", ctx.Time())
	return newCtx, nil
}

//RequestId 请求ID
func (ctx *Context) RequestId() string {
	return ctx.requestID
}

func (ctx *Context) ClientConn() *websocket.Conn {
	return ctx.clientConn
}

func (ctx *Context) UpstreamConn() *websocket.Conn {
	return ctx.upstreamConn
}

func (ctx *Context) WebsocketFinish() {
	close(ctx.finishChan)
	ctx.upstreamConn.Close()
	ctx.clientConn.Close()
}
